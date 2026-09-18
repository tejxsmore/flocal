package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type OnboardingRepository interface {
	CompleteProfile(ctx context.Context, userID, name string, username, image, countryCode *string) error
	UpsertPreferences(ctx context.Context, p *models.UserPreferences) error
	GetPreferences(ctx context.Context, userID string) (*models.UserPreferences, error)
	ReplaceFocusAreas(ctx context.Context, userID string, areas []models.FocusArea) error
	ListFocusAreas(ctx context.Context, userID string) ([]models.UserFocusArea, error)
}

type pgOnboardingRepository struct {
	pool *pgxpool.Pool
}

func NewOnboardingRepository(pool *pgxpool.Pool) OnboardingRepository {
	return &pgOnboardingRepository{pool: pool}
}

func (r *pgOnboardingRepository) CompleteProfile(ctx context.Context, userID, name string, username, image, countryCode *string) error {
	const q = `
		update "user"
		set name = $2,
		    username = $3,
		    image = $4,
		    country_code = coalesce($5, country_code),
		    onboarding_completed = true,
		    updated_at = now()
		where id = $1`

	ct, err := r.pool.Exec(ctx, q, userID, name, username, image, countryCode)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("repository: complete profile: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgOnboardingRepository) UpsertPreferences(ctx context.Context, p *models.UserPreferences) error {
	const q = `
		insert into user_preferences (user_id, primary_goal, daily_time_commitment, completed_at, created_at, updated_at)
		values ($1, $2, $3, $4, now(), now())
		on conflict (user_id) do update
		set primary_goal = excluded.primary_goal,
		    daily_time_commitment = excluded.daily_time_commitment,
		    completed_at = excluded.completed_at,
		    updated_at = now()
		returning created_at, updated_at`

	err := r.pool.QueryRow(ctx, q, p.UserID, p.PrimaryGoal, p.DailyTimeCommitment, p.CompletedAt).
		Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository: upsert preferences: %w", err)
	}
	return nil
}

func (r *pgOnboardingRepository) GetPreferences(ctx context.Context, userID string) (*models.UserPreferences, error) {
	const q = `
		select user_id, primary_goal, daily_time_commitment, completed_at, created_at, updated_at
		from user_preferences
		where user_id = $1`

	var p models.UserPreferences
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.PrimaryGoal, &p.DailyTimeCommitment, &p.CompletedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get preferences: %w", err)
	}
	return &p, nil
}

func (r *pgOnboardingRepository) ReplaceFocusAreas(ctx context.Context, userID string, areas []models.FocusArea) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin replace focus areas: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from user_focus_areas where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: clear focus areas: %w", err)
	}

	for _, area := range areas {
		if _, err := tx.Exec(ctx,
			`insert into user_focus_areas (user_id, focus_area, created_at) values ($1, $2, now())`,
			userID, area,
		); err != nil {
			return fmt.Errorf("repository: insert focus area: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: commit replace focus areas: %w", err)
	}
	return nil
}

func (r *pgOnboardingRepository) ListFocusAreas(ctx context.Context, userID string) ([]models.UserFocusArea, error) {
	const q = `
		select user_id, focus_area, created_at
		from user_focus_areas
		where user_id = $1
		order by created_at asc`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list focus areas: %w", err)
	}
	defer rows.Close()

	areas := make([]models.UserFocusArea, 0)
	for rows.Next() {
		var f models.UserFocusArea
		if err := rows.Scan(&f.UserID, &f.FocusArea, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan focus area: %w", err)
		}
		areas = append(areas, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list focus areas: %w", err)
	}
	return areas, nil
}
