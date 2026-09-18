package models

import (
	"errors"
	"time"
)

type UserPreferences struct {
	UserID              string               `db:"user_id" json:"userId"`
	PrimaryGoal         *OnboardingGoal      `db:"primary_goal" json:"primaryGoal,omitempty"`
	DailyTimeCommitment *DailyTimeCommitment `db:"daily_time_commitment" json:"dailyTimeCommitment,omitempty"`
	CompletedAt         *time.Time           `db:"completed_at" json:"completedAt,omitempty"`
	CreatedAt           time.Time            `db:"created_at" json:"createdAt"`
	UpdatedAt           time.Time            `db:"updated_at" json:"updatedAt"`
}

func (UserPreferences) TableName() string {
	return "user_preferences"
}

func (p UserPreferences) Completed() bool {
	return p.CompletedAt != nil
}

func (p UserPreferences) Validate() error {
	var errs []error

	if p.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if p.PrimaryGoal != nil && !p.PrimaryGoal.Valid() {
		errs = append(errs, errors.New("invalid primaryGoal"))
	}

	if p.DailyTimeCommitment != nil && !p.DailyTimeCommitment.Valid() {
		errs = append(errs, errors.New("invalid dailyTimeCommitment"))
	}

	return errors.Join(errs...)
}

type UserFocusArea struct {
	UserID    string    `db:"user_id" json:"userId"`
	FocusArea FocusArea `db:"focus_area" json:"focusArea"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (UserFocusArea) TableName() string {
	return "user_focus_areas"
}

func (f UserFocusArea) Validate() error {
	var errs []error

	if f.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if !f.FocusArea.Valid() {
		errs = append(errs, errors.New("invalid focusArea"))
	}

	return errors.Join(errs...)
}
