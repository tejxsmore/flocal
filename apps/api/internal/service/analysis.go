package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/google/uuid"

	"flocal/internal/config"
	"flocal/internal/models"
	"flocal/internal/repository"
)

const genericAnalysisFailureMessage = "We couldn't complete the analysis for this session. Please try again."
const pauseThresholdSeconds = 0.3
const maxGrammarCorrections = 3
const maxVocabularySuggestions = 3
const contextSnippetRadius = 4
const minWordCountForAIAnalysis = 15
const minSeverity = 1
const maxSeverity = 3
const defaultSeverity = 2

const initialReasoningMaxTokens = 2000
const maxReasoningMaxTokens = 4000
const initialStandardMaxTokens = 590
const maxStandardMaxTokens = 1200
const tokenEscalationFactor = 2

var ErrSessionNotEligibleForReanalysis = errors.New("service: session is not eligible for reanalysis")
var sentenceSplitRegex = regexp.MustCompile(`[.!?]+`)

var validGrammarErrorTypes = map[string]struct{}{
	"subject-verb": {},
	"tense":        {},
	"article":      {},
	"preposition":  {},
	"word-order":   {},
	"other":        {},
}

var (
	droppedGrammarFindings     atomic.Int64
	droppedVocabFindings       atomic.Int64
	maxTokensExceededRetries   atomic.Int64
	maxTokensExceededExhausted atomic.Int64
)

func DroppedFindingsStats() (droppedGrammar, droppedVocab int64) {
	return droppedGrammarFindings.Load(), droppedVocabFindings.Load()
}

func MaxTokensExceededStats() (retried, exhausted int64) {
	return maxTokensExceededRetries.Load(), maxTokensExceededExhausted.Load()
}

type grammarCorrectionItem struct {
	OriginalText   string
	CorrectedText  string
	Explanation    string
	ErrorType      string
	PositionIndex  int
	Severity       int
	ContextSnippet string
}

type vocabularySuggestionItem struct {
	OriginalWord   string
	SuggestedWords []string
	Reason         string
	ContextSnippet string
	PositionIndex  int
	Severity       int
}

type compactAnalysisResult struct {
	Scores struct {
		Relevance  int
		Coherence  int
		Vocabulary int
		Grammar    int
	}
	Strength    string
	Improvement string
	Tip         string

	GrammarCorrections    []grammarCorrectionItem
	VocabularySuggestions []vocabularySuggestionItem
}

type compactAnalysisResultRaw struct {
	Scores struct {
		Relevance  int `json:"r"`
		Coherence  int `json:"c"`
		Vocabulary int `json:"v"`
		Grammar    int `json:"g"`
	} `json:"sc"`
	Strength    string `json:"st"`
	Improvement string `json:"im"`
	Tip         string `json:"tp"`

	GrammarCorrections []struct {
		OriginalText  string `json:"ot"`
		CorrectedText string `json:"ct"`
		Explanation   string `json:"ex"`
		ErrorType     string `json:"et"`
		PositionIndex int    `json:"pi"`
		Severity      int    `json:"sv"`
	} `json:"gc"`

	VocabularySuggestions []struct {
		OriginalWord   string   `json:"ow"`
		SuggestedWords []string `json:"sw"`
		Reason         string   `json:"rs"`
		PositionIndex  int      `json:"pi"`
		Severity       int      `json:"sv"`
	} `json:"vs"`
}

type openAICallResult struct {
	result       *compactAnalysisResult
	rawResponse  string
	statusCode   int
	finishReason string
	inputTokens  int64
	outputTokens int64
	cachedTokens int64
}

const compactSystemPrompt = `You are a strict, experienced speaking fluency examiner. Score a ~60s spoken response transcript on 4 dimensions, 0-10 integers:
relevance: addresses the topic
coherence: logical flow and structure of ideas
vocabulary: word choice range and precision
grammar: correctness of grammar and sentence construction

Ignore pace, filler words, and pauses; those are scored separately from audio.

Be strict and use the full scale. Do not default to high scores:
0-2: barely functional, largely fails the dimension
3-4: basic, frequent notable issues
5-6: developing, gets the point across with clear weaknesses
7-8: competent, solid with only minor issues
9-10: excellent, reserve for genuinely exceptional responses with essentially no flaws in that dimension
A score of 9 or 10 in grammar or vocabulary is only valid if you found zero errors/weak choices to report for that dimension.

Give one strength, one improvement, one tip. Each a short, specific phrase, max 12 words, grounded in something actually present in the transcript.

Match the number of findings to how much is actually wrong, not to the maximum allowed - a short or mostly-clean response should get fewer (or zero) findings, not padded ones.
List up to 3 grammar errors (only real errors, skip if none) and up to 3 weak, repetitive, or imprecise word choices (only if a clearly better word exists, skip if none). Every item must quote text that literally appears in the transcript - never invent or paraphrase an error that isn't there.
For each grammar error: original text (verbatim substring from transcript), corrected text (must differ meaningfully from the original), explanation (max 6 words), errorType (subject-verb, tense, article, preposition, word-order, or other), positionIndex (0-based index of the error's first word in the transcript's space-separated word list), severity (1=minor/nitpick, 2=noticeable, 3=seriously undermines the sentence).
For each vocabulary suggestion: original word (verbatim, exact word from transcript), up to 2 suggested words that are not the original word, reason (max 5 words), positionIndex, severity (1=minor stylistic upgrade, 2=noticeably weak or repetitive, 3=vague or imprecise enough to blur meaning).

Output only the JSON object, nothing else.`

