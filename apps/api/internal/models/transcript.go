package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Transcript struct {
	ID        uuid.UUID `db:"id" json:"id"`
	SessionID uuid.UUID `db:"session_id" json:"sessionId"`

	RawText   string  `db:"raw_text" json:"rawText"`
	WordCount int     `db:"word_count" json:"wordCount"`
	Language  *string `db:"language" json:"language,omitempty"`

	AvgConfidence *float64 `db:"avg_confidence" json:"avgConfidence,omitempty"`
	WordTimings   *JSONB   `db:"word_timings" json:"wordTimings,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (Transcript) TableName() string {
	return "transcripts"
}

func (t Transcript) HasTimings() bool {
	return t.WordTimings != nil && !t.WordTimings.IsNull()
}

func (t Transcript) LanguageOrDefault() string {
	if t.Language == nil || *t.Language == "" {
		return "en"
	}

	return *t.Language
}

func (t Transcript) WordTimingsList() ([]WordTiming, error) {
	if !t.HasTimings() {
		return nil, nil
	}

	var out []WordTiming

	if err := t.WordTimings.Unmarshal(&out); err != nil {
		return nil, err
	}

	return out, nil
}

type WordTiming struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence,omitempty"`
}

func (t Transcript) Validate() error {
	var errs []error

	if t.SessionID == uuid.Nil {
		errs = append(errs, errors.New("sessionId is required"))
	}

	if t.RawText == "" {
		errs = append(errs, errors.New("rawText is required"))
	}

	if t.WordCount < 0 {
		errs = append(errs, errors.New("wordCount must be >= 0"))
	}

	if t.AvgConfidence != nil && (*t.AvgConfidence < 0 || *t.AvgConfidence > 1) {
		errs = append(errs, errors.New("avgConfidence must be between 0 and 1"))
	}

	if t.WordTimings != nil && !t.WordTimings.IsNull() && !t.WordTimings.IsArray() {
		errs = append(errs, errors.New("wordTimings must be a JSON array"))
	}

	return errors.Join(errs...)
}
