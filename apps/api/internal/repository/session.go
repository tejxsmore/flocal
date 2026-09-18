package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type SessionRepository interface {
	Create(ctx context.Context, s *models.SpeakingSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SpeakingSession, error)
	ListForUser(ctx context.Context, userID string, limit, offset int) ([]models.SpeakingSession, error)
	ListHistoryForUser(ctx context.Context, userID string, filter models.SessionHistoryFilter, limit, offset int) ([]models.SessionHistoryItem, error)
	MarkStarted(ctx context.Context, id uuid.UUID) error
	MarkSubmitted(ctx context.Context, id uuid.UUID, audioS3Key string, audioDurationSeconds float64, audioFormat string, audioSizeBytes int64) error
	MarkCompleted(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, reason string) error
	CountActiveSessionsSince(ctx context.Context, userID string, since time.Time) (int64, error)
	DeleteAllForUser(ctx context.Context, userID string) error
	DeleteByID(ctx context.Context, id uuid.UUID, userID string) error
}

type pgSessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) SessionRepository {
	return &pgSessionRepository{pool: pool}
}

func (r *pgSessionRepository) Create(ctx context.Context, s *models.SpeakingSession) error {
	const q = `
		insert into speaking_sessions (
			id,
			user_id,
			topic_id,
			daily_challenge_id,
			debate_stance,
			prep_time_seconds,
			speak_time_seconds,
			status,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		returning retry_count, created_at, updated_at`

	err := r.pool.QueryRow(
		ctx,
		q,
		s.ID,
		s.UserID,
		s.TopicID,
		s.DailyChallengeID,
		s.DebateStance,
		s.PrepTimeSeconds,
		s.SpeakTimeSeconds,
		s.Status,
	).Scan(
		&s.RetryCount,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create session: %w", err)
	}

	return nil
}

func (r *pgSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SpeakingSession, error) {
	const q = `
		select
			id,
			user_id,
			topic_id,
			daily_challenge_id,
			client_upload_token,
			debate_stance,
			prep_time_seconds,
			speak_time_seconds,
			status,
			processing_step,
			processing_started_at,
			processing_worker_id,
			failure_reason,
			retry_count,
			last_error,
			last_retry_at,
			next_retry_at,
			audio_s3_key,
			audio_duration_seconds,
			audio_format,
			audio_size_bytes,
			audio_expires_at,
			audio_deleted_at,
			share_token_hash,
			share_enabled,
			share_expires_at,
			started_at,
			submitted_at,
			completed_at,
			created_at,
			updated_at
		from speaking_sessions
		where id = $1`

	var s models.SpeakingSession

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID,
		&s.UserID,
		&s.TopicID,
		&s.DailyChallengeID,
		&s.ClientUploadToken,
		&s.DebateStance,
		&s.PrepTimeSeconds,
		&s.SpeakTimeSeconds,
		&s.Status,
		&s.ProcessingStep,
		&s.ProcessingStartedAt,
		&s.ProcessingWorkerID,
		&s.FailureReason,
		&s.RetryCount,
		&s.LastError,
		&s.LastRetryAt,
		&s.NextRetryAt,
		&s.AudioS3Key,
		&s.AudioDurationSeconds,
		&s.AudioFormat,
		&s.AudioSizeBytes,
		&s.AudioExpiresAt,
		&s.AudioDeletedAt,
		&s.ShareTokenHash,
		&s.ShareEnabled,
		&s.ShareExpiresAt,
		&s.StartedAt,
		&s.SubmittedAt,
		&s.CompletedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("repository: get session by id: %w", err)
	}

	return &s, nil
}

func (r *pgSessionRepository) ListForUser(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]models.SpeakingSession, error) {
	const q = `
		select
			id,
			user_id,
			topic_id,
			daily_challenge_id,
			client_upload_token,
			debate_stance,
			prep_time_seconds,
			speak_time_seconds,
			status,
			processing_step,
			processing_started_at,
			processing_worker_id,
			failure_reason,
			retry_count,
			last_error,
			last_retry_at,
			next_retry_at,
			audio_s3_key,
			audio_duration_seconds,
			audio_format,
			audio_size_bytes,
			audio_expires_at,
			audio_deleted_at,
			share_token_hash,
			share_enabled,
			share_expires_at,
			started_at,
			submitted_at,
			completed_at,
			created_at,
			updated_at
		from speaking_sessions
		where user_id = $1
		order by created_at desc
		limit $2 offset $3`

	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list sessions for user: %w", err)
	}
	defer rows.Close()

	sessions := make([]models.SpeakingSession, 0)

	for rows.Next() {
		var s models.SpeakingSession

		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.TopicID,
			&s.DailyChallengeID,
			&s.ClientUploadToken,
			&s.DebateStance,
			&s.PrepTimeSeconds,
			&s.SpeakTimeSeconds,
			&s.Status,
			&s.ProcessingStep,
			&s.ProcessingStartedAt,
			&s.ProcessingWorkerID,
			&s.FailureReason,
			&s.RetryCount,
			&s.LastError,
			&s.LastRetryAt,
			&s.NextRetryAt,
			&s.AudioS3Key,
			&s.AudioDurationSeconds,
			&s.AudioFormat,
			&s.AudioSizeBytes,
			&s.AudioExpiresAt,
			&s.AudioDeletedAt,
			&s.ShareTokenHash,
			&s.ShareEnabled,
			&s.ShareExpiresAt,
			&s.StartedAt,
			&s.SubmittedAt,
			&s.CompletedAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan session: %w", err)
		}

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list sessions for user: %w", err)
	}

	return sessions, nil
}