type SessionAnalysisService struct {
	cfg             config.OpenAIConfig
	sessionRepo     repository.SessionRepository
	transcriptRepo  repository.TranscriptRepository
	analysisRepo    repository.AnalysisRepository
	topicRepo       repository.TopicRepository
	viewRepo        repository.ViewRepository
	vocabularySvc   *VocabularyService
	gamificationSvc *GamificationService
	httpClient      *http.Client
}

func NewSessionAnalysisService(
	cfg config.OpenAIConfig,
	sessionRepo repository.SessionRepository,
	transcriptRepo repository.TranscriptRepository,
	analysisRepo repository.AnalysisRepository,
	topicRepo repository.TopicRepository,
	viewRepo repository.ViewRepository,
	vocabularySvc *VocabularyService,
	gamificationSvc *GamificationService,
) *SessionAnalysisService {
	return &SessionAnalysisService{
		cfg:             cfg,
		sessionRepo:     sessionRepo,
		transcriptRepo:  transcriptRepo,
		analysisRepo:    analysisRepo,
		topicRepo:       topicRepo,
		viewRepo:        viewRepo,
		vocabularySvc:   vocabularySvc,
		gamificationSvc: gamificationSvc,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
	}
}

const freeLifetimeAISessions = 5
const freeDailyAISessions = 1

type AIQuotaStatus struct {
	Eligible      bool  `json:"eligible"`
	Unlimited     bool  `json:"unlimited"`
	LifetimeUsed  int64 `json:"lifetimeUsed"`
	LifetimeLimit int   `json:"lifetimeLimit"`
	DailyUsed     int64 `json:"dailyUsed"`
	DailyLimit    int   `json:"dailyLimit"`
}

func (s *SessionAnalysisService) AIAnalysisQuota(ctx context.Context, userID string) (*AIQuotaStatus, error) {
	plan, err := s.viewRepo.GetUserCurrentPlan(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("get user current plan: %w", err)
	}

	if plan != nil && plan.CanViewAnalysis {
		return &AIQuotaStatus{Eligible: true, Unlimited: true}, nil
	}

	lifetimeUsed, err := s.analysisRepo.CountAnalyzedSessionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count lifetime analyzed sessions: %w", err)
	}

	if lifetimeUsed < freeLifetimeAISessions {
		return &AIQuotaStatus{
			Eligible:      true,
			LifetimeUsed:  lifetimeUsed,
			LifetimeLimit: freeLifetimeAISessions,
			DailyLimit:    freeDailyAISessions,
		}, nil
	}

	dayStart := time.Now().UTC().Truncate(24 * time.Hour)

	dailyUsed, err := s.analysisRepo.CountAnalyzedSessionsForUserSince(ctx, userID, dayStart)
	if err != nil {
		return nil, fmt.Errorf("count daily analyzed sessions: %w", err)
	}

	return &AIQuotaStatus{
		Eligible:      dailyUsed < freeDailyAISessions,
		LifetimeUsed:  lifetimeUsed,
		LifetimeLimit: freeLifetimeAISessions,
		DailyUsed:     dailyUsed,
		DailyLimit:    freeDailyAISessions,
	}, nil
}

func (s *SessionAnalysisService) AIAnalysisEligible(ctx context.Context, userID string) (bool, error) {
	quota, err := s.AIAnalysisQuota(ctx, userID)
	if err != nil {
		return false, err
	}
	return quota.Eligible, nil
}

type charSpan struct {
	start int
	end   int
}

func wordCharSpans(text string) []charSpan {
	var spans []charSpan
	inWord := false
	start := 0

	for i, r := range text {
		if unicode.IsSpace(r) {
			if inWord {
				spans = append(spans, charSpan{start: start, end: i})
				inWord = false
			}
			continue
		}
		if !inWord {
			start = i
			inWord = true
		}
	}

	if inWord {
		spans = append(spans, charSpan{start: start, end: len(text)})
	}

	return spans
}

func spanStartPtr(spans []charSpan, idx int) *int {
	if idx < 0 || idx >= len(spans) {
		return nil
	}
	v := spans[idx].start
	return &v
}

func spanEndPtr(spans []charSpan, idx int) *int {
	if idx < 0 || idx >= len(spans) {
		return nil
	}
	v := spans[idx].end
	return &v
}

func floatPtr(v float64) *float64 { return &v }
func intPtr(v int) *int           { return &v }
func int64Ptr(v int64) *int64     { return &v }
func strPtr(v string) *string     { return &v }

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type speechMetrics struct {
	fillerCount           int
	fillerBreakdown       map[string]int
	fillerRate            float64
	wpm                   float64
	speakingDuration      float64
	pauseCount            int
	longestPauseSeconds   float64
	totalPauseSeconds     float64
	pauseRate             float64
	uniqueWordCount       int
	lexicalDiversity      float64
	sentenceCount         int
	averageSentenceLength float64
}

