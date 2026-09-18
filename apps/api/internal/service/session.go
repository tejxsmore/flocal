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
	ErrTopicNotFound             = errors.New("service: topic not found")
	ErrSessionNotFound           = errors.New("service: session not found")
	ErrSessionAccessDenied       = errors.New("service: session access denied")
	ErrInvalidPrepTime           = errors.New("service: prepTimeSeconds must be one of 0, 300, 600, 900")
	ErrDebateStanceNotApplicable = errors.New("service: debateStance can only be set for debate topics")
	ErrInvalidFilter             = errors.New("service: invalid filter parameter")
)

var validSessionStatuses = map[string]bool{
	"pending":    true,
	"processing": true,
	"completed":  true,
	"failed":     true,
}

var validTopicFormats = map[string]bool{
	"word":          true,
	"quote":         true,
	"debate":        true,
	"situation":     true,
	"story_starter": true,
	"image":         true,
}

var validSortOptions = map[string]bool{
	"newest":     true,
	"oldest":     true,
	"score_desc": true,
	"score_asc":  true,
}

var validPrepTimeSeconds = map[int]bool{
	0:   true,
	300: true,
	600: true,
	900: true,
}

const fixedSpeakTimeSeconds = 60

type SessionService struct {
	sessionRepo repository.SessionRepository
	topicRepo   repository.TopicRepository
}

type SessionHistoryFilter struct {
	Status []string
	Format []string
	From   *time.Time
	To     *time.Time
	Search string
	Sort   string
}

func NewSessionService(
	sessionRepo repository.SessionRepository,
	topicRepo repository.TopicRepository,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		topicRepo:   topicRepo,
	}
}

func (s *SessionService) CreateSession(
	ctx context.Context,
	user *models.User,
	topicID uuid.UUID,
	prepTimeSeconds *int,
	debateStance *string,
) (*models.SpeakingSession, error) {
	topic, err := s.topicRepo.GetByID(ctx, topicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf("service: get topic: %w", err)
	}

	if !topic.Selectable() {
		return nil, ErrTopicNotFound
	}

	if debateStance != nil && topic.Format != models.TopicFormat("debate") {
		return nil, ErrDebateStanceNotApplicable
	}

	prep := 0

	if prepTimeSeconds != nil {
		prep = *prepTimeSeconds
	} else if validPrepTimeSeconds[topic.RecommendedPrepSeconds] {
		prep = topic.RecommendedPrepSeconds
	} else if validPrepTimeSeconds[user.DefaultPrepTimeSeconds] {
		prep = user.DefaultPrepTimeSeconds
	}

	if !validPrepTimeSeconds[prep] {
		return nil, ErrInvalidPrepTime
	}

	session := &models.SpeakingSession{
		ID:               uuid.New(),
		UserID:           user.ID,
		TopicID:          topic.ID,
		PrepTimeSeconds:  prep,
		SpeakTimeSeconds: fixedSpeakTimeSeconds,
		Status:           models.SessionStatusPending,
	}

	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("service: validate session: %w", err)
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("service: create session: %w", err)
	}

	return session, nil
}

func (s *SessionService) GetSession(
	ctx context.Context,
	id uuid.UUID,
	userID string,
) (*models.SpeakingSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf("service: get session: %w", err)
	}

	if session.UserID != userID {
		return nil, ErrSessionAccessDenied
	}

	return session, nil
}

func (s *SessionService) ListSessions(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]models.SpeakingSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	sessions, err := s.sessionRepo.ListForUser(
		ctx,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("service: list sessions: %w", err)
	}

	return sessions, nil
}

func validateHistoryFilter(filter models.SessionHistoryFilter) error {
	for _, s := range filter.Status {
		if !validSessionStatuses[s] {
			return ErrInvalidFilter
		}
	}

	for _, f := range filter.Format {
		if !validTopicFormats[f] {
			return ErrInvalidFilter
		}
	}

	if filter.Sort != "" && !validSortOptions[filter.Sort] {
		return ErrInvalidFilter
	}

	return nil
}

func (s *SessionService) ListMyHistory(
	ctx context.Context,
	userID string,
	filter models.SessionHistoryFilter,
	limit int,
	offset int,
) ([]models.SessionHistoryItem, error) {
	if err := validateHistoryFilter(filter); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	items, err := s.sessionRepo.ListHistoryForUser(ctx, userID, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: list my history: %w", err)
	}

	return items, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, id uuid.UUID, userID string) error {
	if err := s.sessionRepo.DeleteByID(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}

		return fmt.Errorf("service: delete session: %w", err)
	}

	return nil
}

func (s *SessionService) MarkSessionFailed(
	ctx context.Context,
	id uuid.UUID,
	reason string,
) error {
	if err := s.sessionRepo.MarkFailed(ctx, id, reason); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}

		return fmt.Errorf("service: mark session failed: %w", err)
	}

	return nil
}
