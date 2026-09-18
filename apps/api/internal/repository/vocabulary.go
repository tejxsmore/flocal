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

type VocabularyEntry struct {
	WordID         uuid.UUID               `json:"wordId"`
	Word           string                  `json:"word"`
	Definition     *string                 `json:"definition,omitempty"`
	Difficulty     *models.TopicDifficulty `json:"difficulty,omitempty"`
	TimesSeen      int                     `json:"timesSeen"`
	TimesSuggested int                     `json:"timesSuggested"`
	Mastered       bool                    `json:"mastered"`
	FirstSeenAt    time.Time               `json:"firstSeenAt"`
	LastSeenAt     time.Time               `json:"lastSeenAt"`
}

type VocabularyRepository interface {
	GetOrCreateWord(ctx context.Context, word string, definition *string, difficulty *models.TopicDifficulty) (*models.VocabularyWord, error)
	GetWordByID(ctx context.Context, id uuid.UUID) (*models.VocabularyWord, error)
	RecordEncounter(ctx context.Context, userID string, wordID uuid.UUID, suggested bool) (*models.UserVocabulary, error)
	GetUserVocabulary(ctx context.Context, userID string, wordID uuid.UUID) (*models.UserVocabulary, error)
	ListUserVocabulary(ctx context.Context, userID string, mastered *bool, limit, offset int) ([]models.UserVocabulary, error)
	ListUserVocabularyWithWords(ctx context.Context, userID string, mastered *bool, limit, offset int) ([]VocabularyEntry, error)
	CountUserVocabulary(ctx context.Context, userID string, mastered *bool) (int, error)
	SetMastered(ctx context.Context, userID string, wordID uuid.UUID, mastered bool) (*models.UserVocabulary, error)
	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgVocabularyRepository struct {
	pool *pgxpool.Pool
}

func NewVocabularyRepository(pool *pgxpool.Pool) VocabularyRepository {
	return &pgVocabularyRepository{pool: pool}
}

func (r *pgVocabularyRepository) GetOrCreateWord(
	ctx context.Context,
	word string,
	definition *string,
	difficulty *models.TopicDifficulty,
) (*models.VocabularyWord, error) {
	const q = `
		insert into vocabulary_words (id, word, definition, difficulty, created_at)
		values (gen_random_uuid(), $1, $2, $3, now())
		on conflict (word) do update
			set definition = coalesce(excluded.definition, vocabulary_words.definition)
		returning id, word, definition, difficulty, created_at`

	var w models.VocabularyWord

	err := r.pool.QueryRow(ctx, q, word, definition, difficulty).Scan(
		&w.ID, &w.Word, &w.Definition, &w.Difficulty, &w.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: get or create word: %w", err)
	}

	return &w, nil
}

func (r *pgVocabularyRepository) GetWordByID(ctx context.Context, id uuid.UUID) (*models.VocabularyWord, error) {
	const q = `
		select id, word, definition, difficulty, created_at
		from vocabulary_words
		where id = $1`

	var w models.VocabularyWord

	err := r.pool.QueryRow(ctx, q, id).Scan(&w.ID, &w.Word, &w.Definition, &w.Difficulty, &w.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get word by id: %w", err)
	}

	return &w, nil
}

func (r *pgVocabularyRepository) RecordEncounter(
	ctx context.Context,
	userID string,
	wordID uuid.UUID,
	suggested bool,
) (*models.UserVocabulary, error) {
	const q = `
		insert into user_vocabulary (id, user_id, word_id, times_seen, times_suggested, mastered, first_seen_at, last_seen_at)
		values (gen_random_uuid(), $1, $2, 1, case when $3 then 1 else 0 end, false, now(), now())
		on conflict (user_id, word_id) do update
			set times_seen = user_vocabulary.times_seen + 1,
				times_suggested = user_vocabulary.times_suggested + case when $3 then 1 else 0 end,
				last_seen_at = now()
		returning id, user_id, word_id, times_seen, times_suggested, mastered, first_seen_at, last_seen_at`

	var v models.UserVocabulary

	err := r.pool.QueryRow(ctx, q, userID, wordID, suggested).Scan(
		&v.ID, &v.UserID, &v.WordID, &v.TimesSeen, &v.TimesSuggested, &v.Mastered, &v.FirstSeenAt, &v.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: record encounter: %w", err)
	}

	return &v, nil
}

func (r *pgVocabularyRepository) GetUserVocabulary(ctx context.Context, userID string, wordID uuid.UUID) (*models.UserVocabulary, error) {
	const q = `
		select id, user_id, word_id, times_seen, times_suggested, mastered, first_seen_at, last_seen_at
		from user_vocabulary
		where user_id = $1 and word_id = $2`

	var v models.UserVocabulary

	err := r.pool.QueryRow(ctx, q, userID, wordID).Scan(
		&v.ID, &v.UserID, &v.WordID, &v.TimesSeen, &v.TimesSuggested, &v.Mastered, &v.FirstSeenAt, &v.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user vocabulary: %w", err)
	}

	return &v, nil
}

func (r *pgVocabularyRepository) ListUserVocabulary(ctx context.Context, userID string, mastered *bool, limit, offset int) ([]models.UserVocabulary, error) {
	const q = `
		select id, user_id, word_id, times_seen, times_suggested, mastered, first_seen_at, last_seen_at
		from user_vocabulary
		where user_id = $1
		  and ($2::boolean is null or mastered = $2)
		order by last_seen_at desc
		limit $3 offset $4`

	rows, err := r.pool.Query(ctx, q, userID, mastered, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list user vocabulary: %w", err)
	}
	defer rows.Close()

	entries := make([]models.UserVocabulary, 0)
	for rows.Next() {
		var v models.UserVocabulary
		if err := rows.Scan(&v.ID, &v.UserID, &v.WordID, &v.TimesSeen, &v.TimesSuggested, &v.Mastered, &v.FirstSeenAt, &v.LastSeenAt); err != nil {
			return nil, fmt.Errorf("repository: scan user vocabulary: %w", err)
		}
		entries = append(entries, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list user vocabulary: %w", err)
	}

	return entries, nil
}

func (r *pgVocabularyRepository) ListUserVocabularyWithWords(ctx context.Context, userID string, mastered *bool, limit, offset int) ([]VocabularyEntry, error) {
	const q = `
		select
			w.id,
			w.word,
			w.definition,
			w.difficulty,
			uv.times_seen,
			uv.times_suggested,
			uv.mastered,
			uv.first_seen_at,
			uv.last_seen_at
		from user_vocabulary uv
		join vocabulary_words w on w.id = uv.word_id
		where uv.user_id = $1
		  and ($2::boolean is null or uv.mastered = $2)
		order by uv.last_seen_at desc
		limit $3 offset $4`

	rows, err := r.pool.Query(ctx, q, userID, mastered, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list user vocabulary with words: %w", err)
	}
	defer rows.Close()

	entries := make([]VocabularyEntry, 0)
	for rows.Next() {
		var e VocabularyEntry
		if err := rows.Scan(
			&e.WordID, &e.Word, &e.Definition, &e.Difficulty,
			&e.TimesSeen, &e.TimesSuggested, &e.Mastered, &e.FirstSeenAt, &e.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan vocabulary entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list user vocabulary with words: %w", err)
	}

	return entries, nil
}

func (r *pgVocabularyRepository) CountUserVocabulary(ctx context.Context, userID string, mastered *bool) (int, error) {
	const q = `
		select count(*)
		from user_vocabulary
		where user_id = $1
		  and ($2::boolean is null or mastered = $2)`

	var count int
	if err := r.pool.QueryRow(ctx, q, userID, mastered).Scan(&count); err != nil {
		return 0, fmt.Errorf("repository: count user vocabulary: %w", err)
	}

	return count, nil
}

func (r *pgVocabularyRepository) SetMastered(ctx context.Context, userID string, wordID uuid.UUID, mastered bool) (*models.UserVocabulary, error) {
	const q = `
		update user_vocabulary
		set mastered = $3
		where user_id = $1 and word_id = $2
		returning id, user_id, word_id, times_seen, times_suggested, mastered, first_seen_at, last_seen_at`

	var v models.UserVocabulary

	err := r.pool.QueryRow(ctx, q, userID, wordID, mastered).Scan(
		&v.ID, &v.UserID, &v.WordID, &v.TimesSeen, &v.TimesSuggested, &v.Mastered, &v.FirstSeenAt, &v.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: set mastered: %w", err)
	}

	return &v, nil
}

func (r *pgVocabularyRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	if _, err := r.pool.Exec(ctx, `delete from user_vocabulary where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete user vocabulary for user: %w", err)
	}
	return nil
}
