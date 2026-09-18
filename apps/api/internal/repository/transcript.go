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

type TranscriptRepository interface {
	Create(ctx context.Context, t *models.Transcript) error
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.Transcript, error)
	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgTranscriptRepository struct {
	pool *pgxpool.Pool
}

func NewTranscriptRepository(pool *pgxpool.Pool) TranscriptRepository {
	return &pgTranscriptRepository{
		pool: pool,
	}
}

func (r *pgTranscriptRepository) Create(
	ctx context.Context,
	t *models.Transcript,
) error {
	const q = `
		insert into transcripts (
			id,
			session_id,
			raw_text,
			word_count,
			language,
			avg_confidence,
			word_timings,
			created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, now())
		returning created_at`

	err := r.pool.QueryRow(
		ctx,
		q,
		t.ID,
		t.SessionID,
		t.RawText,
		t.WordCount,
		t.Language,
		t.AvgConfidence,
		t.WordTimings,
	).Scan(
		&t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create transcript: %w", err)
	}

	return nil
}

func (r *pgTranscriptRepository) GetBySessionID(
	ctx context.Context,
	sessionID uuid.UUID,
) (*models.Transcript, error) {
	const q = `
		select
			id,
			session_id,
			raw_text,
			word_count,
			language,
			avg_confidence,
			word_timings,
			created_at
		from transcripts
		where session_id = $1`

	var transcript models.Transcript

	err := r.pool.QueryRow(
		ctx,
		q,
		sessionID,
	).Scan(
		&transcript.ID,
		&transcript.SessionID,
		&transcript.RawText,
		&transcript.WordCount,
		&transcript.Language,
		&transcript.AvgConfidence,
		&transcript.WordTimings,
		&transcript.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf(
			"repository: get transcript by session id: %w",
			err,
		)
	}

	return &transcript, nil
}

func (r *pgTranscriptRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	const q = `
		delete from transcripts
		where session_id in (
			select id from speaking_sessions where user_id = $1
		)`

	_, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("repository: delete transcripts for user: %w", err)
	}

	return nil
}
