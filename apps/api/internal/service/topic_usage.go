package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
)

type TopicUsageService struct {
	viewRepo repository.ViewRepository
}

func NewTopicUsageService(viewRepo repository.ViewRepository) *TopicUsageService {
	return &TopicUsageService{viewRepo: viewRepo}
}

func (s *TopicUsageService) GetTopicUsage(ctx context.Context, topicID uuid.UUID) (*models.TopicUsageStats, error) {
	stats, err := s.viewRepo.GetTopicUsageStats(ctx, topicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTopicNotFound
		}
		return nil, fmt.Errorf("service: get topic usage: %w", err)
	}
	return stats, nil
}

func (s *TopicUsageService) ListTopTopics(ctx context.Context, limit int) ([]models.TopicUsageStats, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	stats, err := s.viewRepo.ListTopicUsageStats(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("service: list topic usage: %w", err)
	}
	return stats, nil
}