func computeSpeechMetrics(t *models.Transcript) speechMetrics {
	m := speechMetrics{fillerBreakdown: map[string]int{}}

	var words []DeepgramWord
	if t.WordTimings != nil {
		_ = t.WordTimings.Unmarshal(&words)
	}

	seen := make(map[string]struct{})

	for i, word := range words {
		normalized := strings.ToLower(strings.Trim(word.Word, ".,!?;:\"'"))

		if IsFillerWord(word.Word) {
			m.fillerCount++
			m.fillerBreakdown[normalized]++
		}

		if normalized != "" {
			seen[normalized] = struct{}{}
		}

		if i > 0 {
			gap := word.Start - words[i-1].End
			if gap > pauseThresholdSeconds {
				m.pauseCount++
				m.totalPauseSeconds += gap
				if gap > m.longestPauseSeconds {
					m.longestPauseSeconds = gap
				}
			}
		}
	}

	m.uniqueWordCount = len(seen)

	if len(words) > 0 {
		m.speakingDuration = words[len(words)-1].End - words[0].Start
	}

	if m.speakingDuration > 0 {
		m.wpm = float64(t.WordCount) / (m.speakingDuration / 60)
		m.pauseRate = clampScore(m.totalPauseSeconds/m.speakingDuration, 0, 1)
	}

	if t.WordCount > 0 {
		m.fillerRate = float64(m.fillerCount) / float64(t.WordCount)
		m.lexicalDiversity = clampScore(float64(m.uniqueWordCount)/float64(t.WordCount), 0, 1)
	}

	for _, sentence := range sentenceSplitRegex.Split(t.RawText, -1) {
		if strings.TrimSpace(sentence) != "" {
			m.sentenceCount++
		}
	}
	if m.sentenceCount > 0 {
		m.averageSentenceLength = float64(t.WordCount) / float64(m.sentenceCount)
	}

	return m
}

