package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type GamificationRepository interface {
	CreateXPTransaction(ctx context.Context, t *models.XPTransaction) (created bool, err error)
	GetXPTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.XPTransaction, error)
	ListXPTransactionsForUser(ctx context.Context, userID string, limit, offset int) ([]models.XPTransaction, error)
	GetUserStats(ctx context.Context, userID string) (*models.UserStats, error)
	UpsertUserStatsForSession(ctx context.Context, userID string, score *float64, xpDelta int) (*models.UserStats, error)
	AddUserXP(ctx context.Context, userID string, xpDelta int) error
	UpsertDailyActivity(ctx context.Context, userID string, xpEarned int) error
	ListDailyActivityForUser(ctx context.Context, userID string, limit int) ([]models.DailyActivity, error)
	GetUserSkillStats(ctx context.Context, userID string) (*models.UserSkillStats, error)
	ListActiveBadges(ctx context.Context) ([]models.Badge, error)
	ListUserBadges(ctx context.Context, userID string) ([]models.UserBadge, error)
	GetUserBadge(ctx context.Context, userID string, badgeID uuid.UUID) (*models.UserBadge, error)
	AwardBadge(ctx context.Context, b *models.UserBadge) (created bool, err error)
	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgGamificationRepository struct {
	pool *pgxpool.Pool
}

func NewGamificationRepository(pool *pgxpool.Pool) GamificationRepository {
	return &pgGamificationRepository{pool: pool}
}

// CreateXPTransaction is idempotent on idempotency_key. created reports whether
// this call actually inserted a new row (false if the transaction already existed).
func (r *pgGamificationRepository) CreateXPTransaction(ctx context.Context, t *models.XPTransaction) (bool, error) {
	const q = `
		insert into xp_transactions (id, user_id, session_id, amount, reason, idempotency_key, created_at)
		values ($1,$2,$3,$4,$5,$6,now())
		on conflict (idempotency_key) do nothing
		returning created_at`

	err := r.pool.QueryRow(ctx, q, t.ID, t.UserID, t.SessionID, t.Amount, t.Reason, t.IdempotencyKey).Scan(&t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, getErr := r.GetXPTransactionByIdempotencyKey(ctx, t.IdempotencyKey)
			if getErr != nil {
				return false, fmt.Errorf("repository: create xp transaction: %w", getErr)
			}
			*t = *existing
			return false, nil
		}
		return false, fmt.Errorf("repository: create xp transaction: %w", err)
	}

	return true, nil
}

func (r *pgGamificationRepository) GetXPTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.XPTransaction, error) {
	const q = `
		select id, user_id, session_id, amount, reason, idempotency_key, created_at
		from xp_transactions
		where idempotency_key = $1`

	var t models.XPTransaction
	err := r.pool.QueryRow(ctx, q, idempotencyKey).Scan(
		&t.ID, &t.UserID, &t.SessionID, &t.Amount, &t.Reason, &t.IdempotencyKey, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get xp transaction by idempotency key: %w", err)
	}

	return &t, nil
}

func (r *pgGamificationRepository) ListXPTransactionsForUser(ctx context.Context, userID string, limit, offset int) ([]models.XPTransaction, error) {
	const q = `
		select id, user_id, session_id, amount, reason, idempotency_key, created_at
		from xp_transactions
		where user_id = $1
		order by created_at desc
		limit $2 offset $3`

	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list xp transactions: %w", err)
	}
	defer rows.Close()

	txns := make([]models.XPTransaction, 0)
	for rows.Next() {
		var t models.XPTransaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.SessionID, &t.Amount, &t.Reason, &t.IdempotencyKey, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan xp transaction: %w", err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list xp transactions: %w", err)
	}

	return txns, nil
}

func (r *pgGamificationRepository) GetUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
	const q = `
		select user_id, total_sessions, scored_sessions, average_score, best_score, current_streak_days,
		       longest_streak_days, last_session_date, total_xp, updated_at
		from user_stats
		where user_id = $1`

	var s models.UserStats
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&s.UserID, &s.TotalSessions, &s.ScoredSessions, &s.AverageScore, &s.BestScore, &s.CurrentStreakDays,
		&s.LongestStreakDays, &s.LastSessionDate, &s.TotalXP, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user stats: %w", err)
	}

	return &s, nil
}

