package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type AuditRepository interface {
	Create(ctx context.Context, a *models.AuditLog) error
	ListByEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]models.AuditLog, error)
	ListByActor(ctx context.Context, actorUserID string, limit, offset int) ([]models.AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]models.AuditLog, error)
}

type pgAuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) AuditRepository {
	return &pgAuditRepository{pool: pool}
}

func (r *pgAuditRepository) Create(ctx context.Context, a *models.AuditLog) error {
	const q = `
		insert into audit_log (id, actor_user_id, action, entity_type, entity_id, old_value, new_value, source, created_at)
		values (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, now())
		returning id, created_at`

	err := r.pool.QueryRow(ctx, q,
		a.ActorUserID, a.Action, a.EntityType, a.EntityID, a.OldValue, a.NewValue, a.Source,
	).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create audit log: %w", err)
	}
	return nil
}

func (r *pgAuditRepository) ListByEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]models.AuditLog, error) {
	const q = `
		select id, actor_user_id, action, entity_type, entity_id, old_value, new_value, source, created_at
		from audit_log
		where entity_type = $1 and entity_id = $2
		order by created_at desc
		limit $3 offset $4`

	rows, err := r.pool.Query(ctx, q, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list audit logs by entity: %w", err)
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func (r *pgAuditRepository) ListByActor(ctx context.Context, actorUserID string, limit, offset int) ([]models.AuditLog, error) {
	const q = `
		select id, actor_user_id, action, entity_type, entity_id, old_value, new_value, source, created_at
		from audit_log
		where actor_user_id = $1
		order by created_at desc
		limit $2 offset $3`

	rows, err := r.pool.Query(ctx, q, actorUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list audit logs by actor: %w", err)
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func (r *pgAuditRepository) List(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	const q = `
		select id, actor_user_id, action, entity_type, entity_id, old_value, new_value, source, created_at
		from audit_log
		order by created_at desc
		limit $1 offset $2`

	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list audit logs: %w", err)
	}
	defer rows.Close()

	return scanAuditLogs(rows)
}

func scanAuditLogs(rows pgx.Rows) ([]models.AuditLog, error) {
	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		var a models.AuditLog
		if err := rows.Scan(
			&a.ID, &a.ActorUserID, &a.Action, &a.EntityType, &a.EntityID,
			&a.OldValue, &a.NewValue, &a.Source, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan audit log: %w", err)
		}
		logs = append(logs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list audit logs: %w", err)
	}
	return logs, nil
}