func (s *SessionAnalysisService) AnalyzeSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("service: analyze session: get session: %w", err)
	}

	if !session.Processing() {
		log.Printf("session analysis: session %s is not in processing state, skipping duplicate/stale trigger", sessionID)
		return nil
	}

	eligible, err := s.AIAnalysisEligible(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("service: analyze session: check eligibility: %w", err)
	}

	if !eligible {
		if err := s.sessionRepo.MarkCompleted(ctx, sessionID); err != nil {
			return fmt.Errorf("service: analyze session: mark completed: %w", err)
		}

		if _, err := s.gamificationSvc.AwardSessionXP(ctx, session.UserID, sessionID, nil); err != nil {
			log.Printf("session analysis: award xp failed for session %s: %v", sessionID, err)
		}

		return nil
	}

	transcript, err := s.transcriptRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("service: analyze session: get transcript: %w", err)
	}

	if transcript.WordCount < minWordCountForAIAnalysis {
		log.Printf("session analysis: session %s below minimum word count for AI analysis (%d words)", sessionID, transcript.WordCount)

		if err := s.sessionRepo.MarkCompleted(ctx, sessionID); err != nil {
			return fmt.Errorf("service: analyze session: mark completed: %w", err)
		}

		if _, err := s.gamificationSvc.AwardSessionXP(ctx, session.UserID, sessionID, nil); err != nil {
			log.Printf("session analysis: award xp failed for session %s: %v", sessionID, err)
		}

		return nil
	}

	topic, err := s.topicRepo.GetByID(ctx, session.TopicID)
	if err != nil {
		return fmt.Errorf("service: analyze session: get topic: %w", err)
	}

	metrics := computeSpeechMetrics(transcript)
	deliveryScore := computeDeliveryScore(metrics.wpm, metrics.fillerCount, transcript.WordCount)

	callStart := time.Now()
	callResult, err := s.callOpenAI(ctx, topic.Title, transcript.RawText)
	latencyMs := int(time.Since(callStart).Milliseconds())
	if err != nil {
		statusCode := 0
		finishReason := ""
		if callResult != nil {
			statusCode = callResult.statusCode
			finishReason = callResult.finishReason
		}
		log.Printf("session analysis: openai call failed for session %s: %v (finishReason=%q)", sessionID, err, finishReason)
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, sanitizeOpenAIFailure(statusCode))
		return fmt.Errorf("service: analyze session: openai call: %w", err)
	}

	result := callResult.result

	if err := validateCompactResult(result); err != nil {
		log.Printf("session analysis: invalid openai result for session %s: %v", sessionID, err)
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: invalid openai result: %w", err)
	}

	if err := s.analysisRepo.DeleteForSession(ctx, sessionID); err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: clear previous analysis: %w", err)
	}

	rawJSONB, err := models.NewJSONB(json.RawMessage(callResult.rawResponse))
	if err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: marshal ai response: %w", err)
	}

	logEntry := &models.AIProviderLog{
		ID:           uuid.New(),
		SessionID:    sessionID,
		Source:       models.AIProviderLogSourceOpenAI,
		Operation:    strPtr("session_analysis"),
		InputTokens:  int64Ptr(callResult.inputTokens),
		OutputTokens: int64Ptr(callResult.outputTokens),
		CachedTokens: int64Ptr(callResult.cachedTokens),
		LatencyMs:    intPtr(latencyMs),
		RawResponse:  &rawJSONB,
		PurgeAfter:   time.Now().Add(30 * 24 * time.Hour),
	}

	if err := s.analysisRepo.CreateAIProviderLog(ctx, logEntry); err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: log ai response: %w", err)
	}

	contentScore := float64(result.Scores.Relevance) * 10
	clarityScore := float64(result.Scores.Coherence) * 10
	vocabularyScore := float64(result.Scores.Vocabulary) * 10
	grammarScore := float64(result.Scores.Grammar) * 10
	overallScore := computeOverallScore(clarityScore, deliveryScore, contentScore, vocabularyScore, grammarScore)

	strengthsJSONB, err := models.NewJSONB([]string{result.Strength})
	if err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: marshal strengths: %w", err)
	}

	improvementsJSONB, err := models.NewJSONB([]string{result.Improvement})
	if err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: marshal improvements: %w", err)
	}

	fillerBreakdownJSONB, err := models.NewJSONB(metrics.fillerBreakdown)
	if err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: marshal filler breakdown: %w", err)
	}

	coachMessage := result.Tip

	analysis := &models.SessionAnalysis{
		ID:                      uuid.New(),
		SessionID:               sessionID,
		WordsPerMinute:          &metrics.wpm,
		SpeakingDurationSeconds: &metrics.speakingDuration,
		FillerWordCount:         metrics.fillerCount,
		FillerWordBreakdown:     fillerBreakdownJSONB,
		FillerRate:              floatPtr(metrics.fillerRate),
		PauseCount:              metrics.pauseCount,
		LongestPauseSeconds:     floatPtr(metrics.longestPauseSeconds),
		TotalPauseSeconds:       floatPtr(metrics.totalPauseSeconds),
		PauseRate:               floatPtr(metrics.pauseRate),
		UniqueWordCount:         intPtr(metrics.uniqueWordCount),
		LexicalDiversity:        floatPtr(metrics.lexicalDiversity),
		SentenceCount:           intPtr(metrics.sentenceCount),
		AverageSentenceLength:   floatPtr(metrics.averageSentenceLength),
		OverallScore:            overallScore,
		ClarityScore:            &clarityScore,
		DeliveryScore:           &deliveryScore,
		ContentScore:            &contentScore,
		VocabularyScore:         &vocabularyScore,
		GrammarScore:            &grammarScore,
		Strengths:               strengthsJSONB,
		Improvements:            improvementsJSONB,
		CoachMessage:            &coachMessage,
		AIProvider:              "openai",
		AIModel:                 s.cfg.Model,
		AIPromptVersion:         "v4-strict",
		AnalysisVersion:         "v3",
	}

	if err := s.analysisRepo.CreateSessionAnalysis(ctx, analysis); err != nil {
		_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
		return fmt.Errorf("service: analyze session: save analysis: %w", err)
	}

	wordSpans := wordCharSpans(transcript.RawText)

	if len(result.GrammarCorrections) > 0 {
		corrections := make([]models.GrammarCorrection, 0, len(result.GrammarCorrections))

		for _, gc := range result.GrammarCorrections {
			explanation := gc.Explanation
			errorType := gc.ErrorType
			snippet := gc.ContextSnippet

			corrections = append(corrections, models.GrammarCorrection{
				ID:             uuid.New(),
				SessionID:      sessionID,
				TranscriptID:   transcript.ID,
				OriginalText:   gc.OriginalText,
				CorrectedText:  gc.CorrectedText,
				Explanation:    &explanation,
				ErrorType:      &errorType,
				ContextSnippet: &snippet,
				StartChar:      spanStartPtr(wordSpans, gc.PositionIndex),
				EndChar:        spanEndPtr(wordSpans, gc.PositionIndex),
			})
		}

		if err := s.analysisRepo.CreateGrammarCorrections(ctx, corrections); err != nil {
			_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
			return fmt.Errorf("service: analyze session: save grammar corrections: %w", err)
		}
	}

	if len(result.VocabularySuggestions) > 0 {
		suggestions := make([]models.VocabularySuggestion, 0, len(result.VocabularySuggestions))

		for _, vs := range result.VocabularySuggestions {
			reason := vs.Reason
			snippet := vs.ContextSnippet

			suggestions = append(suggestions, models.VocabularySuggestion{
				ID:             uuid.New(),
				SessionID:      sessionID,
				TranscriptID:   transcript.ID,
				OriginalWord:   vs.OriginalWord,
				SuggestedWords: models.StringArray(vs.SuggestedWords),
				Reason:         &reason,
				ContextSnippet: &snippet,
				StartChar:      spanStartPtr(wordSpans, vs.PositionIndex),
				EndChar:        spanEndPtr(wordSpans, vs.PositionIndex),
			})
		}

		if err := s.analysisRepo.CreateVocabularySuggestions(ctx, suggestions); err != nil {
			_ = s.sessionRepo.MarkFailed(ctx, sessionID, genericAnalysisFailureMessage)
			return fmt.Errorf("service: analyze session: save vocabulary suggestions: %w", err)
		}

		for _, vs := range result.VocabularySuggestions {
			if _, err := s.vocabularySvc.RecordEncounter(ctx, session.UserID, vs.OriginalWord, nil, nil, true); err != nil {
				log.Printf("session analysis: record vocabulary encounter failed for %q: %v", vs.OriginalWord, err)
			}
		}
	}

	if err := s.sessionRepo.MarkCompleted(ctx, sessionID); err != nil {
		return fmt.Errorf("service: analyze session: mark completed: %w", err)
	}

	if _, err := s.gamificationSvc.AwardSessionXP(ctx, session.UserID, sessionID, &overallScore); err != nil {
		log.Printf("session analysis: award xp failed for session %s: %v", sessionID, err)
	}

	if _, err := s.gamificationSvc.EvaluateAndAwardBadges(ctx, session.UserID, &sessionID); err != nil {
		log.Printf("session analysis: evaluate badges failed for session %s: %v", sessionID, err)
	}

	return nil
}

