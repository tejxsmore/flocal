package service

import (
	"context"
	"fmt"

	"flocal/internal/models"
	"flocal/internal/repository"
)

const (
	auditDefaultLimit = 50
	auditMaxLimit     = 200
)

type AuditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Record(ctx context.Context, actorUserID *string, action, entityType, entityID string, oldValue, newValue *models.JSONB, source *string) error {
	entry := &models.AuditLog{
		ActorUserID: actorUserID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		OldValue:    oldValue,
		NewValue:    newValue,
		Source:      source,
	}
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("service: invalid audit log: %w", err)
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		return fmt.Errorf("service: record audit log: %w", err)
	}
	return nil
}

func (s *AuditService) ListForEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]models.AuditLog, error) {
	logs, err := s.repo.ListByEntity(ctx, entityType, entityID, normalizeAuditLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("service: list audit logs for entity: %w", err)
	}
	return logs, nil
}

func (s *AuditService) ListForActor(ctx context.Context, actorUserID string, limit, offset int) ([]models.AuditLog, error) {
	logs, err := s.repo.ListByActor(ctx, actorUserID, normalizeAuditLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("service: list audit logs for actor: %w", err)
	}
	return logs, nil
}

func (s *AuditService) List(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	logs, err := s.repo.List(ctx, normalizeAuditLimit(limit), offset)
	if err != nil {
		return nil, fmt.Errorf("service: list audit logs: %w", err)
	}
	return logs, nil
}

func normalizeAuditLimit(limit int) int {
	if limit <= 0 {
		return auditDefaultLimit
	}
	if limit > auditMaxLimit {
		return auditMaxLimit
	}
	return limit
}
