package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type DailyChallenge struct {
	ID               uuid.UUID       `db:"id" json:"id"`
	TopicID          uuid.UUID       `db:"topic_id" json:"topicId"`
	ChallengeDate    DateOnly        `db:"challenge_date" json:"challengeDate"`
	LanguageCode     string          `db:"language_code" json:"languageCode"`
	Difficulty       TopicDifficulty `db:"difficulty" json:"difficulty"`
	SpeakTimeSeconds int             `db:"speak_time_seconds" json:"speakTimeSeconds"`
	XPReward         int             `db:"xp_reward" json:"xpReward"`
	CreatedAt        time.Time       `db:"created_at" json:"createdAt"`
}

func (DailyChallenge) TableName() string {
	return "daily_challenges"
}

func (c DailyChallenge) Validate() error {
	var errs []error

	if c.ID == uuid.Nil {
		errs = append(errs, errors.New("id is required"))
	}

	if c.TopicID == uuid.Nil {
		errs = append(errs, errors.New("topicId is required"))
	}

	if c.ChallengeDate.IsZero() {
		errs = append(errs, errors.New("challengeDate is required"))
	}

	if c.LanguageCode == "" {
		errs = append(errs, errors.New("languageCode is required"))
	}

	if !c.Difficulty.Valid() {
		errs = append(errs, errors.New("invalid difficulty"))
	}

	if c.SpeakTimeSeconds <= 0 {
		errs = append(
			errs,
			errors.New("speakTimeSeconds must be > 0"),
		)
	}

	if c.XPReward < 0 {
		errs = append(errs, errors.New("xpReward must be >= 0"))
	}

	return errors.Join(errs...)
}

type DailyChallengeCompletion struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	UserID           string     `db:"user_id" json:"userId"`
	DailyChallengeID uuid.UUID  `db:"daily_challenge_id" json:"dailyChallengeId"`
	SessionID        *uuid.UUID `db:"session_id" json:"sessionId,omitempty"`
	CompletedAt      time.Time  `db:"completed_at" json:"completedAt"`
}

func (DailyChallengeCompletion) TableName() string {
	return "daily_challenge_completions"
}

func (c DailyChallengeCompletion) Validate() error {
	var errs []error

	if c.ID == uuid.Nil {
		errs = append(errs, errors.New("id is required"))
	}

	if c.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if c.DailyChallengeID == uuid.Nil {
		errs = append(
			errs,
			errors.New("dailyChallengeId is required"),
		)
	}

	if c.CompletedAt.IsZero() {
		errs = append(
			errs,
			errors.New("completedAt is required"),
		)
	}

	return errors.Join(errs...)
}
