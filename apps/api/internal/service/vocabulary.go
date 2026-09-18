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
	ErrVocabularyWordNotFound = errors.New("service: vocabulary word not found")
	ErrUserVocabularyNotFound = errors.New("service: user vocabulary not found")
)

const (
	defaultVocabularyLimit = 30
	maxVocabularyLimit     = 100
)

type VocabularyService struct {
	repo repository.VocabularyRepository
}

func NewVocabularyService(repo repository.VocabularyRepository) *VocabularyService {
	return &VocabularyService{repo: repo}
}

func (s *VocabularyService) RecordEncounter(
	ctx context.Context,
	userID string,
	word string,
	definition *string,
	difficulty *models.TopicDifficulty,
	suggested bool,
) (*models.UserVocabulary, error) {
	w, err := s.repo.GetOrCreateWord(ctx, word, definition, difficulty)
	if err != nil {
		return nil, fmt.Errorf("service: record encounter: get or create word: %w", err)
	}

	v, err := s.repo.RecordEncounter(ctx, userID, w.ID, suggested)
	if err != nil {
		return nil, fmt.Errorf("service: record encounter: %w", err)
	}

	return v, nil
}

func (s *VocabularyService) ListMyVocabulary(ctx context.Context, userID string, mastered *bool, limit, offset int) ([]repository.VocabularyEntry, int, error) {
	limit, offset = normalizeVocabularyPaging(limit, offset)

	entries, err := s.repo.ListUserVocabularyWithWords(ctx, userID, mastered, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service: list my vocabulary: %w", err)
	}

	total, err := s.repo.CountUserVocabulary(ctx, userID, mastered)
	if err != nil {
		return nil, 0, fmt.Errorf("service: count my vocabulary: %w", err)
	}

	return entries, total, nil
}

func (s *VocabularyService) SetMastered(ctx context.Context, userID string, wordID uuid.UUID, mastered bool) (*models.UserVocabulary, error) {
	v, err := s.repo.SetMastered(ctx, userID, wordID, mastered)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserVocabularyNotFound
		}
		return nil, fmt.Errorf("service: set mastered: %w", err)
	}

	return v, nil
}

func normalizeVocabularyPaging(limit, offset int) (int, int) {
	if limit <= 0 || limit > maxVocabularyLimit {
		limit = defaultVocabularyLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