func (s *SessionAnalysisService) EnsureReanalyzable(
	ctx context.Context,
	sessionID uuid.UUID,
	userID string,
) (*models.SpeakingSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("service: reanalyze session: get session: %w", err)
	}

	if session.UserID != userID {
		return nil, ErrSessionAccessDenied
	}

	if !session.Failed() && !session.Processing() {
		return nil, ErrSessionNotEligibleForReanalysis
	}

	return session, nil
}

func stripFillerWordsForPrompt(text string) (string, []int) {
	words := strings.Fields(text)
	kept := make([]string, 0, len(words))
	indexMap := make([]int, 0, len(words))

	for i, w := range words {
		if IsFillerWord(w) {
			continue
		}
		kept = append(kept, w)
		indexMap = append(indexMap, i)
	}

	return strings.Join(kept, " "), indexMap
}

func translatePositionIndex(strippedIdx int, indexMap []int) int {
	if strippedIdx < 0 || strippedIdx >= len(indexMap) {
		return -1
	}
	return indexMap[strippedIdx]
}

func buildContextSnippet(words []string, idx, radius int) string {
	if idx < 0 || idx >= len(words) {
		return ""
	}
	start := idx - radius
	if start < 0 {
		start = 0
	}
	end := idx + radius + 1
	if end > len(words) {
		end = len(words)
	}
	return strings.Join(words[start:end], " ")
}

func normalizeForMatch(s string) string {
	return strings.ToLower(strings.Trim(s, " \t\n.,!?;:\"'()"))
}

func grammarTextGroundedInTranscript(originalText, transcript string) bool {
	needle := normalizeForMatch(originalText)
	if needle == "" {
		return false
	}
	return strings.Contains(strings.ToLower(transcript), needle)
}

func clampSeverity(v int) int {
	if v < minSeverity || v > maxSeverity {
		return defaultSeverity
	}
	return v
}

func rawToCompactResult(raw *compactAnalysisResultRaw, originalTranscript string, indexMap []int) *compactAnalysisResult {
	result := &compactAnalysisResult{}
	result.Scores.Relevance = raw.Scores.Relevance
	result.Scores.Coherence = raw.Scores.Coherence
	result.Scores.Vocabulary = raw.Scores.Vocabulary
	result.Scores.Grammar = raw.Scores.Grammar
	result.Strength = raw.Strength
	result.Improvement = raw.Improvement
	result.Tip = raw.Tip

	originalWords := strings.Fields(originalTranscript)

	for _, gc := range raw.GrammarCorrections {
		origIdx := translatePositionIndex(gc.PositionIndex, indexMap)

		if origIdx == -1 {
			log.Printf("session analysis: dropping grammar correction with unresolvable position index %d", gc.PositionIndex)
			droppedGrammarFindings.Add(1)
			continue
		}
		if !grammarTextGroundedInTranscript(gc.OriginalText, originalTranscript) {
			log.Printf("session analysis: dropping grammar correction not found verbatim in transcript: %q", gc.OriginalText)
			droppedGrammarFindings.Add(1)
			continue
		}

		result.GrammarCorrections = append(result.GrammarCorrections, grammarCorrectionItem{
			OriginalText:   gc.OriginalText,
			CorrectedText:  gc.CorrectedText,
			Explanation:    gc.Explanation,
			ErrorType:      gc.ErrorType,
			PositionIndex:  origIdx,
			Severity:       clampSeverity(gc.Severity),
			ContextSnippet: buildContextSnippet(originalWords, origIdx, contextSnippetRadius),
		})
	}

	for _, vs := range raw.VocabularySuggestions {
		origIdx := translatePositionIndex(vs.PositionIndex, indexMap)

		if origIdx == -1 || origIdx >= len(originalWords) {
			log.Printf("session analysis: dropping vocabulary suggestion with unresolvable position index %d", vs.PositionIndex)
			droppedVocabFindings.Add(1)
			continue
		}
		if normalizeForMatch(originalWords[origIdx]) != normalizeForMatch(vs.OriginalWord) {
			log.Printf("session analysis: dropping vocabulary suggestion %q - position %d holds %q", vs.OriginalWord, origIdx, originalWords[origIdx])
			droppedVocabFindings.Add(1)
			continue
		}

		result.VocabularySuggestions = append(result.VocabularySuggestions, vocabularySuggestionItem{
			OriginalWord:   vs.OriginalWord,
			SuggestedWords: vs.SuggestedWords,
			Reason:         vs.Reason,
			ContextSnippet: buildContextSnippet(originalWords, origIdx, contextSnippetRadius),
			PositionIndex:  origIdx,
			Severity:       clampSeverity(vs.Severity),
		})
	}

	return result
}

