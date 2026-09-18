package service

import (
	"context"
	"errors"
	"fmt"

	"flocal/internal/models"
	"flocal/internal/repository"
)

var ErrPlanNotFound = errors.New("service: plan not found")

type PlanService struct {
	viewRepo repository.ViewRepository
}

func NewPlanService(viewRepo repository.ViewRepository) *PlanService {
	return &PlanService{viewRepo: viewRepo}
}

func (s *PlanService) GetCurrentPlan(ctx context.Context, userID string) (*models.UserCurrentPlan, error) {
	plan, err := s.viewRepo.GetUserCurrentPlan(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("service: get current plan: %w", err)
	}
	return plan, nil
}
