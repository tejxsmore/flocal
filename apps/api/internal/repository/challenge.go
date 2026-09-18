package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type ChallengeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.DailyChallenge, error)
	GetForDate(ctx context.Context, date models.DateOnly, languageCode string, difficulty models.TopicDifficulty) (*models.DailyChallenge, error)
	Create(ctx context.Context, c *models.DailyChallenge) error
	GetCompletion(ctx context.Context, userID string, challengeID uuid.UUID) (*models.DailyChallengeCompletion, error)
	CreateCompletion(ctx context.Context, c *models.DailyChallengeCompletion) (created bool, err error)
	ListCompletionsForUser(ctx context.Context, userID string, limit, offset int) ([]models.DailyChallengeCompletion, error)
	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgChallengeRepository struct {
	pool *pgxpool.Pool
}

func NewChallengeRepository(pool *pgxpool.Pool) ChallengeRepository {
	return &pgChallengeRepository{pool: pool}
}

func (r *pgChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.DailyChallenge, error) {
	const q = `
		select id, topic_id, challenge_date, language_code, difficulty,
		       speak_time_seconds, xp_reward, created_at
		from daily_challenges
		where id = $1`

	var c models.DailyChallenge

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&c.ID, &c.TopicID, &c.ChallengeDate, &c.LanguageCode, &c.Difficulty,
		&c.SpeakTimeSeconds, &c.XPReward, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get daily challenge by id: %w", err)
	}

	return &c, nil
}

func (r *pgChallengeRepository) GetForDate(ctx context.Context, date models.DateOnly, languageCode string, difficulty models.TopicDifficulty) (*models.DailyChallenge, error) {
	const q = `
		select id, topic_id, challenge_date, language_code, difficulty,
		       speak_time_seconds, xp_reward, created_at
		from daily_challenges
		where challenge_date = $1 and language_code = $2 and difficulty = $3`

	var c models.DailyChallenge

	err := r.pool.QueryRow(ctx, q, date, languageCode, difficulty).Scan(
		&c.ID, &c.TopicID, &c.ChallengeDate, &c.LanguageCode, &c.Difficulty,
		&c.SpeakTimeSeconds, &c.XPReward, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get daily challenge for date: %w", err)
	}

	return &c, nil
}

func (r *pgChallengeRepository) Create(ctx context.Context, c *models.DailyChallenge) error {
	const q = `
		insert into daily_challenges (id, topic_id, challenge_date, language_code, difficulty, speak_time_seconds, xp_reward, created_at)
		values ($1,$2,$3,$4,$5,$6,$7,now())
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		c.ID, c.TopicID, c.ChallengeDate, c.LanguageCode, c.Difficulty, c.SpeakTimeSeconds, c.XPReward,
	).Scan(&c.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return fmt.Errorf("repository: create daily challenge: %w", err)
	}

	return nil
}

func (r *pgChallengeRepository) GetCompletion(ctx context.Context, userID string, challengeID uuid.UUID) (*models.DailyChallengeCompletion, error) {
	const q = `
		select id, user_id, daily_challenge_id, session_id, completed_at
		from daily_challenge_completions
		where user_id = $1 and daily_challenge_id = $2`

	var c models.DailyChallengeCompletion

	err := r.pool.QueryRow(ctx, q, userID, challengeID).Scan(
		&c.ID, &c.UserID, &c.DailyChallengeID, &c.SessionID, &c.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get daily challenge completion: %w", err)
	}

	return &c, nil
}

func (r *pgChallengeRepository) CreateCompletion(ctx context.Context, c *models.DailyChallengeCompletion) (bool, error) {
	const q = `
		insert into daily_challenge_completions (id, user_id, daily_challenge_id, session_id, completed_at)
		values ($1,$2,$3,$4,now())
		on conflict (user_id, daily_challenge_id) do nothing
		returning completed_at`

	err := r.pool.QueryRow(ctx, q, c.ID, c.UserID, c.DailyChallengeID, c.SessionID).Scan(&c.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, getErr := r.GetCompletion(ctx, c.UserID, c.DailyChallengeID)
			if getErr != nil {
				return false, fmt.Errorf("repository: create daily challenge completion: %w", getErr)
			}
			*c = *existing
			return false, nil
		}
		return false, fmt.Errorf("repository: create daily challenge completion: %w", err)
	}

	return true, nil
}

func (r *pgChallengeRepository) ListCompletionsForUser(ctx context.Context, userID string, limit, offset int) ([]models.DailyChallengeCompletion, error) {
	const q = `
		select id, user_id, daily_challenge_id, session_id, completed_at
		from daily_challenge_completions
		where user_id = $1
		order by completed_at desc
		limit $2 offset $3`

	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list daily challenge completions: %w", err)
	}
	defer rows.Close()

	completions := make([]models.DailyChallengeCompletion, 0)
	for rows.Next() {
		var c models.DailyChallengeCompletion
		if err := rows.Scan(&c.ID, &c.UserID, &c.DailyChallengeID, &c.SessionID, &c.CompletedAt); err != nil {
			return nil, fmt.Errorf("repository: scan daily challenge completion: %w", err)
		}
		completions = append(completions, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list daily challenge completions: %w", err)
	}

	return completions, nil
}

func (r *pgChallengeRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	if _, err := r.pool.Exec(ctx, `delete from daily_challenge_completions where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete daily challenge completions for user: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
