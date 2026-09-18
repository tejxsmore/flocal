package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ActorUserID *string   `db:"actor_user_id" json:"actorUserId,omitempty"`
	Action      string    `db:"action" json:"action"`
	EntityType  string    `db:"entity_type" json:"entityType"`
	EntityID    string    `db:"entity_id" json:"entityId"`
	OldValue    *JSONB    `db:"old_value" json:"oldValue,omitempty"`
	NewValue    *JSONB    `db:"new_value" json:"newValue,omitempty"`
	Source      *string   `db:"source" json:"source,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

func (AuditLog) TableName() string {
	return "audit_log"
}

func (a AuditLog) Validate() error {
	var errs []error

	if a.Action == "" {
		errs = append(errs, errors.New("action is required"))
	}

	if a.EntityType == "" {
		errs = append(errs, errors.New("entityType is required"))
	}

	if a.EntityID == "" {
		errs = append(errs, errors.New("entityId is required"))
	}

	return errors.Join(errs...)
}
