package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TopicCategory struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Slug      string    `db:"slug" json:"slug"`
	Name      string    `db:"name" json:"name"`
	Icon      *string   `db:"icon" json:"icon,omitempty"`
	IsActive  bool      `db:"is_active" json:"isActive"`
	SortOrder int       `db:"sort_order" json:"sortOrder"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (TopicCategory) TableName() string {
	return "topic_categories"
}

func (c TopicCategory) Active() bool {
	return c.IsActive
}

func (c TopicCategory) Validate() error {
	var errs []error

	if c.Slug == "" {
		errs = append(errs, errors.New("slug is required"))
	}

	if c.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	if c.SortOrder < 0 {
		errs = append(errs, errors.New("sortOrder must be >= 0"))
	}

	return errors.Join(errs...)
}

type Topic struct {
	ID                     uuid.UUID       `db:"id" json:"id"`
	CategoryID             uuid.UUID       `db:"category_id" json:"categoryId"`
	Title                  string          `db:"title" json:"title"`
	Format                 TopicFormat     `db:"format" json:"format"`
	Difficulty             TopicDifficulty `db:"difficulty" json:"difficulty"`
	RecommendedPrepSeconds int             `db:"recommended_prep_seconds" json:"recommendedPrepSeconds"`
	Tags                   StringArray     `db:"tags" json:"tags"`
	Metadata               JSONB           `db:"metadata" json:"metadata"`
	IsActive               bool            `db:"is_active" json:"isActive"`
	IsPremium              bool            `db:"is_premium" json:"isPremium"`
	Source                 TopicSource     `db:"source" json:"source"`
	LanguageCode           string          `db:"language_code" json:"languageCode"`
	SubmittedBy            *string         `db:"submitted_by" json:"submittedBy,omitempty"`
	DeletedAt              *time.Time      `db:"deleted_at" json:"deletedAt,omitempty"`
	CreatedAt              time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time       `db:"updated_at" json:"updatedAt"`
}

func (Topic) TableName() string {
	return "topics"
}

func (t Topic) Selectable() bool {
	return t.IsActive && t.DeletedAt == nil
}

func (t Topic) Deleted() bool {
	return t.DeletedAt != nil
}

func (t Topic) Premium() bool {
	return t.IsPremium
}

func (t *Topic) ApplyDefaults() {
	if t.Tags == nil {
		t.Tags = StringArray{}
	}

	if t.Metadata.IsNull() {
		t.Metadata = JSONB("{}")
	}
}

func (t Topic) Validate() error {
	var errs []error

	if t.ID == uuid.Nil {
		errs = append(errs, errors.New("id is required"))
	}

	if t.CategoryID == uuid.Nil {
		errs = append(errs, errors.New("categoryId is required"))
	}

	if t.Title == "" {
		errs = append(errs, errors.New("title is required"))
	}

	if !t.Format.Valid() {
		errs = append(errs, errors.New("invalid format"))
	}

	if !t.Difficulty.Valid() {
		errs = append(errs, errors.New("invalid difficulty"))
	}

	if !t.Source.Valid() {
		errs = append(errs, errors.New("invalid source"))
	}

	if t.RecommendedPrepSeconds < 0 {
		errs = append(
			errs,
			errors.New("recommendedPrepSeconds must be >= 0"),
		)
	}

	if t.LanguageCode == "" {
		errs = append(errs, errors.New("languageCode is required"))
	}

	if !t.Metadata.IsNull() && !t.Metadata.IsObject() {
		errs = append(errs, errors.New("metadata must be a JSON object"))
	}

	return errors.Join(errs...)
}

type TopicSubmission struct {
	ID          uuid.UUID             `db:"id" json:"id"`
	SubmittedBy string                `db:"submitted_by" json:"submittedBy"`
	CategoryID  *uuid.UUID            `db:"category_id" json:"categoryId,omitempty"`
	Title       string                `db:"title" json:"title"`
	Format      TopicFormat           `db:"format" json:"format"`
	Difficulty  TopicDifficulty       `db:"difficulty" json:"difficulty"`
	Tags        StringArray           `db:"tags" json:"tags"`
	Status      TopicSubmissionStatus `db:"status" json:"status"`
	ReviewNote  *string               `db:"review_note" json:"reviewNote,omitempty"`
	ReviewedBy  *string               `db:"reviewed_by" json:"reviewedBy,omitempty"`
	ReviewedAt  *time.Time            `db:"reviewed_at" json:"reviewedAt,omitempty"`
	TopicID     *uuid.UUID            `db:"topic_id" json:"topicId,omitempty"`
	CreatedAt   time.Time             `db:"created_at" json:"createdAt"`
}

func (TopicSubmission) TableName() string {
	return "topic_submissions"
}

func (s TopicSubmission) Pending() bool {
	return s.Status == TopicSubmissionStatusPending
}

func (s TopicSubmission) Reviewed() bool {
	return s.ReviewedAt != nil
}

func (s TopicSubmission) Accepted() bool {
	return s.TopicID != nil
}

func (s TopicSubmission) Rejected() bool {
	return s.Status == TopicSubmissionStatusRejected
}

func (s *TopicSubmission) ApplyDefaults() {
	if s.Tags == nil {
		s.Tags = StringArray{}
	}
}

func (s TopicSubmission) Validate() error {
	var errs []error

	if s.ID == uuid.Nil {
		errs = append(errs, errors.New("id is required"))
	}

	if s.SubmittedBy == "" {
		errs = append(errs, errors.New("submittedBy is required"))
	}

	if s.Title == "" {
		errs = append(errs, errors.New("title is required"))
	}

	if !s.Format.Valid() {
		errs = append(errs, errors.New("invalid format"))
	}

	if !s.Difficulty.Valid() {
		errs = append(errs, errors.New("invalid difficulty"))
	}

	if !s.Status.Valid() {
		errs = append(errs, errors.New("invalid status"))
	}

	if s.Status == TopicSubmissionStatusApproved && s.TopicID == nil {
		errs = append(
			errs,
			errors.New("topicId is required for approved submissions"),
		)
	}

	return errors.Join(errs...)
}

type Collection struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Slug        string    `db:"slug" json:"slug"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	IsActive    bool      `db:"is_active" json:"isActive"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

func (Collection) TableName() string {
	return "collections"
}

func (c Collection) Active() bool {
	return c.IsActive
}

func (c Collection) Validate() error {
	var errs []error

	if c.ID == uuid.Nil {
		errs = append(errs, errors.New("id is required"))
	}

	if c.Slug == "" {
		errs = append(errs, errors.New("slug is required"))
	}

	if c.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	return errors.Join(errs...)
}

type TopicCollection struct {
	TopicID      uuid.UUID `db:"topic_id" json:"topicId"`
	CollectionID uuid.UUID `db:"collection_id" json:"collectionId"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

