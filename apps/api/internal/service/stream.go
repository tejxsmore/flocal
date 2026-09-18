package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
	"flocal/internal/utils"
)

type SessionStreamService struct {
	sessionRepo    repository.SessionRepository
	transcriptRepo repository.TranscriptRepository
	s3Client       *s3.Client
	s3Bucket       string
	s3UploadPrefix string
}

func NewSessionStreamService(
	sessionRepo repository.SessionRepository,
	transcriptRepo repository.TranscriptRepository,
	s3Client *s3.Client,
	s3Bucket string,
	s3UploadPrefix string,
) *SessionStreamService {
	return &SessionStreamService{
		sessionRepo:    sessionRepo,
		transcriptRepo: transcriptRepo,
		s3Client:       s3Client,
		s3Bucket:       s3Bucket,
		s3UploadPrefix: s3UploadPrefix,
	}
}

func (s *SessionStreamService) StartStream(
	ctx context.Context,
	session *models.SpeakingSession,
) error {
	if session.Status != models.SessionStatusPending {
		return fmt.Errorf(
			"service: session %s is not pending",
			session.ID,
		)
	}

	if err := s.sessionRepo.MarkStarted(ctx, session.ID); err != nil {
		return fmt.Errorf("service: mark session started: %w", err)
	}

	return nil
}

type FinalizedTranscript struct {
	RawText       string
	WordCount     int
	AvgConfidence float64
	Words         []DeepgramWord
	Language      string
}

func (s *SessionStreamService) FinalizeStream(
	ctx context.Context,
	session *models.SpeakingSession,
	audio []byte,
	ft FinalizedTranscript,
) error {
	key := fmt.Sprintf(
		"%s%s.webm",
		s.s3UploadPrefix,
		session.ID,
	)

	if err := utils.UploadAudio(
		ctx,
		s.s3Client,
		s.s3Bucket,
		key,
		audio,
		"audio/webm",
	); err != nil {
		return fmt.Errorf("service: upload audio: %w", err)
	}

	wordTimings, err := models.NewJSONB(ft.Words)
	if err != nil {
		return fmt.Errorf(
			"service: marshal word timings: %w",
			err,
		)
	}

	avgConfidence := ft.AvgConfidence

	language := ft.Language
	if language == "" {
		language = "en"
	}

	transcript := &models.Transcript{
		ID:            uuid.New(),
		SessionID:     session.ID,
		RawText:       ft.RawText,
		WordCount:     ft.WordCount,
		Language:      &language,
		AvgConfidence: &avgConfidence,
		WordTimings:   &wordTimings,
	}

	if err := transcript.Validate(); err != nil {
		return fmt.Errorf("service: validate transcript: %w", err)
	}

	if err := s.transcriptRepo.Create(ctx, transcript); err != nil {
		return fmt.Errorf(
			"service: create transcript: %w",
			err,
		)
	}

	durationSeconds := float64(session.SpeakTimeSeconds)

	if len(ft.Words) > 0 {
		durationSeconds = ft.Words[len(ft.Words)-1].End

		if ft.Words[0].Start > 0 {
			durationSeconds -= ft.Words[0].Start
		}
	}

	if durationSeconds <= 0 {
		durationSeconds = float64(session.SpeakTimeSeconds)
	}

	if err := s.sessionRepo.MarkSubmitted(
		ctx,
		session.ID,
		key,
		durationSeconds,
		"webm",
		int64(len(audio)),
	); err != nil {
		return fmt.Errorf(
			"service: mark session submitted: %w",
			err,
		)
	}

	return nil
}

func (s *SessionStreamService) FinalizePracticeStream(
	ctx context.Context,
	session *models.SpeakingSession,
	audio []byte,
) error {
	key := fmt.Sprintf(
		"%s%s.webm",
		s.s3UploadPrefix,
		session.ID,
	)

	if err := utils.UploadAudio(
		ctx,
		s.s3Client,
		s.s3Bucket,
		key,
		audio,
		"audio/webm",
	); err != nil {
		return fmt.Errorf("service: upload audio: %w", err)
	}

	if err := s.sessionRepo.MarkSubmitted(
		ctx,
		session.ID,
		key,
		float64(session.SpeakTimeSeconds),
		"webm",
		int64(len(audio)),
	); err != nil {
		return fmt.Errorf("service: mark session submitted: %w", err)
	}

	if err := s.sessionRepo.MarkCompleted(ctx, session.ID); err != nil {
		return fmt.Errorf("service: mark session completed: %w", err)
	}

	return nil
}
