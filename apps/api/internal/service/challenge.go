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

var (
	ErrDailyChallengeNotFound = errors.New("service: daily challenge not found")
	ErrChallengeAlreadyDone   = errors.New("service: daily challenge already completed")
)

const defaultCompletionsLimit = 30

type ChallengeService struct {
	repo repository.ChallengeRepository
}

func NewChallengeService(repo repository.ChallengeRepository) *ChallengeService {
	return &ChallengeService{repo: repo}
}

func (s *ChallengeService) GetToday(ctx context.Context, languageCode string, difficulty models.TopicDifficulty) (*models.DailyChallenge, error) {
	if languageCode == "" {
		languageCode = "en"
	}
	if !difficulty.Valid() {
		difficulty = models.DifficultyMedium
	}

	today := models.NewDateOnly(time.Now())

	challenge, err := s.repo.GetForDate(ctx, today, languageCode, difficulty)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDailyChallengeNotFound
		}
		return nil, fmt.Errorf("service: get today's challenge: %w", err)
	}

	return challenge, nil
}

func (s *ChallengeService) CompleteChallenge(ctx context.Context, userID string, challengeID uuid.UUID, sessionID *uuid.UUID) (*models.DailyChallengeCompletion, bool, error) {
	if _, err := s.repo.GetByID(ctx, challengeID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, false, ErrDailyChallengeNotFound
		}
		return nil, false, fmt.Errorf("service: complete challenge: get challenge: %w", err)
	}

	completion := &models.DailyChallengeCompletion{
		ID:               uuid.New(),
		UserID:           userID,
		DailyChallengeID: challengeID,
		SessionID:        sessionID,
	}

	created, err := s.repo.CreateCompletion(ctx, completion)
	if err != nil {
		return nil, false, fmt.Errorf("service: complete challenge: %w", err)
	}

	return completion, created, nil
}

func (s *ChallengeService) ListMyCompletions(ctx context.Context, userID string, limit, offset int) ([]models.DailyChallengeCompletion, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultCompletionsLimit
	}
	if offset < 0 {
		offset = 0
	}

	completions, err := s.repo.ListCompletionsForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: list my completions: %w", err)
	}

	return completions, nil
}
