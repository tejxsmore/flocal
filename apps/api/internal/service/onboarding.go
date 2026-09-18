package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"flocal/internal/models"
	"flocal/internal/repository"
)

var ErrNoFocusAreas = errors.New("service: at least one focus area is required")

type OnboardingService struct {
	repo     repository.OnboardingRepository
	authRepo repository.AuthRepository
	auditSvc *AuditService
}

func NewOnboardingService(repo repository.OnboardingRepository, authRepo repository.AuthRepository, auditSvc *AuditService) *OnboardingService {
	return &OnboardingService{repo: repo, authRepo: authRepo, auditSvc: auditSvc}
}

func (s *OnboardingService) logAudit(ctx context.Context, actorUserID *string, action, entityType, entityID string, newValue *models.JSONB) {
	if s.auditSvc == nil {
		return
	}
	source := "api"
	if err := s.auditSvc.Record(ctx, actorUserID, action, entityType, entityID, nil, newValue, &source); err != nil {
		log.Printf("onboarding: audit log failed: %v", err)
	}
}

func toJSONB(v any) *models.JSONB {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	jsonb := models.JSONB(data)
	return &jsonb
}

type CompleteProfileInput struct {
	Name        string
	Username    *string
	Image       *string
	CountryCode *string
}

func (s *OnboardingService) CompleteProfile(ctx context.Context, userID string, input CompleteProfileInput) (*models.User, error) {
	username := normalizeUsername(input.Username)
	countryCode := normalizeCountryCode(input.CountryCode)

	if err := s.repo.CompleteProfile(ctx, userID, input.Name, username, input.Image, countryCode); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("service: complete profile: %w", err)
	}

	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: reload user after profile onboarding: %w", err)
	}

	s.logAudit(ctx, &userID, "onboarding.profile_completed", "user", userID, nil)

	return user, nil
}

type CompletePreferencesInput struct {
	PrimaryGoal         *models.OnboardingGoal
	DailyTimeCommitment *models.DailyTimeCommitment
	FocusAreas          []models.FocusArea
}

func (s *OnboardingService) CompletePreferences(ctx context.Context, userID string, input CompletePreferencesInput) (*models.UserPreferences, []models.UserFocusArea, error) {
	if input.PrimaryGoal != nil && !input.PrimaryGoal.Valid() {
		return nil, nil, errors.New("service: invalid primaryGoal")
	}
	if input.DailyTimeCommitment != nil && !input.DailyTimeCommitment.Valid() {
		return nil, nil, errors.New("service: invalid dailyTimeCommitment")
	}

	seen := make(map[models.FocusArea]bool, len(input.FocusAreas))
	for _, area := range input.FocusAreas {
		if !area.Valid() {
			return nil, nil, fmt.Errorf("service: invalid focusArea %q", area)
		}
		seen[area] = true
	}
	dedup := make([]models.FocusArea, 0, len(seen))
	for area := range seen {
		dedup = append(dedup, area)
	}

	now := time.Now()
	prefs := &models.UserPreferences{
		UserID:              userID,
		PrimaryGoal:         input.PrimaryGoal,
		DailyTimeCommitment: input.DailyTimeCommitment,
		CompletedAt:         &now,
	}

	if err := s.repo.UpsertPreferences(ctx, prefs); err != nil {
		return nil, nil, fmt.Errorf("service: save preferences: %w", err)
	}

	if err := s.repo.ReplaceFocusAreas(ctx, userID, dedup); err != nil {
		return nil, nil, fmt.Errorf("service: save focus areas: %w", err)
	}

	areas, err := s.repo.ListFocusAreas(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: reload focus areas: %w", err)
	}

	s.logAudit(ctx, &userID, "onboarding.preferences_completed", "user", userID,
		toJSONB(map[string]any{
			"primaryGoal":         input.PrimaryGoal,
			"dailyTimeCommitment": input.DailyTimeCommitment,
		}),
	)

	return prefs, areas, nil
}

func (s *OnboardingService) GetPreferences(ctx context.Context, userID string) (*models.UserPreferences, []models.UserFocusArea, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	areas, err := s.repo.ListFocusAreas(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("service: list focus areas: %w", err)
	}

	return prefs, areas, nil
}