func (TopicCollection) TableName() string {
	return "topic_collections"
}

func (tc TopicCollection) Validate() error {
	var errs []error

	if tc.TopicID == uuid.Nil {
		errs = append(errs, errors.New("topicId is required"))
	}

	if tc.CollectionID == uuid.Nil {
		errs = append(errs, errors.New("collectionId is required"))
	}

	return errors.Join(errs...)
}

type TopicBookmark struct {
	UserID    string    `db:"user_id" json:"userId"`
	TopicID   uuid.UUID `db:"topic_id" json:"topicId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (TopicBookmark) TableName() string {
	return "topic_bookmarks"
}

func (b TopicBookmark) Validate() error {
	var errs []error

	if b.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if b.TopicID == uuid.Nil {
		errs = append(errs, errors.New("topicId is required"))
	}

	return errors.Join(errs...)
}

type UserCategoryPreference struct {
	UserID     string    `db:"user_id" json:"userId"`
	CategoryID uuid.UUID `db:"category_id" json:"categoryId"`
	Weight     int       `db:"weight" json:"weight"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}

func (UserCategoryPreference) TableName() string {
	return "user_category_preferences"
}

func (p UserCategoryPreference) Validate() error {
	var errs []error

	if p.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if p.CategoryID == uuid.Nil {
		errs = append(errs, errors.New("categoryId is required"))
	}

	if p.Weight < 0 {
		errs = append(errs, errors.New("weight must be >= 0"))
	}

	return errors.Join(errs...)
}

type TopicSpin struct {
	ID               uuid.UUID    `db:"id" json:"id"`
	UserID           string       `db:"user_id" json:"userId"`
	TopicID          uuid.UUID    `db:"topic_id" json:"topicId"`
	CategoryFilterID *uuid.UUID   `db:"category_filter_id" json:"categoryFilterId,omitempty"`
	FormatFilter     *TopicFormat `db:"format_filter" json:"formatFilter,omitempty"`
	SessionID        *uuid.UUID   `db:"session_id" json:"sessionId,omitempty"`
	CreatedAt        time.Time    `db:"created_at" json:"createdAt"`
}

func (TopicSpin) TableName() string {
	return "topic_spins"
}

func (s TopicSpin) Validate() error {
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

	if s.FormatFilter != nil && !s.FormatFilter.Valid() {
		errs = append(errs, errors.New("invalid formatFilter"))
	}

	return errors.Join(errs...)
}

type UserTopicStats struct {
	UserID          string     `db:"user_id" json:"userId"`
	TopicID         uuid.UUID  `db:"topic_id" json:"topicId"`
	TimesSpun       int        `db:"times_spun" json:"timesSpun"`
	TimesCompleted  int        `db:"times_completed" json:"timesCompleted"`
	AverageScore    *float64   `db:"average_score" json:"averageScore,omitempty"`
	LastCompletedAt *time.Time `db:"last_completed_at" json:"lastCompletedAt,omitempty"`
}

func (UserTopicStats) TableName() string {
	return "user_topic_stats"
}

func (s UserTopicStats) Attempted() bool {
	return s.TimesSpun > 0
}

func (s UserTopicStats) Completed() bool {
	return s.TimesCompleted > 0
}

func (s UserTopicStats) Validate() error {
	var errs []error

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if s.TopicID == uuid.Nil {
		errs = append(errs, errors.New("topicId is required"))
	}

	if s.TimesSpun < 0 {
		errs = append(errs, errors.New("timesSpun must be >= 0"))
	}

	if s.TimesCompleted < 0 {
		errs = append(errs, errors.New("timesCompleted must be >= 0"))
	}

	if s.TimesCompleted > s.TimesSpun {
		errs = append(errs, errors.New("timesCompleted cannot exceed timesSpun"))
	}

	if s.AverageScore != nil && (*s.AverageScore < 0 || *s.AverageScore > 100) {
		errs = append(errs, errors.New("averageScore must be between 0 and 100"))
	}

	return errors.Join(errs...)
}