func (s *SessionAnalysisService) buildOpenAIRequestBody(topicTitle, promptText string, wordCount, maxCompletionTokens int, useReasoning bool) map[string]any {
	userContent := fmt.Sprintf(
		"Topic: %s\n\nTranscript (%d words):\n%s\n\nThis response is %d words - keep the number of findings proportionate to that length; don't force findings that aren't clearly there just to fill the allowance.",
		topicTitle, wordCount, promptText, wordCount,
	)

	body := map[string]any{
		"model": s.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": compactSystemPrompt},
			{"role": "user", "content": userContent},
		},
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "speech_analysis",
				"strict": true,
				"schema": compactResultJSONSchema(),
			},
		},
		"max_completion_tokens": maxCompletionTokens,
	}

	if useReasoning {
		body["reasoning_effort"] = "low"
	} else {
		body["temperature"] = 0.2
	}

	return body
}

func isMaxTokensExceededError(statusCode int, body []byte) bool {
	if statusCode != http.StatusBadRequest {
		return false
	}

	var envelope struct {
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return false
	}

	msg := strings.ToLower(envelope.Error.Message)
	if strings.Contains(msg, "max_tokens") || strings.Contains(msg, "max_completion_tokens") || strings.Contains(msg, "output limit") {
		return true
	}

	return envelope.Error.Code == "max_tokens_exceeded" || envelope.Error.Code == "length_finish_reason"
}

