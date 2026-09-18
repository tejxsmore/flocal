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

type TopicRepository interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.TopicCategory, error)
	ListActiveCategories(ctx context.Context) ([]models.TopicCategory, error)
	RandomActive(ctx context.Context, categoryID *uuid.UUID, format *models.TopicFormat) (*models.Topic, error)
	RecordSpin(ctx context.Context, spin *models.TopicSpin) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Topic, error)
}

type pgTopicRepository struct {
	pool *pgxpool.Pool
}

func NewTopicRepository(pool *pgxpool.Pool) TopicRepository {
	return &pgTopicRepository{
		pool: pool,
	}
}

func (r *pgTopicRepository) GetCategoryByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.TopicCategory, error) {
	const q = `
		select
			id,
			slug,
			name,
			icon,
			is_active,
			sort_order,
			created_at
		from topic_categories
		where id = $1`

	var category models.TopicCategory

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&category.ID,
		&category.Slug,
		&category.Name,
		&category.Icon,
		&category.IsActive,
		&category.SortOrder,
		&category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("repository: get category by id: %w", err)
	}

	return &category, nil
}

func (r *pgTopicRepository) ListActiveCategories(
	ctx context.Context,
) ([]models.TopicCategory, error) {
	const q = `
		select
			id,
			slug,
			name,
			icon,
			is_active,
			sort_order,
			created_at
		from topic_categories
		where is_active = true
		order by sort_order asc, name asc`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repository: list active categories: %w", err)
	}
	defer rows.Close()

	categories := make([]models.TopicCategory, 0)

	for rows.Next() {
		var category models.TopicCategory

		if err := rows.Scan(
			&category.ID,
			&category.Slug,
			&category.Name,
			&category.Icon,
			&category.IsActive,
			&category.SortOrder,
			&category.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list active categories: %w", err)
	}

	return categories, nil
}

func (r *pgTopicRepository) RandomActive(
	ctx context.Context,
	categoryID *uuid.UUID,
	format *models.TopicFormat,
) (*models.Topic, error) {
	const q = `
		select
			id,
			category_id,
			title,
			format,
			difficulty,
			recommended_prep_seconds,
			tags,
			metadata,
			is_active,
			is_premium,
			source,
			language_code,
			submitted_by,
			deleted_at,
			created_at,
			updated_at
		from topics
		where is_active = true
		  and deleted_at is null
		  and ($1::uuid is null or category_id = $1)
		  and ($2::topic_format is null or format = $2)
		order by random()
		limit 1`

	var topic models.Topic

	err := r.pool.QueryRow(
		ctx,
		q,
		categoryID,
		format,
	).Scan(
		&topic.ID,
		&topic.CategoryID,
		&topic.Title,
		&topic.Format,
		&topic.Difficulty,
		&topic.RecommendedPrepSeconds,
		&topic.Tags,
		&topic.Metadata,
		&topic.IsActive,
		&topic.IsPremium,
		&topic.Source,
		&topic.LanguageCode,
		&topic.SubmittedBy,
		&topic.DeletedAt,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("repository: random active topic: %w", err)
	}

	return &topic, nil
}

func (r *pgTopicRepository) RecordSpin(
	ctx context.Context,
	spin *models.TopicSpin,
) error {
	const q = `
		insert into topic_spins (
			id,
			user_id,
			topic_id,
			category_filter_id,
			format_filter,
			session_id,
			created_at
		)
		values ($1, $2, $3, $4, $5, $6, now())
		returning created_at`

	err := r.pool.QueryRow(
		ctx,
		q,
		spin.ID,
		spin.UserID,
		spin.TopicID,
		spin.CategoryFilterID,
		spin.FormatFilter,
		spin.SessionID,
	).Scan(
		&spin.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: record spin: %w", err)
	}

	return nil
}

func (r *pgTopicRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Topic, error) {
	const q = `
		select
			id,
			category_id,
			title,
			format,
			difficulty,
			recommended_prep_seconds,
			tags,
			metadata,
			is_active,
			is_premium,
			source,
			language_code,
			submitted_by,
			deleted_at,
			created_at,
			updated_at
		from topics
		where id = $1`

	var topic models.Topic

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&topic.ID,
		&topic.CategoryID,
		&topic.Title,
		&topic.Format,
		&topic.Difficulty,
		&topic.RecommendedPrepSeconds,
		&topic.Tags,
		&topic.Metadata,
		&topic.IsActive,
		&topic.IsPremium,
		&topic.Source,
		&topic.LanguageCode,
		&topic.SubmittedBy,
		&topic.DeletedAt,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("repository: get topic by id: %w", err)
	}

	return &topic, nil
}
