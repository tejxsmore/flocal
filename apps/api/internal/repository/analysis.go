package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type AnalysisRepository interface {
	CreateSessionAnalysis(ctx context.Context, a *models.SessionAnalysis) error
	GetSessionAnalysisBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.SessionAnalysis, error)
	CreateGrammarCorrections(ctx context.Context, corrections []models.GrammarCorrection) error
	CreateVocabularySuggestions(ctx context.Context, suggestions []models.VocabularySuggestion) error
	GetGrammarCorrectionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]models.GrammarCorrection, error)
	GetVocabularySuggestionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]models.VocabularySuggestion, error)
	CountAnalyzedSessionsForUser(ctx context.Context, userID string) (int64, error)
	CountAnalyzedSessionsForUserSince(ctx context.Context, userID string, since time.Time) (int64, error)
	CreateAIProviderLog(ctx context.Context, log *models.AIProviderLog) error
	DeleteForSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgAnalysisRepository struct {
	pool *pgxpool.Pool
}

func NewAnalysisRepository(pool *pgxpool.Pool) AnalysisRepository {
	return &pgAnalysisRepository{pool: pool}
}

func (r *pgAnalysisRepository) CreateSessionAnalysis(ctx context.Context, a *models.SessionAnalysis) error {
	const q = `
		insert into session_analysis (
			id, session_id, words_per_minute, speaking_duration_seconds,
			filler_word_count, filler_word_breakdown, filler_rate,
			pause_count, longest_pause_seconds, total_pause_seconds, pause_rate,
			unique_word_count, lexical_diversity, sentence_count, average_sentence_length,
			overall_score, clarity_score, delivery_score, content_score,
			vocabulary_score, grammar_score, strengths, improvements, coach_message,
			ai_provider, ai_model, ai_prompt_version, prompt_template_id, analysis_version,
			ai_temperature, ai_seed, request_id, created_at
		)
		values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,
			$25,$26,$27,$28,$29,$30,$31,$32,now()
		)
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		a.ID, a.SessionID, a.WordsPerMinute, a.SpeakingDurationSeconds,
		a.FillerWordCount, a.FillerWordBreakdown, a.FillerRate,
		a.PauseCount, a.LongestPauseSeconds, a.TotalPauseSeconds, a.PauseRate,
		a.UniqueWordCount, a.LexicalDiversity, a.SentenceCount, a.AverageSentenceLength,
		a.OverallScore, a.ClarityScore, a.DeliveryScore, a.ContentScore,
		a.VocabularyScore, a.GrammarScore, a.Strengths, a.Improvements, a.CoachMessage,
		a.AIProvider, a.AIModel, a.AIPromptVersion, a.PromptTemplateID, a.AnalysisVersion,
		a.AITemperature, a.AISeed, a.RequestID,
	).Scan(&a.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create session analysis: %w", err)
	}
	return nil
}

func (r *pgAnalysisRepository) GetSessionAnalysisBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.SessionAnalysis, error) {
	const q = `
		select id, session_id, words_per_minute, speaking_duration_seconds,
		       filler_word_count, filler_word_breakdown, filler_rate,
		       pause_count, longest_pause_seconds, total_pause_seconds, pause_rate,
		       unique_word_count, lexical_diversity, sentence_count, average_sentence_length,
		       overall_score, clarity_score, delivery_score, content_score,
		       vocabulary_score, grammar_score, strengths, improvements, coach_message,
		       ai_provider, ai_model, ai_prompt_version, prompt_template_id, analysis_version,
		       ai_temperature, ai_seed, request_id, created_at
		from session_analysis
		where session_id = $1`

	var a models.SessionAnalysis
	err := r.pool.QueryRow(ctx, q, sessionID).Scan(
		&a.ID, &a.SessionID, &a.WordsPerMinute, &a.SpeakingDurationSeconds,
		&a.FillerWordCount, &a.FillerWordBreakdown, &a.FillerRate,
		&a.PauseCount, &a.LongestPauseSeconds, &a.TotalPauseSeconds, &a.PauseRate,
		&a.UniqueWordCount, &a.LexicalDiversity, &a.SentenceCount, &a.AverageSentenceLength,
		&a.OverallScore, &a.ClarityScore, &a.DeliveryScore, &a.ContentScore,
		&a.VocabularyScore, &a.GrammarScore, &a.Strengths, &a.Improvements, &a.CoachMessage,
		&a.AIProvider, &a.AIModel, &a.AIPromptVersion, &a.PromptTemplateID, &a.AnalysisVersion,
		&a.AITemperature, &a.AISeed, &a.RequestID, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get session analysis: %w", err)
	}
	return &a, nil
}

func (r *pgAnalysisRepository) CreateGrammarCorrections(ctx context.Context, corrections []models.GrammarCorrection) error {
	if len(corrections) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	const q = `
		insert into grammar_corrections (
			id, session_id, transcript_id, original_text, corrected_text,
			explanation, error_type, context_snippet, start_char, end_char, created_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now())`
	for _, gc := range corrections {
		batch.Queue(q, gc.ID, gc.SessionID, gc.TranscriptID, gc.OriginalText, gc.CorrectedText, gc.Explanation, gc.ErrorType, gc.ContextSnippet, gc.StartChar, gc.EndChar)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range corrections {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("repository: create grammar corrections: %w", err)
		}
	}
	return nil
}

func (r *pgAnalysisRepository) GetGrammarCorrectionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]models.GrammarCorrection, error) {
	const q = `
		select id, session_id, transcript_id, original_text, corrected_text,
		       explanation, error_type, context_snippet, start_char, end_char, created_at
		from grammar_corrections
		where session_id = $1
		order by start_char asc nulls last, created_at asc`

	rows, err := r.pool.Query(ctx, q, sessionID)
	if err != nil {
		return nil, fmt.Errorf("repository: get grammar corrections: %w", err)
	}
	defer rows.Close()

	corrections := make([]models.GrammarCorrection, 0)
	for rows.Next() {
		var gc models.GrammarCorrection
		if err := rows.Scan(
			&gc.ID, &gc.SessionID, &gc.TranscriptID, &gc.OriginalText, &gc.CorrectedText,
			&gc.Explanation, &gc.ErrorType, &gc.ContextSnippet, &gc.StartChar, &gc.EndChar, &gc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan grammar correction: %w", err)
		}
		corrections = append(corrections, gc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: get grammar corrections: %w", err)
	}
	return corrections, nil
}

func (r *pgAnalysisRepository) CreateVocabularySuggestions(ctx context.Context, suggestions []models.VocabularySuggestion) error {
	if len(suggestions) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	const q = `
		insert into vocabulary_suggestions (
			id, session_id, transcript_id, original_word, suggested_words,
			reason, context_snippet, start_char, end_char, created_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())`
	for _, vs := range suggestions {
		batch.Queue(q, vs.ID, vs.SessionID, vs.TranscriptID, vs.OriginalWord, vs.SuggestedWords, vs.Reason, vs.ContextSnippet, vs.StartChar, vs.EndChar)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range suggestions {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("repository: create vocabulary suggestions: %w", err)
		}
	}
	return nil
}

func (r *pgAnalysisRepository) GetVocabularySuggestionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]models.VocabularySuggestion, error) {
	const q = `
		select id, session_id, transcript_id, original_word, suggested_words,
		       reason, context_snippet, start_char, end_char, created_at
		from vocabulary_suggestions
		where session_id = $1
		order by start_char asc nulls last, created_at asc`

	rows, err := r.pool.Query(ctx, q, sessionID)
	if err != nil {
		return nil, fmt.Errorf("repository: get vocabulary suggestions: %w", err)
	}
	defer rows.Close()

	suggestions := make([]models.VocabularySuggestion, 0)
	for rows.Next() {
		var vs models.VocabularySuggestion
		if err := rows.Scan(
			&vs.ID, &vs.SessionID, &vs.TranscriptID, &vs.OriginalWord, &vs.SuggestedWords,
			&vs.Reason, &vs.ContextSnippet, &vs.StartChar, &vs.EndChar, &vs.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan vocabulary suggestion: %w", err)
		}
		suggestions = append(suggestions, vs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: get vocabulary suggestions: %w", err)
	}
	return suggestions, nil
}

func (r *pgAnalysisRepository) CountAnalyzedSessionsForUser(ctx context.Context, userID string) (int64, error) {
	const q = `
		select count(*)
		from session_analysis sa
		join speaking_sessions s on s.id = sa.session_id
		where s.user_id = $1`

	var count int64
	err := r.pool.QueryRow(ctx, q, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("repository: count analyzed sessions for user: %w", err)
	}
	return count, nil
}

func (r *pgAnalysisRepository) CountAnalyzedSessionsForUserSince(ctx context.Context, userID string, since time.Time) (int64, error) {
	const q = `
		select count(*)
		from session_analysis sa
		join speaking_sessions s on s.id = sa.session_id
		where s.user_id = $1
		  and sa.created_at >= $2`

	var count int64
	err := r.pool.QueryRow(ctx, q, userID, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("repository: count analyzed sessions for user since: %w", err)
	}
	return count, nil
}

func (r *pgAnalysisRepository) CreateAIProviderLog(ctx context.Context, log *models.AIProviderLog) error {
	const q = `
		insert into ai_provider_logs (
			id, session_id, processing_attempt_id, source, operation,
			input_tokens, output_tokens, cached_tokens, latency_ms, cost_subunits,
			raw_response, created_at, purge_after
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),$12)
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		log.ID, log.SessionID, log.ProcessingAttemptID, log.Source, log.Operation,
		log.InputTokens, log.OutputTokens, log.CachedTokens, log.LatencyMs, log.CostSubunits,
		log.RawResponse, log.PurgeAfter,
	).Scan(&log.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create ai provider log: %w", err)
	}
	return nil
}

func (r *pgAnalysisRepository) DeleteForSession(ctx context.Context, sessionID uuid.UUID) error {
	batch := &pgx.Batch{}
	batch.Queue(`delete from grammar_corrections where session_id = $1`, sessionID)
	batch.Queue(`delete from vocabulary_suggestions where session_id = $1`, sessionID)
	batch.Queue(`delete from session_analysis where session_id = $1`, sessionID)

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < 3; i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("repository: delete analysis for session: %w", err)
		}
	}

	return nil
}

func (r *pgAnalysisRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	batch := &pgx.Batch{}
	batch.Queue(`
		delete from grammar_corrections
		where session_id in (select id from speaking_sessions where user_id = $1)`, userID)
	batch.Queue(`
		delete from vocabulary_suggestions
		where session_id in (select id from speaking_sessions where user_id = $1)`, userID)
	batch.Queue(`
		delete from session_analysis
		where session_id in (select id from speaking_sessions where user_id = $1)`, userID)

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < 3; i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("repository: delete analysis for user: %w", err)
		}
	}

	return nil
}
