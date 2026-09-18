package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
)

var (
	ErrCategoryNotFound  = errors.New("service: category not found")
	ErrNoTopicsAvailable = errors.New("service: no topics available")
)

type TopicService struct {
	repo repository.TopicRepository
}

func NewTopicService(repo repository.TopicRepository) *TopicService {
	return &TopicService{repo: repo}
}

func (s *TopicService) SpinRandomTopic(ctx context.Context, userID string, categoryID *uuid.UUID, format *models.TopicFormat) (*models.Topic, error) {
	if categoryID != nil {
		cat, err := s.repo.GetCategoryByID(ctx, *categoryID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrCategoryNotFound
			}
			return nil, fmt.Errorf("service: get category: %w", err)
		}
		if !cat.IsActive {
			return nil, ErrCategoryNotFound
		}
	}

	topic, err := s.repo.RandomActive(ctx, categoryID, format)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNoTopicsAvailable
		}
		return nil, fmt.Errorf("service: random topic: %w", err)
	}

	spin := &models.TopicSpin{
		ID:               uuid.New(),
		UserID:           userID,
		TopicID:          topic.ID,
		CategoryFilterID: categoryID,
		FormatFilter:     format,
	}
	if err := s.repo.RecordSpin(ctx, spin); err != nil {
		return nil, fmt.Errorf("service: record spin: %w", err)
	}

	return topic, nil
}

func (s *TopicService) ListCategories(ctx context.Context) ([]models.TopicCategory, error) {
	categories, err := s.repo.ListActiveCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list categories: %w", err)
	}
	return categories, nil
}
