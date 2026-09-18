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

type ViewRepository interface {
	GetUserCurrentPlan(ctx context.Context, userID string) (*models.UserCurrentPlan, error)
	GetSessionReport(ctx context.Context, sessionID uuid.UUID) (*models.SessionReport, error)
	GetTopicUsageStats(ctx context.Context, topicID uuid.UUID) (*models.TopicUsageStats, error)
	ListTopicUsageStats(ctx context.Context, limit int) ([]models.TopicUsageStats, error)
	GetLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error)
	GetCountryLeaderboard(ctx context.Context, countryCode string, limit, offset int) ([]models.CountryLeaderboardEntry, error)
	GetMyLeaderboardRank(ctx context.Context, userID string) (*models.LeaderboardEntry, error)
}

type pgViewRepository struct {
	pool *pgxpool.Pool
}

func NewViewRepository(pool *pgxpool.Pool) ViewRepository {
	return &pgViewRepository{pool: pool}
}

func (r *pgViewRepository) GetUserCurrentPlan(ctx context.Context, userID string) (*models.UserCurrentPlan, error) {
	const q = `
		select user_id, plan_slug, daily_session_limit, can_view_analysis, plan_renews_at, subscription_status
		from user_current_plan
		where user_id = $1`

	var p models.UserCurrentPlan
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.PlanSlug, &p.DailySessionLimit, &p.CanViewAnalysis, &p.PlanRenewsAt, &p.SubscriptionStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user current plan: %w", err)
	}
	return &p, nil
}

func (r *pgViewRepository) GetSessionReport(ctx context.Context, sessionID uuid.UUID) (*models.SessionReport, error) {
	const q = `
		select session_id, user_id, topic_id, topic_title, topic_format, topic_category, status,
		       debate_stance, prep_time_seconds, speak_time_seconds, audio_s3_key, audio_expires_at,
		       share_token_hash, share_enabled, share_expires_at, transcript,
		       words_per_minute, filler_word_count, filler_rate, pause_rate, longest_pause_seconds,
		       unique_word_count, lexical_diversity, sentence_count, average_sentence_length,
		       overall_score, clarity_score, delivery_score, content_score, vocabulary_score, grammar_score,
		       strengths, improvements, coach_message, created_at, completed_at
		from session_report
		where session_id = $1`

	var rep models.SessionReport
	err := r.pool.QueryRow(ctx, q, sessionID).Scan(
		&rep.SessionID, &rep.UserID, &rep.TopicID, &rep.TopicTitle, &rep.TopicFormat, &rep.TopicCategory, &rep.Status,
		&rep.DebateStance, &rep.PrepTimeSeconds, &rep.SpeakTimeSeconds, &rep.AudioS3Key, &rep.AudioExpiresAt,
		&rep.ShareTokenHash, &rep.ShareEnabled, &rep.ShareExpiresAt, &rep.Transcript,
		&rep.WordsPerMinute, &rep.FillerWordCount, &rep.FillerRate, &rep.PauseRate, &rep.LongestPauseSeconds,
		&rep.UniqueWordCount, &rep.LexicalDiversity, &rep.SentenceCount, &rep.AverageSentenceLength,
		&rep.OverallScore, &rep.ClarityScore, &rep.DeliveryScore, &rep.ContentScore, &rep.VocabularyScore, &rep.GrammarScore,
		&rep.Strengths, &rep.Improvements, &rep.CoachMessage, &rep.CreatedAt, &rep.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get session report: %w", err)
	}
	return &rep, nil
}

func (r *pgViewRepository) GetTopicUsageStats(ctx context.Context, topicID uuid.UUID) (*models.TopicUsageStats, error) {
	const q = `
		select
			t.id as topic_id,
			t.title as topic_title,
			coalesce(v.spin_count, 0) as spin_count
		from topics t
		left join topic_usage_stats v on v.topic_id = t.id
		where t.id = $1`

	var t models.TopicUsageStats
	err := r.pool.QueryRow(ctx, q, topicID).Scan(&t.TopicID, &t.TopicTitle, &t.SpinCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get topic usage stats: %w", err)
	}
	return &t, nil
}

func (r *pgViewRepository) ListTopicUsageStats(ctx context.Context, limit int) ([]models.TopicUsageStats, error) {
	const q = `
		select
			t.id as topic_id,
			t.title as topic_title,
			v.spin_count
		from topic_usage_stats v
		join topics t on t.id = v.topic_id
		order by v.spin_count desc, t.title asc
		limit $1`

	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("repository: list topic usage stats: %w", err)
	}
	defer rows.Close()

	stats := make([]models.TopicUsageStats, 0)
	for rows.Next() {
		var t models.TopicUsageStats
		if err := rows.Scan(&t.TopicID, &t.TopicTitle, &t.SpinCount); err != nil {
			return nil, fmt.Errorf("repository: scan topic usage stats: %w", err)
		}
		stats = append(stats, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list topic usage stats: %w", err)
	}
	return stats, nil
}

func (r *pgViewRepository) GetLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	const q = `
		select user_id, username, image, country_code, total_xp, current_streak_days, total_sessions, rank
		from leaderboard
		order by rank asc
		limit $1 offset $2`

	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: get leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]models.LeaderboardEntry, 0)
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(
			&e.UserID, &e.Username, &e.Image, &e.CountryCode,
			&e.TotalXP, &e.CurrentStreakDays, &e.TotalSessions, &e.Rank,
		); err != nil {
			return nil, fmt.Errorf("repository: scan leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: get leaderboard: %w", err)
	}
	return entries, nil
}

func (r *pgViewRepository) GetCountryLeaderboard(ctx context.Context, countryCode string, limit, offset int) ([]models.CountryLeaderboardEntry, error) {
	const q = `
		select user_id, username, image, country_code, total_xp, current_streak_days, total_sessions, rank
		from country_leaderboard
		where country_code = $1
		order by rank asc
		limit $2 offset $3`

	rows, err := r.pool.Query(ctx, q, countryCode, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: get country leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]models.CountryLeaderboardEntry, 0)
	for rows.Next() {
		var e models.CountryLeaderboardEntry
		if err := rows.Scan(
			&e.UserID, &e.Username, &e.Image, &e.CountryCode,
			&e.TotalXP, &e.CurrentStreakDays, &e.TotalSessions, &e.Rank,
		); err != nil {
			return nil, fmt.Errorf("repository: scan country leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: get country leaderboard: %w", err)
	}
	return entries, nil
}

func (r *pgViewRepository) GetMyLeaderboardRank(ctx context.Context, userID string) (*models.LeaderboardEntry, error) {
	const q = `
		select user_id, username, image, country_code, total_xp, current_streak_days, total_sessions, rank
		from leaderboard
		where user_id = $1`

	var e models.LeaderboardEntry
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&e.UserID, &e.Username, &e.Image, &e.CountryCode,
		&e.TotalXP, &e.CurrentStreakDays, &e.TotalSessions, &e.Rank,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get my leaderboard rank: %w", err)
	}
	return &e, nil
}