// UpsertUserStatsForSession creates or updates a user's aggregate stats row for
// a completed session. score may be nil for sessions that skipped AI analysis
// (over quota) - in that case average_score/best_score are left untouched and
// scored_sessions does not advance, but total_sessions, streak, and total_xp
// still do. average_score is weighted by scored_sessions specifically (not
// total_sessions), so unscored sessions never dilute the average. Streak
// logic: if the user's last_session_date is today, the streak is unchanged
// (second session same day); if it was yesterday, the streak extends by one;
// otherwise it resets to 1.
func (r *pgGamificationRepository) UpsertUserStatsForSession(ctx context.Context, userID string, score *float64, xpDelta int) (*models.UserStats, error) {
	const q = `
		insert into user_stats (user_id, total_sessions, scored_sessions, average_score, best_score, current_streak_days, longest_streak_days, last_session_date, total_xp, updated_at)
		values (
			$1, 1,
			case when $2 is null then 0 else 1 end,
			coalesce($2, 0), coalesce($2, 0), 1, 1, current_date, $3, now()
		)
		on conflict (user_id) do update set
			total_sessions = user_stats.total_sessions + 1,
			scored_sessions = user_stats.scored_sessions + case when $2 is null then 0 else 1 end,
			average_score = case
				when $2 is null then user_stats.average_score
				else ((user_stats.average_score * user_stats.scored_sessions) + $2) / (user_stats.scored_sessions + 1)
			end,
			best_score = case
				when $2 is null then user_stats.best_score
				when $2 > user_stats.best_score then $2
				else user_stats.best_score
			end,
			current_streak_days = case
				when user_stats.last_session_date = current_date then user_stats.current_streak_days
				when user_stats.last_session_date = current_date - 1 then user_stats.current_streak_days + 1
				else 1
			end,
			longest_streak_days = greatest(
				user_stats.longest_streak_days,
				case
					when user_stats.last_session_date = current_date then user_stats.current_streak_days
					when user_stats.last_session_date = current_date - 1 then user_stats.current_streak_days + 1
					else 1
				end
			),
			last_session_date = current_date,
			total_xp = user_stats.total_xp + $3,
			updated_at = now()
		returning user_id, total_sessions, scored_sessions, average_score, best_score, current_streak_days,
		          longest_streak_days, last_session_date, total_xp, updated_at`

	var s models.UserStats
	err := r.pool.QueryRow(ctx, q, userID, score, xpDelta).Scan(
		&s.UserID, &s.TotalSessions, &s.ScoredSessions, &s.AverageScore, &s.BestScore, &s.CurrentStreakDays,
		&s.LongestStreakDays, &s.LastSessionDate, &s.TotalXP, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: upsert user stats for session: %w", err)
	}

	return &s, nil
}

// AddUserXP adds xpDelta to a user's total_xp without touching session/streak
// fields. Used for XP awarded outside the session-completion flow (badges).
func (r *pgGamificationRepository) AddUserXP(ctx context.Context, userID string, xpDelta int) error {
	const q = `
		insert into user_stats (user_id, total_sessions, scored_sessions, average_score, best_score, current_streak_days, longest_streak_days, total_xp, updated_at)
		values ($1, 0, 0, 0, 0, 0, 0, $2, now())
		on conflict (user_id) do update set
			total_xp = user_stats.total_xp + $2,
			updated_at = now()`

	if _, err := r.pool.Exec(ctx, q, userID, xpDelta); err != nil {
		return fmt.Errorf("repository: add user xp: %w", err)
	}

	return nil
}

// UpsertDailyActivity increments today's session count and XP total for the
// user, creating today's row if it doesn't exist yet. Requires a unique
// constraint on daily_activity(user_id, activity_date).
func (r *pgGamificationRepository) UpsertDailyActivity(ctx context.Context, userID string, xpEarned int) error {
	const q = `
		insert into daily_activity (id, user_id, activity_date, sessions_count, xp_earned, created_at, updated_at)
		values ($1, $2, current_date, 1, $3, now(), now())
		on conflict (user_id, activity_date) do update set
			sessions_count = daily_activity.sessions_count + 1,
			xp_earned = daily_activity.xp_earned + $3,
			updated_at = now()`

	if _, err := r.pool.Exec(ctx, q, uuid.New(), userID, xpEarned); err != nil {
		return fmt.Errorf("repository: upsert daily activity: %w", err)
	}

	return nil
}

