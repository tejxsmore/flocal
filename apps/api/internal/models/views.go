package models

import (
	"time"

	"github.com/google/uuid"
)

type UserCurrentPlan struct {
	UserID             string              `db:"user_id" json:"userId"`
	PlanSlug           string              `db:"plan_slug" json:"planSlug"`
	DailySessionLimit  *int                `db:"daily_session_limit" json:"dailySessionLimit,omitempty"`
	CanViewAnalysis    bool                `db:"can_view_analysis" json:"canViewAnalysis"`
	PlanRenewsAt       *time.Time          `db:"plan_renews_at" json:"planRenewsAt,omitempty"`
	SubscriptionStatus *SubscriptionStatus `db:"subscription_status" json:"subscriptionStatus,omitempty"`
}

func (UserCurrentPlan) TableName() string {
	return "user_current_plan"
}

func (p UserCurrentPlan) Unlimited() bool {
	return p.DailySessionLimit == nil
}

func (p UserCurrentPlan) HasSubscription() bool {
	return p.SubscriptionStatus != nil
}

func (p UserCurrentPlan) Entitled() bool {
	if p.SubscriptionStatus == nil {
		return false
	}

	return p.SubscriptionStatus.IsEntitled()
}

type TopicUsageStats struct {
	TopicID    uuid.UUID `db:"topic_id" json:"topicId"`
	TopicTitle string    `db:"topic_title" json:"topicTitle"`
	SpinCount  int64     `db:"spin_count" json:"spinCount"`
}

func (TopicUsageStats) TableName() string {
	return "topic_usage_stats"
}

type LeaderboardEntry struct {
	UserID            string  `db:"user_id" json:"userId"`
	Username          *string `db:"username" json:"username,omitempty"`
	Image             *string `db:"image" json:"image,omitempty"`
	CountryCode       *string `db:"country_code" json:"countryCode,omitempty"` // NEW
	TotalXP           int64   `db:"total_xp" json:"totalXp"`
	CurrentStreakDays int     `db:"current_streak_days" json:"currentStreakDays"`
	TotalSessions     int     `db:"total_sessions" json:"totalSessions"`
	Rank              int64   `db:"rank" json:"rank"`
}

func (LeaderboardEntry) TableName() string {
	return "leaderboard"
}

type CountryLeaderboardEntry struct {
	UserID            string  `db:"user_id" json:"userId"`
	Username          *string `db:"username" json:"username,omitempty"`
	Image             *string `db:"image" json:"image,omitempty"`
	CountryCode       *string `db:"country_code" json:"countryCode,omitempty"`
	TotalXP           int64   `db:"total_xp" json:"totalXp"`
	CurrentStreakDays int     `db:"current_streak_days" json:"currentStreakDays"`
	TotalSessions     int     `db:"total_sessions" json:"totalSessions"`
	Rank              int64   `db:"rank" json:"rank"`
}

func (CountryLeaderboardEntry) TableName() string {
	return "country_leaderboard"
}

type SessionReport struct {
	SessionID        uuid.UUID     `db:"session_id" json:"sessionId"`
	UserID           string        `db:"user_id" json:"userId"`
	TopicID          uuid.UUID     `db:"topic_id" json:"topicId"`
	TopicTitle       string        `db:"topic_title" json:"topicTitle"`
	TopicFormat      TopicFormat   `db:"topic_format" json:"topicFormat"`
	TopicCategory    string        `db:"topic_category" json:"topicCategory"`
	Status           SessionStatus `db:"status" json:"status"`
	DebateStance     *string       `db:"debate_stance" json:"debateStance,omitempty"` // NEW
	PrepTimeSeconds  int           `db:"prep_time_seconds" json:"prepTimeSeconds"`
	SpeakTimeSeconds int           `db:"speak_time_seconds" json:"speakTimeSeconds"`

	AudioS3Key     *string    `db:"audio_s3_key" json:"-"`
	AudioExpiresAt *time.Time `db:"audio_expires_at" json:"audioExpiresAt,omitempty"`

	ShareTokenHash *string    `db:"share_token_hash" json:"-"`
	ShareEnabled   bool       `db:"share_enabled" json:"shareEnabled"`
	ShareExpiresAt *time.Time `db:"share_expires_at" json:"shareExpiresAt,omitempty"`

	AudioPlaybackURL *string `db:"-" json:"audioPlaybackUrl,omitempty"`

	Transcript *string `db:"transcript" json:"transcript,omitempty"`

	WordsPerMinute        *float64 `db:"words_per_minute" json:"wordsPerMinute,omitempty"`
	FillerWordCount       *int     `db:"filler_word_count" json:"fillerWordCount,omitempty"`
	FillerRate            *float64 `db:"filler_rate" json:"fillerRate,omitempty"`
	PauseRate             *float64 `db:"pause_rate" json:"pauseRate,omitempty"`
	LongestPauseSeconds   *float64 `db:"longest_pause_seconds" json:"longestPauseSeconds,omitempty"`
	UniqueWordCount       *int     `db:"unique_word_count" json:"uniqueWordCount,omitempty"`
	LexicalDiversity      *float64 `db:"lexical_diversity" json:"lexicalDiversity,omitempty"`
	SentenceCount         *int     `db:"sentence_count" json:"sentenceCount,omitempty"`
	AverageSentenceLength *float64 `db:"average_sentence_length" json:"averageSentenceLength,omitempty"`

	OverallScore    *float64 `db:"overall_score" json:"overallScore,omitempty"`
	ClarityScore    *float64 `db:"clarity_score" json:"clarityScore,omitempty"`
	DeliveryScore   *float64 `db:"delivery_score" json:"deliveryScore,omitempty"`
	ContentScore    *float64 `db:"content_score" json:"contentScore,omitempty"`
	VocabularyScore *float64 `db:"vocabulary_score" json:"vocabularyScore,omitempty"`
	GrammarScore    *float64 `db:"grammar_score" json:"grammarScore,omitempty"`
	Strengths       *JSONB   `db:"strengths" json:"strengths,omitempty"`
	Improvements    *JSONB   `db:"improvements" json:"improvements,omitempty"`
	CoachMessage    *string  `db:"coach_message" json:"coachMessage,omitempty"`

	GrammarCorrections    []GrammarCorrection    `db:"-" json:"grammarCorrections,omitempty"`
	VocabularySuggestions []VocabularySuggestion `db:"-" json:"vocabularySuggestions,omitempty"`

	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	CompletedAt *time.Time `db:"completed_at" json:"completedAt,omitempty"`
}