func (s *SessionAnalysisService) callOpenAI(
	ctx context.Context,
	topicTitle string,
	transcriptText string,
) (*openAICallResult, error) {
	promptText, indexMap := stripFillerWordsForPrompt(transcriptText)
	wordCount := len(strings.Fields(promptText))

	useReasoning := supportsReasoningEffort(s.cfg.Model)

	maxCompletionTokens := initialStandardMaxTokens
	tokenCap := maxStandardMaxTokens
	if useReasoning {
		maxCompletionTokens = initialReasoningMaxTokens
		tokenCap = maxReasoningMaxTokens
	}

	baseURL := s.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	endpoint := strings.TrimSuffix(baseURL, "/") + "/chat/completions"

	const maxAttempts = 4

	var lastErr error
	var lastStatus int

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		body := s.buildOpenAIRequestBody(topicTitle, promptText, wordCount, maxCompletionTokens, useReasoning)

		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("call openai: %w", err)
			if attempt == maxAttempts || !s.waitBackoff(ctx, attempt) {
				return &openAICallResult{statusCode: lastStatus}, lastErr
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("read response: %w", readErr)
			if attempt == maxAttempts || !s.waitBackoff(ctx, attempt) {
				return &openAICallResult{statusCode: lastStatus}, lastErr
			}
			continue
		}

		lastStatus = resp.StatusCode

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))

			if isMaxTokensExceededError(resp.StatusCode, respBody) {
				next := minInt(maxCompletionTokens*tokenEscalationFactor, tokenCap)

				if next <= maxCompletionTokens || attempt == maxAttempts {
					maxTokensExceededExhausted.Add(1)
					log.Printf("session analysis: max_tokens exceeded at max_completion_tokens=%d (cap=%d), giving up", maxCompletionTokens, tokenCap)
					return &openAICallResult{statusCode: resp.StatusCode}, lastErr
				}

				maxTokensExceededRetries.Add(1)
				log.Printf("session analysis: max_tokens exceeded (400) at max_completion_tokens=%d, escalating to %d for retry", maxCompletionTokens, next)
				maxCompletionTokens = next

				if !s.waitBackoff(ctx, attempt) {
					return &openAICallResult{statusCode: resp.StatusCode}, ctx.Err()
				}
				continue
			}

			retryable := isRetryableOpenAIStatus(resp.StatusCode, respBody)
			if attempt == maxAttempts || !retryable {
				return &openAICallResult{statusCode: resp.StatusCode}, lastErr
			}
			if !s.waitBackoff(ctx, attempt) {
				return &openAICallResult{statusCode: resp.StatusCode}, ctx.Err()
			}
			continue
		}

		var envelope struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				PromptTokens        int64 `json:"prompt_tokens"`
				CompletionTokens    int64 `json:"completion_tokens"`
				PromptTokensDetails struct {
					CachedTokens int64 `json:"cached_tokens"`
				} `json:"prompt_tokens_details"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return &openAICallResult{statusCode: resp.StatusCode}, fmt.Errorf("parse openai envelope: %w", err)
		}
		if len(envelope.Choices) == 0 {
			return &openAICallResult{statusCode: resp.StatusCode}, fmt.Errorf("openai returned no choices")
		}

		choice := envelope.Choices[0]
		content := choice.Message.Content

		truncated := choice.FinishReason == "length"

		if strings.TrimSpace(content) == "" {
			err := fmt.Errorf(
				"openai returned empty content (finish_reason=%q, completion_tokens=%d, max_completion_tokens=%d)",
				choice.FinishReason, envelope.Usage.CompletionTokens, maxCompletionTokens,
			)
			if truncated && attempt < maxAttempts {
				lastErr = err
				next := minInt(maxCompletionTokens*tokenEscalationFactor, tokenCap)
				log.Printf("session analysis: empty content at max_completion_tokens=%d, escalating to %d for retry", maxCompletionTokens, next)
				maxCompletionTokens = next
				if !s.waitBackoff(ctx, attempt) {
					return &openAICallResult{statusCode: resp.StatusCode, finishReason: choice.FinishReason}, ctx.Err()
				}
				continue
			}
			return &openAICallResult{statusCode: resp.StatusCode, finishReason: choice.FinishReason}, err
		}

		var raw compactAnalysisResultRaw
		if err := json.Unmarshal([]byte(content), &raw); err != nil {
			if truncated && attempt < maxAttempts {
				lastErr = fmt.Errorf("parse openai result json (truncated, finish_reason=length): %w", err)
				next := minInt(maxCompletionTokens*tokenEscalationFactor, tokenCap)
				log.Printf("session analysis: truncated json at max_completion_tokens=%d, escalating to %d for retry", maxCompletionTokens, next)
				maxCompletionTokens = next
				if !s.waitBackoff(ctx, attempt) {
					return &openAICallResult{statusCode: resp.StatusCode, finishReason: choice.FinishReason}, ctx.Err()
				}
				continue
			}
			return &openAICallResult{statusCode: resp.StatusCode, finishReason: choice.FinishReason}, fmt.Errorf("parse openai result json: %w", err)
		}

		result := rawToCompactResult(&raw, transcriptText, indexMap)

		return &openAICallResult{
			result:       result,
			rawResponse:  string(respBody),
			statusCode:   resp.StatusCode,
			finishReason: choice.FinishReason,
			inputTokens:  envelope.Usage.PromptTokens,
			outputTokens: envelope.Usage.CompletionTokens,
			cachedTokens: envelope.Usage.PromptTokensDetails.CachedTokens,
		}, nil
	}

	return &openAICallResult{statusCode: lastStatus}, lastErr
}

func (s *SessionAnalysisService) waitBackoff(ctx context.Context, attempt int) bool {
	backoff := time.Duration(1<<uint(attempt)) * time.Second
	jitter := time.Duration(rand.Intn(500)) * time.Millisecond

	select {
	case <-ctx.Done():
		return false
	case <-time.After(backoff + jitter):
		return true
	}
}

func supportsReasoningEffort(model string) bool {
	return strings.HasPrefix(model, "gpt-5") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}

func isRetryableOpenAIStatus(statusCode int, body []byte) bool {
	if statusCode == http.StatusTooManyRequests {
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &envelope) == nil && envelope.Error.Code == "insufficient_quota" {
			return false
		}
		return true
	}
	return statusCode >= 500 && statusCode < 600
}

func sanitizeOpenAIFailure(statusCode int) string {
	switch {
	case statusCode == http.StatusTooManyRequests:
		return "Analysis is temporarily unavailable due to high demand. Please try again shortly."
	case statusCode >= 500:
		return "The analysis service is temporarily unavailable. Please try again shortly."
	default:
		return genericAnalysisFailureMessage
	}
}

func compactResultJSONSchema() map[string]any {
	grammarCorrectionSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ot": map[string]any{"type": "string"},
			"ct": map[string]any{"type": "string"},
			"ex": map[string]any{"type": "string"},
			"et": map[string]any{
				"type": "string",
				"enum": []string{"subject-verb", "tense", "article", "preposition", "word-order", "other"},
			},
			"pi": map[string]any{"type": "integer"},
			"sv": map[string]any{
				"type": "integer",
				"enum": []int{1, 2, 3},
			},
		},
		"required":             []string{"ot", "ct", "ex", "et", "pi", "sv"},
		"additionalProperties": false,
	}

	vocabularySuggestionSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ow": map[string]any{"type": "string"},
			"sw": map[string]any{
				"type":     "array",
				"items":    map[string]any{"type": "string"},
				"maxItems": 2,
			},
			"rs": map[string]any{"type": "string"},
			"pi": map[string]any{"type": "integer"},
			"sv": map[string]any{
				"type": "integer",
				"enum": []int{1, 2, 3},
			},
		},
		"required":             []string{"ow", "sw", "rs", "pi", "sv"},
		"additionalProperties": false,
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"sc": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"r": map[string]any{"type": "integer"},
					"c": map[string]any{"type": "integer"},
					"v": map[string]any{"type": "integer"},
					"g": map[string]any{"type": "integer"},
				},
				"required":             []string{"r", "c", "v", "g"},
				"additionalProperties": false,
			},
			"st": map[string]any{"type": "string"},
			"im": map[string]any{"type": "string"},
			"tp": map[string]any{"type": "string"},
			"gc": map[string]any{
				"type":     "array",
				"items":    grammarCorrectionSchema,
				"maxItems": maxGrammarCorrections,
			},
			"vs": map[string]any{
				"type":     "array",
				"items":    vocabularySuggestionSchema,
				"maxItems": maxVocabularySuggestions,
			},
		},
		"required":             []string{"sc", "st", "im", "tp", "gc", "vs"},
		"additionalProperties": false,
	}
}

func validateCompactResult(r *compactAnalysisResult) error {
	if r == nil {
		return fmt.Errorf("analysis result is nil")
	}

	scores := []struct {
		name  string
		value int
	}{
		{"relevance", r.Scores.Relevance},
		{"coherence", r.Scores.Coherence},
		{"vocabulary", r.Scores.Vocabulary},
		{"grammar", r.Scores.Grammar},
	}

	for _, sc := range scores {
		if sc.value < 0 || sc.value > 10 {
			return fmt.Errorf("%s must be between 0 and 10, got %d", sc.name, sc.value)
		}
	}

	if strings.TrimSpace(r.Strength) == "" {
		return fmt.Errorf("strength is required")
	}
	if strings.TrimSpace(r.Improvement) == "" {
		return fmt.Errorf("improvement is required")
	}
	if strings.TrimSpace(r.Tip) == "" {
		return fmt.Errorf("tip is required")
	}

	r.GrammarCorrections = filterAndDedupGrammar(r.GrammarCorrections)
	r.VocabularySuggestions = filterAndDedupVocab(r.VocabularySuggestions)

	if len(r.GrammarCorrections) > maxGrammarCorrections {
		r.GrammarCorrections = r.GrammarCorrections[:maxGrammarCorrections]
	}
	if len(r.VocabularySuggestions) > maxVocabularySuggestions {
		r.VocabularySuggestions = r.VocabularySuggestions[:maxVocabularySuggestions]
	}

	sort.SliceStable(r.GrammarCorrections, func(i, j int) bool {
		return r.GrammarCorrections[i].Severity > r.GrammarCorrections[j].Severity
	})
	sort.SliceStable(r.VocabularySuggestions, func(i, j int) bool {
		return r.VocabularySuggestions[i].Severity > r.VocabularySuggestions[j].Severity
	})

	if len(r.GrammarCorrections) > 0 && r.Scores.Grammar >= 9 {
		r.Scores.Grammar = 8
	}
	if len(r.VocabularySuggestions) > 0 && r.Scores.Vocabulary >= 9 {
		r.Scores.Vocabulary = 8
	}

	return nil
}

func filterAndDedupGrammar(items []grammarCorrectionItem) []grammarCorrectionItem {
	byPosition := make(map[int]grammarCorrectionItem, len(items))
	order := make([]int, 0, len(items))

	for _, gc := range items {
		if strings.TrimSpace(gc.OriginalText) == "" || strings.TrimSpace(gc.CorrectedText) == "" {
			droppedGrammarFindings.Add(1)
			continue
		}
		if _, ok := validGrammarErrorTypes[gc.ErrorType]; !ok {
			droppedGrammarFindings.Add(1)
			continue
		}
		if normalizeForMatch(gc.OriginalText) == normalizeForMatch(gc.CorrectedText) {
			droppedGrammarFindings.Add(1)
			continue
		}

		existing, seen := byPosition[gc.PositionIndex]
		if seen {
			droppedGrammarFindings.Add(1)
			if gc.Severity <= existing.Severity {
				continue
			}
		} else {
			order = append(order, gc.PositionIndex)
		}
		byPosition[gc.PositionIndex] = gc
	}

	result := make([]grammarCorrectionItem, 0, len(order))
	for _, pos := range order {
		result = append(result, byPosition[pos])
	}
	return result
}

func filterAndDedupVocab(items []vocabularySuggestionItem) []vocabularySuggestionItem {
	byPosition := make(map[int]vocabularySuggestionItem, len(items))
	order := make([]int, 0, len(items))

	for _, vs := range items {
		if strings.TrimSpace(vs.OriginalWord) == "" {
			droppedVocabFindings.Add(1)
			continue
		}

		cleanedWords := make([]string, 0, len(vs.SuggestedWords))
		for _, w := range vs.SuggestedWords {
			if strings.TrimSpace(w) == "" {
				continue
			}
			if normalizeForMatch(w) == normalizeForMatch(vs.OriginalWord) {
				continue
			}
			cleanedWords = append(cleanedWords, w)
		}
		if len(cleanedWords) == 0 {
			droppedVocabFindings.Add(1)
			continue
		}
		vs.SuggestedWords = cleanedWords

		existing, seen := byPosition[vs.PositionIndex]
		if seen {
			droppedVocabFindings.Add(1)
			if vs.Severity <= existing.Severity {
				continue
			}
		} else {
			order = append(order, vs.PositionIndex)
		}
		byPosition[vs.PositionIndex] = vs
	}

	result := make([]vocabularySuggestionItem, 0, len(order))
	for _, pos := range order {
		result = append(result, byPosition[pos])
	}
	return result
}

func computeDeliveryScore(wpm float64, fillerCount, wordCount int) float64 {
	const (
		idealMinWPM = 110.0
		idealMaxWPM = 160.0
	)

	paceScore := 100.0
	switch {
	case wpm <= 0:
		paceScore = 50.0
	case wpm < idealMinWPM:
		deficit := idealMinWPM - wpm
		paceScore = clampScore(100-deficit*1.2, 0, 100)
	case wpm > idealMaxWPM:
		excess := wpm - idealMaxWPM
		paceScore = clampScore(100-excess*1.0, 0, 100)
	}

	fillerRatio := 0.0
	if wordCount > 0 {
		fillerRatio = float64(fillerCount) / float64(wordCount)
	}
	fillerScore := clampScore(100-fillerRatio*400, 0, 100)

	return clampScore(0.6*paceScore+0.4*fillerScore, 0, 100)
}

func computeOverallScore(clarity, delivery, content, vocabulary, grammar float64) float64 {
	return clampScore(
		0.20*clarity+0.20*delivery+0.25*content+0.15*vocabulary+0.20*grammar,
		0, 100,
	)
}

func clampScore(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
