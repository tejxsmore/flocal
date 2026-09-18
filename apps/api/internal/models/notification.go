package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type NotificationPreferences struct {
	UserID               string     `db:"user_id" json:"userId"`
	DailyReminderEnabled bool       `db:"daily_reminder_enabled" json:"dailyReminderEnabled"`
	StreakRiskEnabled    bool       `db:"streak_risk_enabled" json:"streakRiskEnabled"`
	SessionReadyEnabled  bool       `db:"session_ready_enabled" json:"sessionReadyEnabled"`
	ReminderTime         *TimeOfDay `db:"reminder_time" json:"reminderTime,omitempty"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updatedAt"`
}

func (NotificationPreferences) TableName() string {
	return "notification_preferences"
}

func (p NotificationPreferences) Validate() error {
	var errs []error

	if p.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	return errors.Join(errs...)
}

type Notification struct {
	ID                uuid.UUID                  `db:"id" json:"id"`
	UserID            string                     `db:"user_id" json:"userId"`
	Type              NotificationType           `db:"type" json:"type"`
	Channel           NotificationChannel        `db:"channel" json:"channel"`
	DeliveryStatus    NotificationDeliveryStatus `db:"delivery_status" json:"deliveryStatus"`
	Title             string                     `db:"title" json:"title"`
	Body              *string                    `db:"body" json:"body,omitempty"`
	Data              JSONB                      `db:"data" json:"data"`
	ProviderMessageID *string                    `db:"provider_message_id" json:"providerMessageId,omitempty"`
	ReadAt            *time.Time                 `db:"read_at" json:"readAt,omitempty"`
	SentAt            *time.Time                 `db:"sent_at" json:"sentAt,omitempty"`
	FailedAt          *time.Time                 `db:"failed_at" json:"failedAt,omitempty"`
	CreatedAt         time.Time                  `db:"created_at" json:"createdAt"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (n Notification) Read() bool {
	return n.ReadAt != nil
}

func (n Notification) Sent() bool {
	return n.SentAt != nil
}

func (n Notification) Failed() bool {
	return n.FailedAt != nil
}

func (n *Notification) ApplyDefaults() {
	if n.Data.IsNull() {
		n.Data = JSONB("{}")
	}
}

func (n Notification) Validate() error {
	var errs []error

	if n.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if !n.Type.Valid() {
		errs = append(errs, errors.New("invalid type"))
	}

	if !n.Channel.Valid() {
		errs = append(errs, errors.New("invalid channel"))
	}

	if !n.DeliveryStatus.Valid() {
		errs = append(errs, errors.New("invalid deliveryStatus"))
	}

	if n.Title == "" {
		errs = append(errs, errors.New("title is required"))
	}

	if !n.Data.IsNull() && !n.Data.IsObject() {
		errs = append(errs, errors.New("data must be a JSON object"))
	}

	return errors.Join(errs...)
}
