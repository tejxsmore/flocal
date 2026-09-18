package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type XPTransaction struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	UserID         string     `db:"user_id" json:"userId"`
	SessionID      *uuid.UUID `db:"session_id" json:"sessionId,omitempty"`
	Amount         int        `db:"amount" json:"amount"`
	Reason         XPReason   `db:"reason" json:"reason"`
	IdempotencyKey string     `db:"idempotency_key" json:"-"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
}

func (XPTransaction) TableName() string {
	return "xp_transactions"
}

func (t XPTransaction) Positive() bool {
	return t.Amount > 0
}

func (t XPTransaction) Negative() bool {
	return t.Amount < 0
}

func (t XPTransaction) Validate() error {
	var errs []error

	if t.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if t.Amount == 0 {
		errs = append(errs, errors.New("amount must not be zero"))
	}

	if !t.Reason.Valid() {
		errs = append(errs, errors.New("invalid reason"))
	}

	if t.IdempotencyKey == "" {
		errs = append(errs, errors.New("idempotencyKey is required"))
	}

	return errors.Join(errs...)
}

type DailyActivity struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        string    `db:"user_id" json:"userId"`
	ActivityDate  DateOnly  `db:"activity_date" json:"activityDate"`
	SessionsCount int       `db:"sessions_count" json:"sessionsCount"`
	XPEarned      int       `db:"xp_earned" json:"xpEarned"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at" json:"updatedAt"`
}

func (DailyActivity) TableName() string {
	return "daily_activity"
}

func (a DailyActivity) Validate() error {
	var errs []error

	if a.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if a.ActivityDate.IsZero() {
		errs = append(errs, errors.New("activityDate is required"))
	}

	if a.SessionsCount < 0 {
		errs = append(
			errs,
			errors.New("sessionsCount must be >= 0"),
		)
	}

	if a.XPEarned < 0 {
		errs = append(
			errs,
			errors.New("xpEarned must be >= 0"),
		)
	}

	return errors.Join(errs...)
}

