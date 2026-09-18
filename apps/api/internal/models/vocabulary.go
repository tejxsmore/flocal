package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type VocabularyWord struct {
	ID         uuid.UUID        `db:"id" json:"id"`
	Word       string           `db:"word" json:"word"`
	Definition *string          `db:"definition" json:"definition,omitempty"`
	Difficulty *TopicDifficulty `db:"difficulty" json:"difficulty,omitempty"`
	CreatedAt  time.Time        `db:"created_at" json:"createdAt"`
}

func (VocabularyWord) TableName() string {
	return "vocabulary_words"
}

func (w VocabularyWord) Validate() error {
	var errs []error

	if w.Word == "" {
		errs = append(errs, errors.New("word is required"))
	}

	if w.Difficulty != nil && !w.Difficulty.Valid() {
		errs = append(errs, errors.New("invalid difficulty"))
	}

	return errors.Join(errs...)
}

type UserVocabulary struct {
	ID             uuid.UUID `db:"id" json:"id"`
	UserID         string    `db:"user_id" json:"userId"`
	WordID         uuid.UUID `db:"word_id" json:"wordId"`
	TimesSeen      int       `db:"times_seen" json:"timesSeen"`
	TimesSuggested int       `db:"times_suggested" json:"timesSuggested"`
	Mastered       bool      `db:"mastered" json:"mastered"`
	FirstSeenAt    time.Time `db:"first_seen_at" json:"firstSeenAt"`
	LastSeenAt     time.Time `db:"last_seen_at" json:"lastSeenAt"`
}

func (UserVocabulary) TableName() string {
	return "user_vocabulary"
}

func (v UserVocabulary) Validate() error {
	var errs []error

	if v.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if v.WordID == uuid.Nil {
		errs = append(errs, errors.New("wordId is required"))
	}

	if v.TimesSeen < 0 {
		errs = append(errs, errors.New("timesSeen must be >= 0"))
	}

	if v.TimesSuggested < 0 {
		errs = append(errs, errors.New("timesSuggested must be >= 0"))
	}

	if !v.FirstSeenAt.IsZero() &&
		!v.LastSeenAt.IsZero() &&
		v.LastSeenAt.Before(v.FirstSeenAt) {
		errs = append(
			errs,
			errors.New("lastSeenAt must be after or equal to firstSeenAt"),
		)
	}

	return errors.Join(errs...)
}
