package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
)

var ErrNotificationNotFound = errors.New("service: notification not found")

const (
	defaultNotificationLimit = 20
	maxNotificationLimit     = 100
)

type UpdatePreferencesInput struct {
	DailyReminderEnabled bool
	StreakRiskEnabled    bool
	SessionReadyEnabled  bool
	ReminderTime         *models.TimeOfDay
}

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) GetMyPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error) {
	p, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return defaultPreferences(userID), nil
		}
		return nil, fmt.Errorf("service: get my preferences: %w", err)
	}

	return p, nil
}

func (s *NotificationService) UpdateMyPreferences(ctx context.Context, userID string, input UpdatePreferencesInput) (*models.NotificationPreferences, error) {
	p := &models.NotificationPreferences{
		UserID:               userID,
		DailyReminderEnabled: input.DailyReminderEnabled,
		StreakRiskEnabled:    input.StreakRiskEnabled,
		SessionReadyEnabled:  input.SessionReadyEnabled,
		ReminderTime:         input.ReminderTime,
	}

	if err := s.repo.UpsertPreferences(ctx, p); err != nil {
		return nil, fmt.Errorf("service: update my preferences: %w", err)
	}

	return p, nil
}

func (s *NotificationService) ListMyNotifications(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]models.Notification, error) {
	limit, offset = normalizeNotificationPaging(limit, offset)

	notifications, err := s.repo.ListForUser(ctx, userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: list my notifications: %w", err)
	}

	return notifications, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("service: get unread count: %w", err)
	}

	return count, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID string, id uuid.UUID) error {
	if err := s.repo.MarkRead(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("service: mark read: %w", err)
	}

	return nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	if err := s.repo.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("service: mark all read: %w", err)
	}

	return nil
}

func (s *NotificationService) CreateNotification(
	ctx context.Context,
	userID string,
	notificationType models.NotificationType,
	channel models.NotificationChannel,
	title string,
	body *string,
	data models.JSONB,
) (*models.Notification, error) {
	if !s.preferenceAllows(ctx, userID, notificationType) {
		return nil, nil
	}

	if data.IsNull() {
		data = models.JSONB("{}")
	}

	n := &models.Notification{
		ID:             uuid.New(),
		UserID:         userID,
		Type:           notificationType,
		Channel:        channel,
		DeliveryStatus: models.NotificationDeliveryPending,
		Title:          title,
		Body:           body,
		Data:           data,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("service: create notification: %w", err)
	}

	return n, nil
}

func (s *NotificationService) preferenceAllows(ctx context.Context, userID string, notificationType models.NotificationType) bool {
	switch notificationType {
	case models.NotificationTypeDailyReminder, models.NotificationTypeStreakRisk, models.NotificationTypeSessionReady:
	default:
		return true
	}

	prefs, err := s.GetMyPreferences(ctx, userID)
	if err != nil {
		return true
	}

	switch notificationType {
	case models.NotificationTypeDailyReminder:
		return prefs.DailyReminderEnabled
	case models.NotificationTypeStreakRisk:
		return prefs.StreakRiskEnabled
	case models.NotificationTypeSessionReady:
		return prefs.SessionReadyEnabled
	default:
		return true
	}
}

func defaultPreferences(userID string) *models.NotificationPreferences {
	defaultTime := models.NewTimeOfDay(time.Date(0, 1, 1, 19, 0, 0, 0, time.UTC))
	return &models.NotificationPreferences{
		UserID:               userID,
		DailyReminderEnabled: true,
		StreakRiskEnabled:    true,
		SessionReadyEnabled:  true,
		ReminderTime:         &defaultTime,
	}
}

func normalizeNotificationPaging(limit, offset int) (int, int) {
	if limit <= 0 || limit > maxNotificationLimit {
		limit = defaultNotificationLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
