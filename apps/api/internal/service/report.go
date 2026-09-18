package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
	"flocal/internal/utils"
)

var ErrSessionReportNotFound = errors.New("service: session report not found")

type SessionReportService struct {
	viewRepo     repository.ViewRepository
	analysisRepo repository.AnalysisRepository
	s3Client     *s3.Client
	s3Bucket     string
}

func NewSessionReportService(
	viewRepo repository.ViewRepository,
	analysisRepo repository.AnalysisRepository,
	s3Client *s3.Client,
	s3Bucket string,
) *SessionReportService {
	return &SessionReportService{
		viewRepo:     viewRepo,
		analysisRepo: analysisRepo,
		s3Client:     s3Client,
		s3Bucket:     s3Bucket,
	}
}

func (s *SessionReportService) GetSessionReport(ctx context.Context, sessionID uuid.UUID, userID string) (*models.SessionReport, error) {
	report, err := s.viewRepo.GetSessionReport(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionReportNotFound
		}
		return nil, fmt.Errorf("service: get session report: %w", err)
	}

	if report.UserID != userID {
		return nil, ErrSessionReportNotFound
	}

	plan, err := s.viewRepo.GetUserCurrentPlan(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("service: get user current plan: %w", err)
	}

	if err := s.attachAudioURL(ctx, report); err != nil {
		return nil, err
	}

	if plan != nil && !plan.CanViewAnalysis {
		stripAnalysisFields(report)
		return report, nil
	}

	if report.Completed() {
		corrections, err := s.analysisRepo.GetGrammarCorrectionsBySessionID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("service: get grammar corrections: %w", err)
		}
		report.GrammarCorrections = corrections

		suggestions, err := s.analysisRepo.GetVocabularySuggestionsBySessionID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("service: get vocabulary suggestions: %w", err)
		}
		report.VocabularySuggestions = suggestions
	}

	return report, nil
}

func (s *SessionReportService) attachAudioURL(ctx context.Context, report *models.SessionReport) error {
	if !report.AudioAvailable() {
		return nil
	}

	url, err := utils.PresignAudioURL(ctx, s.s3Client, s.s3Bucket, *report.AudioS3Key, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("service: presign audio url: %w", err)
	}

	report.AudioPlaybackURL = &url
	return nil
}

func stripAnalysisFields(r *models.SessionReport) {
	r.OverallScore = nil
	r.ClarityScore = nil
	r.DeliveryScore = nil
	r.ContentScore = nil
	r.VocabularyScore = nil
	r.GrammarScore = nil
	r.Strengths = nil
	r.Improvements = nil
	r.CoachMessage = nil
	r.GrammarCorrections = nil
	r.VocabularySuggestions = nil
}
