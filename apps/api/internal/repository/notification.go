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

type NotificationRepository interface {
	GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error)
	UpsertPreferences(ctx context.Context, p *models.NotificationPreferences) error
	Create(ctx context.Context, n *models.Notification) error
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*models.Notification, error)
	ListForUser(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]models.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, id uuid.UUID, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status models.NotificationDeliveryStatus, providerMessageID *string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) NotificationRepository {
	return &pgNotificationRepository{pool: pool}
}

func (r *pgNotificationRepository) GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error) {
	const q = `
		select user_id, daily_reminder_enabled, streak_risk_enabled, session_ready_enabled, reminder_time, updated_at
		from notification_preferences
		where user_id = $1`

	var p models.NotificationPreferences

	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.DailyReminderEnabled, &p.StreakRiskEnabled, &p.SessionReadyEnabled, &p.ReminderTime, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get notification preferences: %w", err)
	}

	return &p, nil
}

func (r *pgNotificationRepository) UpsertPreferences(ctx context.Context, p *models.NotificationPreferences) error {
	const q = `
		insert into notification_preferences (user_id, daily_reminder_enabled, streak_risk_enabled, session_ready_enabled, reminder_time, updated_at)
		values ($1, $2, $3, $4, $5, now())
		on conflict (user_id) do update
			set daily_reminder_enabled = excluded.daily_reminder_enabled,
				streak_risk_enabled = excluded.streak_risk_enabled,
				session_ready_enabled = excluded.session_ready_enabled,
				reminder_time = excluded.reminder_time,
				updated_at = now()
		returning updated_at`

	err := r.pool.QueryRow(
		ctx, q, p.UserID, p.DailyReminderEnabled, p.StreakRiskEnabled, p.SessionReadyEnabled, p.ReminderTime,
	).Scan(&p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository: upsert notification preferences: %w", err)
	}

	return nil
}

func (r *pgNotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	const q = `
		insert into notifications (id, user_id, type, channel, delivery_status, title, body, data, provider_message_id, created_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())
		returning created_at`

	err := r.pool.QueryRow(
		ctx, q, n.ID, n.UserID, n.Type, n.Channel, n.DeliveryStatus, n.Title, n.Body, n.Data, n.ProviderMessageID,
	).Scan(&n.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create notification: %w", err)
	}

	return nil
}

func (r *pgNotificationRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*models.Notification, error) {
	const q = `
		select id, user_id, type, channel, delivery_status, title, body, data,
		       provider_message_id, read_at, sent_at, failed_at, created_at
		from notifications
		where id = $1 and user_id = $2`

	var n models.Notification

	err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Channel, &n.DeliveryStatus, &n.Title, &n.Body, &n.Data,
		&n.ProviderMessageID, &n.ReadAt, &n.SentAt, &n.FailedAt, &n.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get notification by id: %w", err)
	}

	return &n, nil
}

func (r *pgNotificationRepository) ListForUser(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]models.Notification, error) {
	const q = `
		select id, user_id, type, channel, delivery_status, title, body, data,
		       provider_message_id, read_at, sent_at, failed_at, created_at
		from notifications
		where user_id = $1
		  and ($2 = false or read_at is null)
		order by created_at desc
		limit $3 offset $4`

	rows, err := r.pool.Query(ctx, q, userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]models.Notification, 0)
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Channel, &n.DeliveryStatus, &n.Title, &n.Body, &n.Data,
			&n.ProviderMessageID, &n.ReadAt, &n.SentAt, &n.FailedAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list notifications: %w", err)
	}

	return notifications, nil
}

func (r *pgNotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	const q = `select count(*) from notifications where user_id = $1 and read_at is null`

	var count int
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("repository: count unread notifications: %w", err)
	}

	return count, nil
}

func (r *pgNotificationRepository) MarkRead(ctx context.Context, id uuid.UUID, userID string) error {
	const q = `
		update notifications
		set read_at = now()
		where id = $1 and user_id = $2 and read_at is null`

	tag, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("repository: mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if _, getErr := r.GetByID(ctx, id, userID); getErr != nil {
			return getErr
		}
	}

	return nil
}

func (r *pgNotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	const q = `
		update notifications
		set read_at = now()
		where user_id = $1 and read_at is null`

	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("repository: mark all notifications read: %w", err)
	}

	return nil
}

func (r *pgNotificationRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status models.NotificationDeliveryStatus, providerMessageID *string) error {
	const q = `
		update notifications
		set delivery_status = $2,
			provider_message_id = coalesce($3, provider_message_id),
			sent_at = case when $2 = 'sent' then now() else sent_at end,
			failed_at = case when $2 = 'failed' then now() else failed_at end
		where id = $1`

	tag, err := r.pool.Exec(ctx, q, id, status, providerMessageID)
	if err != nil {
		return fmt.Errorf("repository: update delivery status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgNotificationRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	if _, err := r.pool.Exec(ctx, `delete from notifications where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete notifications for user: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `delete from notification_preferences where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete notification preferences for user: %w", err)
	}
	return nil
}
