package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type SessionAnalysis struct {
	ID        uuid.UUID `db:"id" json:"id"`
	SessionID uuid.UUID `db:"session_id" json:"sessionId"`

	WordsPerMinute          *float64 `db:"words_per_minute" json:"wordsPerMinute,omitempty"`
	SpeakingDurationSeconds *float64 `db:"speaking_duration_seconds" json:"speakingDurationSeconds,omitempty"`

	FillerWordCount     int      `db:"filler_word_count" json:"fillerWordCount"`
	FillerWordBreakdown JSONB    `db:"filler_word_breakdown" json:"fillerWordBreakdown"`
	FillerRate          *float64 `db:"filler_rate" json:"fillerRate,omitempty"`

	PauseCount            int      `db:"pause_count" json:"pauseCount"`
	LongestPauseSeconds   *float64 `db:"longest_pause_seconds" json:"longestPauseSeconds,omitempty"`
	TotalPauseSeconds     *float64 `db:"total_pause_seconds" json:"totalPauseSeconds,omitempty"`
	PauseRate             *float64 `db:"pause_rate" json:"pauseRate,omitempty"`
	UniqueWordCount       *int     `db:"unique_word_count" json:"uniqueWordCount,omitempty"`
	LexicalDiversity      *float64 `db:"lexical_diversity" json:"lexicalDiversity,omitempty"`
	SentenceCount         *int     `db:"sentence_count" json:"sentenceCount,omitempty"`
	AverageSentenceLength *float64 `db:"average_sentence_length" json:"averageSentenceLength,omitempty"`

	OverallScore    float64  `db:"overall_score" json:"overallScore"`
	ClarityScore    *float64 `db:"clarity_score" json:"clarityScore,omitempty"`
	DeliveryScore   *float64 `db:"delivery_score" json:"deliveryScore,omitempty"`
	ContentScore    *float64 `db:"content_score" json:"contentScore,omitempty"`
	VocabularyScore *float64 `db:"vocabulary_score" json:"vocabularyScore,omitempty"`
	GrammarScore    *float64 `db:"grammar_score" json:"grammarScore,omitempty"`

	Strengths    JSONB   `db:"strengths" json:"strengths"`
	Improvements JSONB   `db:"improvements" json:"improvements"`
	CoachMessage *string `db:"coach_message" json:"coachMessage,omitempty"`

	AIProvider       string   `db:"ai_provider" json:"aiProvider"`
	AIModel          string   `db:"ai_model" json:"aiModel"`
	AIPromptVersion  string   `db:"ai_prompt_version" json:"aiPromptVersion"`
	PromptTemplateID *string  `db:"prompt_template_id" json:"promptTemplateId,omitempty"`
	AnalysisVersion  string   `db:"analysis_version" json:"analysisVersion"`
	AITemperature    *float64 `db:"ai_temperature" json:"aiTemperature,omitempty"`
	AISeed           *string  `db:"ai_seed" json:"-"`
	RequestID        *string  `db:"request_id" json:"requestId,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (SessionAnalysis) TableName() string {
	return "session_analysis"
}

func (a *SessionAnalysis) ApplyDefaults() {
	if a.FillerWordBreakdown.IsNull() {
		a.FillerWordBreakdown = JSONB("{}")
	}

	if a.Strengths.IsNull() {
		a.Strengths = JSONB("[]")
	}

	if a.Improvements.IsNull() {
		a.Improvements = JSONB("[]")
	}
}

func (a SessionAnalysis) Validate() error {
	var errs []error

	if a.SessionID == uuid.Nil {
		errs = append(errs, errors.New("sessionId is required"))
	}

	if a.FillerWordCount < 0 {
		errs = append(errs, errors.New("fillerWordCount must be >= 0"))
	}

	if !a.FillerWordBreakdown.IsNull() && !a.FillerWordBreakdown.IsObject() {
		errs = append(errs, errors.New("fillerWordBreakdown must be a JSON object"))
	}

	if a.PauseCount < 0 {
		errs = append(errs, errors.New("pauseCount must be >= 0"))
	}

	if a.UniqueWordCount != nil && *a.UniqueWordCount < 0 {
		errs = append(errs, errors.New("uniqueWordCount must be >= 0"))
	}

	if a.SentenceCount != nil && *a.SentenceCount < 0 {
		errs = append(errs, errors.New("sentenceCount must be >= 0"))
	}

	if a.WordsPerMinute != nil && *a.WordsPerMinute < 0 {
		errs = append(errs, errors.New("wordsPerMinute must be >= 0"))
	}

	if a.SpeakingDurationSeconds != nil && *a.SpeakingDurationSeconds < 0 {
		errs = append(errs, errors.New("speakingDurationSeconds must be >= 0"))
	}

	if a.FillerRate != nil && *a.FillerRate < 0 {
		errs = append(errs, errors.New("fillerRate must be >= 0"))
	}

	if a.LongestPauseSeconds != nil && *a.LongestPauseSeconds < 0 {
		errs = append(errs, errors.New("longestPauseSeconds must be >= 0"))
	}

	if a.TotalPauseSeconds != nil && *a.TotalPauseSeconds < 0 {
		errs = append(errs, errors.New("totalPauseSeconds must be >= 0"))
	}

	if a.PauseRate != nil && *a.PauseRate < 0 {
		errs = append(errs, errors.New("pauseRate must be >= 0"))
	}

	if a.LexicalDiversity != nil &&
		(*a.LexicalDiversity < 0 || *a.LexicalDiversity > 1) {
		errs = append(errs, errors.New("lexicalDiversity must be between 0 and 1"))
	}

	if a.AverageSentenceLength != nil && *a.AverageSentenceLength < 0 {
		errs = append(errs, errors.New("averageSentenceLength must be >= 0"))
	}

	if a.OverallScore < 0 || a.OverallScore > 100 {
		errs = append(errs, errors.New("overallScore must be between 0 and 100"))
	}

	for name, value := range map[string]*float64{
		"clarityScore":    a.ClarityScore,
		"deliveryScore":   a.DeliveryScore,
		"contentScore":    a.ContentScore,
		"vocabularyScore": a.VocabularyScore,
		"grammarScore":    a.GrammarScore,
	} {
		if value != nil && (*value < 0 || *value > 100) {
			errs = append(
				errs,
				errors.New(name+" must be between 0 and 100"),
			)
		}
	}

	if !a.Strengths.IsNull() && !a.Strengths.IsArray() {
		errs = append(errs, errors.New("strengths must be a JSON array"))
	}

	if !a.Improvements.IsNull() && !a.Improvements.IsArray() {
		errs = append(errs, errors.New("improvements must be a JSON array"))
	}

	if a.AITemperature != nil &&
		(*a.AITemperature < 0 || *a.AITemperature > 2) {
		errs = append(errs, errors.New("aiTemperature must be between 0 and 2"))
	}

	if a.AIProvider == "" {
		errs = append(errs, errors.New("aiProvider is required"))
	}

	if a.AIModel == "" {
		errs = append(errs, errors.New("aiModel is required"))
	}

	if a.AIPromptVersion == "" {
		errs = append(errs, errors.New("aiPromptVersion is required"))
	}

	if a.AnalysisVersion == "" {
		errs = append(errs, errors.New("analysisVersion is required"))
	}

	return errors.Join(errs...)
}

func (a SessionAnalysis) StrengthsList() ([]string, error) {
	var out []string

	if err := a.Strengths.Unmarshal(&out); err != nil {
		return nil, err
	}

	return out, nil
}

func (a SessionAnalysis) ImprovementsList() ([]string, error) {
	var out []string

	if err := a.Improvements.Unmarshal(&out); err != nil {
		return nil, err
	}

	return out, nil
}

func (a SessionAnalysis) FillerWordBreakdownMap() (map[string]int, error) {
	out := make(map[string]int)

	if err := a.FillerWordBreakdown.Unmarshal(&out); err != nil {
		return nil, err
	}

	return out, nil
}

type AIProviderLog struct {
	ID        uuid.UUID `db:"id" json:"id"`
	SessionID uuid.UUID `db:"session_id" json:"sessionId"`

	ProcessingAttemptID *uuid.UUID          `db:"processing_attempt_id" json:"processingAttemptId,omitempty"`
	Source              AIProviderLogSource `db:"source" json:"source"`
	Operation           *string             `db:"operation" json:"operation,omitempty"`

	InputTokens  *int64 `db:"input_tokens" json:"inputTokens,omitempty"`
	OutputTokens *int64 `db:"output_tokens" json:"outputTokens,omitempty"`
	CachedTokens *int64 `db:"cached_tokens" json:"cachedTokens,omitempty"`
	LatencyMs    *int   `db:"latency_ms" json:"latencyMs,omitempty"`
	CostSubunits *int64 `db:"cost_subunits" json:"costSubunits,omitempty"`

	RawResponse *JSONB `db:"raw_response" json:"rawResponse,omitempty"`

	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	PurgeAfter time.Time `db:"purge_after" json:"purgeAfter"`
}

func (AIProviderLog) TableName() string {
	return "ai_provider_logs"
}

func (l AIProviderLog) Expired() bool {
	return time.Now().After(l.PurgeAfter)
}

func (l AIProviderLog) Validate() error {
	var errs []error

	if l.SessionID == uuid.Nil {
		errs = append(errs, errors.New("sessionId is required"))
	}

	if !l.Source.Valid() {
		errs = append(errs, errors.New("invalid source"))
	}

	if l.InputTokens != nil && *l.InputTokens < 0 {
		errs = append(errs, errors.New("inputTokens must be >= 0"))
	}

	if l.OutputTokens != nil && *l.OutputTokens < 0 {
		errs = append(errs, errors.New("outputTokens must be >= 0"))
	}

	if l.CachedTokens != nil && *l.CachedTokens < 0 {
		errs = append(errs, errors.New("cachedTokens must be >= 0"))
	}

	if l.LatencyMs != nil && *l.LatencyMs < 0 {
		errs = append(errs, errors.New("latencyMs must be >= 0"))
	}

	if l.CostSubunits != nil && *l.CostSubunits < 0 {
		errs = append(errs, errors.New("costSubunits must be >= 0"))
	}

	return errors.Join(errs...)
}

type GrammarCorrection struct {
	ID           uuid.UUID `db:"id" json:"id"`
	SessionID    uuid.UUID `db:"session_id" json:"sessionId"`
	TranscriptID uuid.UUID `db:"transcript_id" json:"transcriptId"`

	OriginalText   string  `db:"original_text" json:"originalText"`
	CorrectedText  string  `db:"corrected_text" json:"correctedText"`
	Explanation    *string `db:"explanation" json:"explanation,omitempty"`
	ErrorType      *string `db:"error_type" json:"errorType,omitempty"`
	ContextSnippet *string `db:"context_snippet" json:"contextSnippet,omitempty"`

	StartChar *int `db:"start_char" json:"startChar,omitempty"`
	EndChar   *int `db:"end_char" json:"endChar,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (GrammarCorrection) TableName() string {
	return "grammar_corrections"
}

