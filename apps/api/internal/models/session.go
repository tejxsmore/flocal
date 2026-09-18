package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type SpeakingSession struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	UserID           string     `db:"user_id" json:"userId"`
	TopicID          uuid.UUID  `db:"topic_id" json:"topicId"`
	DailyChallengeID *uuid.UUID `db:"daily_challenge_id" json:"dailyChallengeId,omitempty"`

	ClientUploadToken *string `db:"client_upload_token" json:"-"`

	DebateStance *string `db:"debate_stance" json:"debateStance,omitempty"`

	PrepTimeSeconds  int `db:"prep_time_seconds" json:"prepTimeSeconds"`
	SpeakTimeSeconds int `db:"speak_time_seconds" json:"speakTimeSeconds"`

	Status              SessionStatus `db:"status" json:"status"`
	ProcessingStep      *string       `db:"processing_step" json:"processingStep,omitempty"`
	ProcessingAttemptID *uuid.UUID    `db:"processing_attempt_id" json:"-"`
	ProcessingStartedAt *time.Time    `db:"processing_started_at" json:"processingStartedAt,omitempty"`
	ProcessingWorkerID  *string       `db:"processing_worker_id" json:"-"`
	FailureReason       *string       `db:"failure_reason" json:"failureReason,omitempty"`
	RetryCount          int           `db:"retry_count" json:"retryCount"`
	LastError           *string       `db:"last_error" json:"lastError,omitempty"`
	LastRetryAt         *time.Time    `db:"last_retry_at" json:"lastRetryAt,omitempty"`
	NextRetryAt         *time.Time    `db:"next_retry_at" json:"nextRetryAt,omitempty"`

	AudioS3Key           *string    `db:"audio_s3_key" json:"-"`
	AudioDurationSeconds *float64   `db:"audio_duration_seconds" json:"audioDurationSeconds,omitempty"`
	AudioFormat          *string    `db:"audio_format" json:"audioFormat,omitempty"`
	AudioSizeBytes       *int64     `db:"audio_size_bytes" json:"audioSizeBytes,omitempty"`
	AudioExpiresAt       *time.Time `db:"audio_expires_at" json:"audioExpiresAt,omitempty"`
	AudioDeletedAt       *time.Time `db:"audio_deleted_at" json:"audioDeletedAt,omitempty"`

	ShareTokenHash *string    `db:"share_token_hash" json:"-"`
	ShareEnabled   bool       `db:"share_enabled" json:"shareEnabled"`
	ShareExpiresAt *time.Time `db:"share_expires_at" json:"shareExpiresAt,omitempty"`

	StartedAt   *time.Time `db:"started_at" json:"startedAt,omitempty"`
	SubmittedAt *time.Time `db:"submitted_at" json:"submittedAt,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completedAt,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type SessionHistoryItem struct {
	ID               uuid.UUID     `json:"id"`
	TopicID          uuid.UUID     `json:"topicId"`
	TopicTitle       string        `json:"topicTitle"`
	TopicFormat      TopicFormat   `json:"topicFormat"`
	Status           SessionStatus `json:"status"`
	PrepTimeSeconds  int           `json:"prepTimeSeconds"`
	SpeakTimeSeconds int           `json:"speakTimeSeconds"`
	OverallScore     *float64      `json:"overallScore"`
	WordsPerMinute   *float64      `json:"wordsPerMinute"`
	CreatedAt        time.Time     `json:"createdAt"`
	CompletedAt      *time.Time    `json:"completedAt"`
}

type SessionHistoryFilter struct {
	Status []string
	Format []string
	From   *time.Time
	To     *time.Time
	Search string
	Sort   string
}

func (SpeakingSession) TableName() string {
	return "speaking_sessions"
}

func (s SpeakingSession) AudioAvailable() bool {
	if s.AudioS3Key == nil || s.AudioDeletedAt != nil {
		return false
	}

	if s.AudioExpiresAt != nil && time.Now().After(*s.AudioExpiresAt) {
		return false
	}

	return true
}

func (s SpeakingSession) Shareable() bool {
	if !s.ShareEnabled || s.ShareTokenHash == nil {
		return false
	}

	if s.ShareExpiresAt != nil && time.Now().After(*s.ShareExpiresAt) {
		return false
	}

	return true
}

func (s SpeakingSession) Pending() bool {
	return s.Status == SessionStatusPending
}

func (s SpeakingSession) Processing() bool {
	return s.Status == SessionStatusProcessing
}

func (s SpeakingSession) Completed() bool {
	return s.Status == SessionStatusCompleted
}

func (s SpeakingSession) Failed() bool {
	return s.Status == SessionStatusFailed
}

func (s SpeakingSession) Terminal() bool {
	return s.Status.Terminal()
}

func (s SpeakingSession) Retryable() bool {
	return s.Status == SessionStatusFailed && s.NextRetryAt != nil
}

func (s SpeakingSession) HasDebateStance() bool {
	return s.DebateStance != nil
}

func (s SpeakingSession) Validate() error {
	var errs []error

	if s.ID == uuid.Nil {
		errs = append(errs, errors.New("id is required"))
	}

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if s.TopicID == uuid.Nil {
		errs = append(errs, errors.New("topicId is required"))
	}

	if !s.Status.Valid() {
		errs = append(errs, errors.New("invalid status"))
	}

	if s.PrepTimeSeconds < 0 {
		errs = append(
			errs,
			errors.New("prepTimeSeconds must be >= 0"),
		)
	}

	if s.SpeakTimeSeconds <= 0 {
		errs = append(
			errs,
			errors.New("speakTimeSeconds must be > 0"),
		)
	}

	if s.DebateStance != nil && *s.DebateStance != "for" && *s.DebateStance != "against" {
		errs = append(errs, errors.New("debateStance must be 'for' or 'against'"))
	}

	if s.RetryCount < 0 {
		errs = append(
			errs,
			errors.New("retryCount must be >= 0"),
		)
	}

	if s.AudioDurationSeconds != nil && *s.AudioDurationSeconds < 0 {
		errs = append(
			errs,
			errors.New("audioDurationSeconds must be >= 0"),
		)
	}

	if s.AudioSizeBytes != nil && *s.AudioSizeBytes < 0 {
		errs = append(
			errs,
			errors.New("audioSizeBytes must be >= 0"),
		)
	}

	if s.Status == SessionStatusCompleted && s.CompletedAt == nil {
		errs = append(
			errs,
			errors.New("completedAt is required for completed sessions"),
		)
	}

	if s.Status == SessionStatusProcessing && s.ProcessingStartedAt == nil {
		errs = append(
			errs,
			errors.New("processingStartedAt is required for processing sessions"),
		)
	}

	if s.Status == SessionStatusFailed && s.FailureReason == nil {
		errs = append(
			errs,
			errors.New("failureReason is required for failed sessions"),
		)
	}

	if s.StartedAt != nil &&
		s.SubmittedAt != nil &&
		s.StartedAt.After(*s.SubmittedAt) {
		errs = append(
			errs,
			errors.New("startedAt must be before or equal to submittedAt"),
		)
	}

	if s.SubmittedAt != nil &&
		s.CompletedAt != nil &&
		s.SubmittedAt.After(*s.CompletedAt) {
		errs = append(
			errs,
			errors.New("submittedAt must be before or equal to completedAt"),
		)
	}

	if s.AudioDeletedAt != nil &&
		s.AudioExpiresAt != nil &&
		s.AudioDeletedAt.Before(*s.AudioExpiresAt) {
		errs = append(
			errs,
			errors.New("audioDeletedAt must be after or equal to audioExpiresAt"),
		)
	}

	if s.ShareEnabled && s.ShareTokenHash == nil {
		errs = append(
			errs,
			errors.New("shareTokenHash is required when sharing is enabled"),
		)
	}

	return errors.Join(errs...)
}