func (r *pgGamificationRepository) ListDailyActivityForUser(ctx context.Context, userID string, limit int) ([]models.DailyActivity, error) {
	const q = `
		select id, user_id, activity_date, sessions_count, xp_earned, created_at, updated_at
		from daily_activity
		where user_id = $1
		order by activity_date desc
		limit $2`

	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repository: list daily activity: %w", err)
	}
	defer rows.Close()

	activities := make([]models.DailyActivity, 0)
	for rows.Next() {
		var a models.DailyActivity
		if err := rows.Scan(&a.ID, &a.UserID, &a.ActivityDate, &a.SessionsCount, &a.XPEarned, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan daily activity: %w", err)
		}
		activities = append(activities, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list daily activity: %w", err)
	}

	return activities, nil
}

func (r *pgGamificationRepository) GetUserSkillStats(ctx context.Context, userID string) (*models.UserSkillStats, error) {
	const q = `
		select user_id, avg_clarity, avg_delivery, avg_content, avg_vocabulary, avg_grammar,
		       sessions_counted, updated_at
		from user_skill_stats
		where user_id = $1`

	var s models.UserSkillStats
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&s.UserID, &s.AvgClarity, &s.AvgDelivery, &s.AvgContent, &s.AvgVocabulary, &s.AvgGrammar,
		&s.SessionsCounted, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user skill stats: %w", err)
	}

	return &s, nil
}

func (r *pgGamificationRepository) ListActiveBadges(ctx context.Context) ([]models.Badge, error) {
	const q = `
		select id, slug, name, description, icon, criteria_type, criteria_value,
		       is_active, sort_order, created_at
		from badges
		where is_active = true
		order by sort_order asc, name asc`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repository: list active badges: %w", err)
	}
	defer rows.Close()

	badges := make([]models.Badge, 0)
	for rows.Next() {
		var b models.Badge
		if err := rows.Scan(
			&b.ID, &b.Slug, &b.Name, &b.Description, &b.Icon, &b.CriteriaType, &b.CriteriaValue,
			&b.IsActive, &b.SortOrder, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan badge: %w", err)
		}
		badges = append(badges, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list active badges: %w", err)
	}

	return badges, nil
}

func (r *pgGamificationRepository) ListUserBadges(ctx context.Context, userID string) ([]models.UserBadge, error) {
	const q = `
		select id, user_id, badge_id, session_id, earned_at
		from user_badges
		where user_id = $1
		order by earned_at desc`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list user badges: %w", err)
	}
	defer rows.Close()

	badges := make([]models.UserBadge, 0)
	for rows.Next() {
		var b models.UserBadge
		if err := rows.Scan(&b.ID, &b.UserID, &b.BadgeID, &b.SessionID, &b.EarnedAt); err != nil {
			return nil, fmt.Errorf("repository: scan user badge: %w", err)
		}
		badges = append(badges, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list user badges: %w", err)
	}

	return badges, nil
}

func (r *pgGamificationRepository) GetUserBadge(ctx context.Context, userID string, badgeID uuid.UUID) (*models.UserBadge, error) {
	const q = `
		select id, user_id, badge_id, session_id, earned_at
		from user_badges
		where user_id = $1 and badge_id = $2`

	var b models.UserBadge
	err := r.pool.QueryRow(ctx, q, userID, badgeID).Scan(&b.ID, &b.UserID, &b.BadgeID, &b.SessionID, &b.EarnedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user badge: %w", err)
	}

	return &b, nil
}

// AwardBadge is idempotent on (user_id, badge_id). created reports whether this
// call actually inserted a new row (false if the user already had the badge).
func (r *pgGamificationRepository) AwardBadge(ctx context.Context, b *models.UserBadge) (bool, error) {
	const q = `
		insert into user_badges (id, user_id, badge_id, session_id, earned_at)
		values ($1,$2,$3,$4,now())
		on conflict (user_id, badge_id) do nothing
		returning earned_at`

	err := r.pool.QueryRow(ctx, q, b.ID, b.UserID, b.BadgeID, b.SessionID).Scan(&b.EarnedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, getErr := r.GetUserBadge(ctx, b.UserID, b.BadgeID)
			if getErr != nil {
				return false, fmt.Errorf("repository: award badge: %w", getErr)
			}
			*b = *existing
			return false, nil
		}
		return false, fmt.Errorf("repository: award badge: %w", err)
	}

	return true, nil
}

func (r *pgGamificationRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	if _, err := r.pool.Exec(ctx, `delete from xp_transactions where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete xp transactions for user: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `delete from daily_activity where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete daily activity for user: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `delete from user_skill_stats where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete user skill stats for user: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `delete from user_badges where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete user badges for user: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `delete from leaderboard_snapshots where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete leaderboard snapshots for user: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `delete from user_stats where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete user stats for user: %w", err)
	}
	return nil
}