type UserStats struct {
	UserID            string    `db:"user_id" json:"userId"`
	TotalSessions     int       `db:"total_sessions" json:"totalSessions"`
	ScoredSessions    int       `db:"scored_sessions" json:"scoredSessions"`
	AverageScore      float64   `db:"average_score" json:"averageScore"`
	BestScore         float64   `db:"best_score" json:"bestScore"`
	CurrentStreakDays int       `db:"current_streak_days" json:"currentStreakDays"`
	LongestStreakDays int       `db:"longest_streak_days" json:"longestStreakDays"`
	LastSessionDate   *DateOnly `db:"last_session_date" json:"lastSessionDate,omitempty"`
	TotalXP           int64     `db:"total_xp" json:"totalXp"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
}

func (UserStats) TableName() string {
	return "user_stats"
}

func (s UserStats) HasActiveStreak() bool {
	return s.CurrentStreakDays > 0
}

func (s UserStats) Validate() error {
	var errs []error

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if s.TotalSessions < 0 {
		errs = append(
			errs,
			errors.New("totalSessions must be >= 0"),
		)
	}

	if s.ScoredSessions < 0 {
		errs = append(
			errs,
			errors.New("scoredSessions must be >= 0"),
		)
	}

	if s.ScoredSessions > s.TotalSessions {
		errs = append(
			errs,
			errors.New("scoredSessions cannot exceed totalSessions"),
		)
	}

	if s.AverageScore < 0 || s.AverageScore > 100 {
		errs = append(
			errs,
			errors.New("averageScore must be between 0 and 100"),
		)
	}

	if s.BestScore < 0 || s.BestScore > 100 {
		errs = append(
			errs,
			errors.New("bestScore must be between 0 and 100"),
		)
	}

	if s.CurrentStreakDays < 0 {
		errs = append(
			errs,
			errors.New("currentStreakDays must be >= 0"),
		)
	}

	if s.LongestStreakDays < 0 {
		errs = append(
			errs,
			errors.New("longestStreakDays must be >= 0"),
		)
	}

	if s.TotalXP < 0 {
		errs = append(
			errs,
			errors.New("totalXp must be >= 0"),
		)
	}

	if s.CurrentStreakDays > s.LongestStreakDays {
		errs = append(
			errs,
			errors.New("currentStreakDays cannot exceed longestStreakDays"),
		)
	}

	return errors.Join(errs...)
}

type UserSkillStats struct {
	UserID          string    `db:"user_id" json:"userId"`
	AvgClarity      *float64  `db:"avg_clarity" json:"avgClarity,omitempty"`
	AvgDelivery     *float64  `db:"avg_delivery" json:"avgDelivery,omitempty"`
	AvgContent      *float64  `db:"avg_content" json:"avgContent,omitempty"`
	AvgVocabulary   *float64  `db:"avg_vocabulary" json:"avgVocabulary,omitempty"`
	AvgGrammar      *float64  `db:"avg_grammar" json:"avgGrammar,omitempty"`
	SessionsCounted int       `db:"sessions_counted" json:"sessionsCounted"`
	UpdatedAt       time.Time `db:"updated_at" json:"updatedAt"`
}

func (UserSkillStats) TableName() string {
	return "user_skill_stats"
}

func (s UserSkillStats) Validate() error {
	var errs []error

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	for name, value := range map[string]*float64{
		"avgClarity":    s.AvgClarity,
		"avgDelivery":   s.AvgDelivery,
		"avgContent":    s.AvgContent,
		"avgVocabulary": s.AvgVocabulary,
		"avgGrammar":    s.AvgGrammar,
	} {
		if value != nil && (*value < 0 || *value > 100) {
			errs = append(
				errs,
				errors.New(name+" must be between 0 and 100"),
			)
		}
	}

	if s.SessionsCounted < 0 {
		errs = append(
			errs,
			errors.New("sessionsCounted must be >= 0"),
		)
	}

	return errors.Join(errs...)
}

type Badge struct {
	ID            uuid.UUID         `db:"id" json:"id"`
	Slug          string            `db:"slug" json:"slug"`
	Name          string            `db:"name" json:"name"`
	Description   *string           `db:"description" json:"description,omitempty"`
	Icon          *string           `db:"icon" json:"icon,omitempty"`
	CriteriaType  BadgeCriteriaType `db:"criteria_type" json:"criteriaType"`
	CriteriaValue *int              `db:"criteria_value" json:"criteriaValue,omitempty"`
	IsActive      bool              `db:"is_active" json:"isActive"`
	SortOrder     int               `db:"sort_order" json:"sortOrder"`
	CreatedAt     time.Time         `db:"created_at" json:"createdAt"`
}

func (Badge) TableName() string {
	return "badges"
}

func (b Badge) Validate() error {
	var errs []error

	if b.Slug == "" {
		errs = append(errs, errors.New("slug is required"))
	}

	if b.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	if !b.CriteriaType.Valid() {
		errs = append(
			errs,
			errors.New("invalid criteriaType"),
		)
	}

	if b.CriteriaValue != nil && *b.CriteriaValue < 0 {
		errs = append(
			errs,
			errors.New("criteriaValue must be >= 0"),
		)
	}

	if b.CriteriaType != BadgeCriteriaTypeManual &&
		b.CriteriaValue == nil {
		errs = append(
			errs,
			errors.New("criteriaValue is required for non-manual badges"),
		)
	}

	if b.SortOrder < 0 {
		errs = append(
			errs,
			errors.New("sortOrder must be >= 0"),
		)
	}

	return errors.Join(errs...)
}

func (b Badge) RequiresCriteriaValue() bool {
	return b.CriteriaType != BadgeCriteriaTypeManual
}

func (b Badge) Manual() bool {
	return b.CriteriaType == BadgeCriteriaTypeManual
}

type UserBadge struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    string     `db:"user_id" json:"userId"`
	BadgeID   uuid.UUID  `db:"badge_id" json:"badgeId"`
	SessionID *uuid.UUID `db:"session_id" json:"sessionId,omitempty"`
	EarnedAt  time.Time  `db:"earned_at" json:"earnedAt"`
}

func (UserBadge) TableName() string {
	return "user_badges"
}

func (b UserBadge) Validate() error {
	var errs []error

	if b.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if b.BadgeID == uuid.Nil {
		errs = append(errs, errors.New("badgeId is required"))
	}

	if b.EarnedAt.IsZero() {
		errs = append(errs, errors.New("earnedAt is required"))
	}

	return errors.Join(errs...)
}

type LeaderboardSnapshot struct {
	ID                uuid.UUID         `db:"id" json:"id"`
	Period            LeaderboardPeriod `db:"period" json:"period"`
	PeriodStart       DateOnly          `db:"period_start" json:"periodStart"`
	PeriodEnd         DateOnly          `db:"period_end" json:"periodEnd"`
	CountryCode       *string           `db:"country_code" json:"countryCode,omitempty"` // NEW: nil = global row, set = country row
	UserID            string            `db:"user_id" json:"userId"`
	Rank              int               `db:"rank" json:"rank"`
	TotalXP           int64             `db:"total_xp" json:"totalXp"`
	CurrentStreakDays int               `db:"current_streak_days" json:"currentStreakDays"`
	TotalSessions     int               `db:"total_sessions" json:"totalSessions"`
	CreatedAt         time.Time         `db:"created_at" json:"createdAt"`
}

func (LeaderboardSnapshot) TableName() string {
	return "leaderboard_snapshots"
}

func (s LeaderboardSnapshot) IsGlobal() bool {
	return s.CountryCode == nil
}

func (s LeaderboardSnapshot) IsCountry() bool {
	return s.CountryCode != nil
}

func (s LeaderboardSnapshot) Validate() error {
	var errs []error

	if !s.Period.Valid() {
		errs = append(errs, errors.New("invalid period"))
	}

	if s.PeriodStart.IsZero() {
		errs = append(errs, errors.New("periodStart is required"))
	}

	if s.PeriodEnd.IsZero() {
		errs = append(errs, errors.New("periodEnd is required"))
	}

	if !s.PeriodStart.IsZero() &&
		!s.PeriodEnd.IsZero() &&
		s.PeriodStart.Time().After(s.PeriodEnd.Time()) {
		errs = append(
			errs,
			errors.New("periodStart must be before or equal to periodEnd"),
		)
	}

	// NEW: mirrors the db check on user.country_code.
	if s.CountryCode != nil && !isValidCountryCode(*s.CountryCode) {
		errs = append(errs, errors.New("countryCode must be a 2-letter uppercase ISO code"))
	}

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if s.Rank <= 0 {
		errs = append(errs, errors.New("rank must be > 0"))
	}

	if s.TotalXP < 0 {
		errs = append(errs, errors.New("totalXp must be >= 0"))
	}

	if s.CurrentStreakDays < 0 {
		errs = append(
			errs,
			errors.New("currentStreakDays must be >= 0"),
		)
	}

	if s.TotalSessions < 0 {
		errs = append(
			errs,
			errors.New("totalSessions must be >= 0"),
		)
	}

	return errors.Join(errs...)
}