func (g GrammarCorrection) Validate() error {
	var errs []error

	if g.SessionID == uuid.Nil {
		errs = append(errs, errors.New("sessionId is required"))
	}

	if g.TranscriptID == uuid.Nil {
		errs = append(errs, errors.New("transcriptId is required"))
	}

	if g.OriginalText == "" {
		errs = append(errs, errors.New("originalText is required"))
	}

	if g.CorrectedText == "" {
		errs = append(errs, errors.New("correctedText is required"))
	}

	if g.StartChar != nil && *g.StartChar < 0 {
		errs = append(errs, errors.New("startChar must be >= 0"))
	}

	if g.EndChar != nil {
		if *g.EndChar < 0 {
			errs = append(errs, errors.New("endChar must be >= 0"))
		}

		if g.StartChar != nil && *g.EndChar < *g.StartChar {
			errs = append(
				errs,
				errors.New("endChar must be >= startChar"),
			)
		}
	}

	return errors.Join(errs...)
}

type VocabularySuggestion struct {
	ID           uuid.UUID `db:"id" json:"id"`
	SessionID    uuid.UUID `db:"session_id" json:"sessionId"`
	TranscriptID uuid.UUID `db:"transcript_id" json:"transcriptId"`

	OriginalWord   string      `db:"original_word" json:"originalWord"`
	SuggestedWords StringArray `db:"suggested_words" json:"suggestedWords"`
	Reason         *string     `db:"reason" json:"reason,omitempty"`
	ContextSnippet *string     `db:"context_snippet" json:"contextSnippet,omitempty"`

	StartChar *int `db:"start_char" json:"startChar,omitempty"`
	EndChar   *int `db:"end_char" json:"endChar,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func (VocabularySuggestion) TableName() string {
	return "vocabulary_suggestions"
}

func (v *VocabularySuggestion) ApplyDefaults() {
	if v.SuggestedWords == nil {
		v.SuggestedWords = StringArray{}
	}
}

func (v VocabularySuggestion) Validate() error {
	var errs []error

	if v.SessionID == uuid.Nil {
		errs = append(errs, errors.New("sessionId is required"))
	}

	if v.TranscriptID == uuid.Nil {
		errs = append(errs, errors.New("transcriptId is required"))
	}

	if v.OriginalWord == "" {
		errs = append(errs, errors.New("originalWord is required"))
	}

	if v.SuggestedWords == nil {
		errs = append(errs, errors.New("suggestedWords is required"))
	}

	if v.StartChar != nil && *v.StartChar < 0 {
		errs = append(errs, errors.New("startChar must be >= 0"))
	}

	if v.EndChar != nil {
		if *v.EndChar < 0 {
			errs = append(errs, errors.New("endChar must be >= 0"))
		}

		if v.StartChar != nil && *v.EndChar < *v.StartChar {
			errs = append(
				errs,
				errors.New("endChar must be >= startChar"),
			)
		}
	}

	return errors.Join(errs...)
}