func (SessionReport) TableName() string {
	return "session_report"
}

func (r SessionReport) AudioAvailable() bool {
	if r.AudioS3Key == nil {
		return false
	}

	if r.AudioExpiresAt != nil && time.Now().After(*r.AudioExpiresAt) {
		return false
	}

	return true
}

func (r SessionReport) Shareable() bool {
	if !r.ShareEnabled || r.ShareTokenHash == nil {
		return false
	}

	if r.ShareExpiresAt != nil && time.Now().After(*r.ShareExpiresAt) {
		return false
	}

	return true
}

func (r SessionReport) Completed() bool {
	return r.Status == SessionStatusCompleted
}

func (r SessionReport) HasTranscript() bool {
	return r.Transcript != nil && *r.Transcript != ""
}

func (r SessionReport) HasAnalysis() bool {
	return r.OverallScore != nil
}

func (r SessionReport) HasDebateStance() bool {
	return r.DebateStance != nil
}

type SessionSummary struct {
	SessionID        uuid.UUID     `db:"session_id" json:"sessionId"`
	UserID           string        `db:"user_id" json:"userId"`
	TopicID          uuid.UUID     `db:"topic_id" json:"topicId"`
	TopicTitle       string        `db:"topic_title" json:"topicTitle"`
	TopicFormat      TopicFormat   `db:"topic_format" json:"topicFormat"`
	TopicCategory    string        `db:"topic_category" json:"topicCategory"`
	Status           SessionStatus `db:"status" json:"status"`
	DebateStance     *string       `db:"debate_stance" json:"debateStance,omitempty"` // NEW
	PrepTimeSeconds  int           `db:"prep_time_seconds" json:"prepTimeSeconds"`
	SpeakTimeSeconds int           `db:"speak_time_seconds" json:"speakTimeSeconds"`

	Transcript *string `db:"transcript" json:"transcript,omitempty"`

	WordsPerMinute  *float64 `db:"words_per_minute" json:"wordsPerMinute,omitempty"`
	FillerWordCount *int     `db:"filler_word_count" json:"fillerWordCount,omitempty"`

	OverallScore    *float64 `db:"overall_score" json:"overallScore,omitempty"`
	ClarityScore    *float64 `db:"clarity_score" json:"clarityScore,omitempty"`
	DeliveryScore   *float64 `db:"delivery_score" json:"deliveryScore,omitempty"`
	ContentScore    *float64 `db:"content_score" json:"contentScore,omitempty"`
	VocabularyScore *float64 `db:"vocabulary_score" json:"vocabularyScore,omitempty"`
	GrammarScore    *float64 `db:"grammar_score" json:"grammarScore,omitempty"`
	Strengths       *JSONB   `db:"strengths" json:"strengths,omitempty"`
	Improvements    *JSONB   `db:"improvements" json:"improvements,omitempty"`
	CoachMessage    *string  `db:"coach_message" json:"coachMessage,omitempty"`

	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	CompletedAt *time.Time `db:"completed_at" json:"completedAt,omitempty"`
}

func (SessionSummary) TableName() string {
	return "session_summary"
}

func (s SessionSummary) Completed() bool {
	return s.Status == SessionStatusCompleted
}

func (s SessionSummary) HasTranscript() bool {
	return s.Transcript != nil && *s.Transcript != ""
}

func (s SessionSummary) HasAnalysis() bool {
	return s.OverallScore != nil
}

type SessionSharePublic struct {
	SessionID     uuid.UUID   `db:"session_id" json:"sessionId"`
	TopicTitle    string      `db:"topic_title" json:"topicTitle"`
	TopicFormat   TopicFormat `db:"topic_format" json:"topicFormat"`
	TopicCategory string      `db:"topic_category" json:"topicCategory"`

	Transcript *string `db:"transcript" json:"transcript,omitempty"`

	OverallScore    *float64 `db:"overall_score" json:"overallScore,omitempty"`
	ClarityScore    *float64 `db:"clarity_score" json:"clarityScore,omitempty"`
	DeliveryScore   *float64 `db:"delivery_score" json:"deliveryScore,omitempty"`
	ContentScore    *float64 `db:"content_score" json:"contentScore,omitempty"`
	VocabularyScore *float64 `db:"vocabulary_score" json:"vocabularyScore,omitempty"`
	GrammarScore    *float64 `db:"grammar_score" json:"grammarScore,omitempty"`
	WordsPerMinute  *float64 `db:"words_per_minute" json:"wordsPerMinute,omitempty"`

	CompletedAt *time.Time `db:"completed_at" json:"completedAt,omitempty"`
}

func (SessionSharePublic) TableName() string {
	return "session_share_public"
}

func (s SessionSharePublic) HasTranscript() bool {
	return s.Transcript != nil && *s.Transcript != ""
}