func (r *pgSessionRepository) ListHistoryForUser(
	ctx context.Context,
	userID string,
	filter models.SessionHistoryFilter,
	limit int,
	offset int,
) ([]models.SessionHistoryItem, error) {
	query := `
		select
			s.id,
			s.topic_id,
			t.title,
			t.format,
			s.status,
			s.prep_time_seconds,
			s.speak_time_seconds,
			sa.overall_score,
			sa.words_per_minute,
			s.created_at,
			s.completed_at
		from speaking_sessions s
		join topics t on t.id = s.topic_id
		left join session_analysis sa on sa.session_id = s.id
		where s.user_id = $1`

	args := []interface{}{userID}
	argN := 2

	if len(filter.Status) > 0 {
		query += fmt.Sprintf(" and s.status::text = any($%d::text[])", argN)
		args = append(args, filter.Status)
		argN++
	}

	if len(filter.Format) > 0 {
		query += fmt.Sprintf(" and t.format::text = any($%d::text[])", argN)
		args = append(args, filter.Format)
		argN++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" and s.created_at >= $%d", argN)
		args = append(args, *filter.From)
		argN++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" and s.created_at <= $%d", argN)
		args = append(args, *filter.To)
		argN++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" and t.title ilike $%d", argN)
		args = append(args, "%"+filter.Search+"%")
		argN++
	}

	switch filter.Sort {
	case "oldest":
		query += " order by s.created_at asc"
	case "score_desc":
		query += " order by sa.overall_score desc nulls last"
	case "score_asc":
		query += " order by sa.overall_score asc nulls last"
	default:
		query += " order by s.created_at desc"
	}

	query += fmt.Sprintf(" limit $%d offset $%d", argN, argN+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: list history for user: %w", err)
	}
	defer rows.Close()

	items := make([]models.SessionHistoryItem, 0)

	for rows.Next() {
		var item models.SessionHistoryItem

		if err := rows.Scan(
			&item.ID,
			&item.TopicID,
			&item.TopicTitle,
			&item.TopicFormat,
			&item.Status,
			&item.PrepTimeSeconds,
			&item.SpeakTimeSeconds,
			&item.OverallScore,
			&item.WordsPerMinute,
			&item.CreatedAt,
			&item.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan history item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list history for user: %w", err)
	}

	return items, nil
}

func (r *pgSessionRepository) MarkStarted(ctx context.Context, id uuid.UUID) error {
	const q = `
		update speaking_sessions
		set started_at = now()
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repository: mark session started: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgSessionRepository) MarkSubmitted(
	ctx context.Context,
	id uuid.UUID,
	audioS3Key string,
	audioDurationSeconds float64,
	audioFormat string,
	audioSizeBytes int64,
) error {
	const q = `
		update speaking_sessions
		set status = $2,
		    submitted_at = now(),
		    processing_started_at = now(),
		    audio_s3_key = $3,
		    audio_duration_seconds = $4,
		    audio_format = $5,
		    audio_size_bytes = $6
		where id = $1`

	result, err := r.pool.Exec(
		ctx,
		q,
		id,
		models.SessionStatusProcessing,
		audioS3Key,
		audioDurationSeconds,
		audioFormat,
		audioSizeBytes,
	)
	if err != nil {
		return fmt.Errorf("repository: mark session submitted: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgSessionRepository) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	const q = `
		update speaking_sessions
		set status = $2,
		    completed_at = now(),
		    processing_step = null,
		    processing_started_at = null,
		    processing_worker_id = null,
		    failure_reason = null,
		    last_error = null,
		    next_retry_at = null
		where id = $1`

	result, err := r.pool.Exec(
		ctx,
		q,
		id,
		models.SessionStatusCompleted,
	)
	if err != nil {
		return fmt.Errorf("repository: mark session completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgSessionRepository) MarkFailed(
	ctx context.Context,
	id uuid.UUID,
	reason string,
) error {
	const q = `
		update speaking_sessions
		set status = $2,
		    failure_reason = $3,
		    last_error = $3,
		    last_retry_at = now(),
		    retry_count = retry_count + 1,
		    processing_started_at = null,
		    processing_worker_id = null
		where id = $1`

	result, err := r.pool.Exec(
		ctx,
		q,
		id,
		models.SessionStatusFailed,
		reason,
	)
	if err != nil {
		return fmt.Errorf("repository: mark session failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgSessionRepository) CountActiveSessionsSince(ctx context.Context, userID string, since time.Time) (int64, error) {
	const q = `
		select count(*)
		from speaking_sessions
		where user_id = $1
		  and created_at >= $2
		  and status in ($3, $4)`

	var count int64
	err := r.pool.QueryRow(ctx, q, userID, since, models.SessionStatusProcessing, models.SessionStatusCompleted).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("repository: count active sessions since: %w", err)
	}

	return count, nil
}

func (r *pgSessionRepository) DeleteByID(ctx context.Context, id uuid.UUID, userID string) error {
	const q = `delete from speaking_sessions where id = $1 and user_id = $2`

	result, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("repository: delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgSessionRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	const q = `delete from speaking_sessions where user_id = $1`
	_, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("repository: delete all sessions for user: %w", err)
	}
	return nil
}
