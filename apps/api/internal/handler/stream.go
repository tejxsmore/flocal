package handler

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"flocal/internal/config"
	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
)

type SessionStreamHandler struct {
	cfg         *config.Config
	sessionSvc  *service.SessionService
	streamSvc   *service.SessionStreamService
	analysisSvc *service.SessionAnalysisService
	upgrader    websocket.Upgrader
}

func NewSessionStreamHandler(
	cfg *config.Config,
	sessionSvc *service.SessionService,
	streamSvc *service.SessionStreamService,
	analysisSvc *service.SessionAnalysisService,
) *SessionStreamHandler {
	allowed := make(map[string]bool, len(cfg.App.AllowedOrigins))

	for _, origin := range cfg.App.AllowedOrigins {
		allowed[origin] = true
	}

	return &SessionStreamHandler{
		cfg:         cfg,
		sessionSvc:  sessionSvc,
		streamSvc:   streamSvc,
		analysisSvc: analysisSvc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return allowed[r.Header.Get("Origin")]
			},
		},
	}
}

func (h *SessionStreamHandler) RegisterRoutes(
	rg *gin.RouterGroup,
	authValidator middleware.SessionValidator,
	cookies middleware.CookieConfig,
) {
	rg.GET(
		"/:id/stream",
		middleware.RequireAuth(authValidator, cookies),
		h.Stream,
	)
}

func (h *SessionStreamHandler) Stream(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(
			c,
			http.StatusBadRequest,
			middleware.CodeBadRequest,
			"id must be a valid UUID.",
		)
		return
	}

	session, err := h.sessionSvc.GetSession(
		c.Request.Context(),
		sessionID,
		user.ID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound),
			errors.Is(err, service.ErrSessionAccessDenied):
			middleware.NotFound(c, "That session does not exist.")
		default:
			log.Printf(
				"session stream: get session failed: %v",
				err,
			)
			middleware.Internal(c, "Could not load session.")
		}
		return
	}

	if !session.Pending() {
		middleware.Fail(
			c,
			http.StatusBadRequest,
			middleware.CodeBadRequest,
			"This session cannot be streamed to.",
		)
		return
	}

	aiEligible, err := h.analysisSvc.AIAnalysisEligible(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf(
			"session stream: check ai eligibility failed: %v",
			err,
		)
		aiEligible = true
	}

	clientConn, err := h.upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)
	if err != nil {
		log.Printf(
			"session stream: upgrade failed: %v",
			err,
		)
		return
	}
	defer clientConn.Close()

	if err := h.streamSvc.StartStream(
		c.Request.Context(),
		session,
	); err != nil {
		log.Printf(
			"session stream: start failed: %v",
			err,
		)

		_ = clientConn.WriteJSON(gin.H{
			"type":    "error",
			"message": "This session cannot be streamed to.",
		})

		return
	}

	if !aiEligible {
		h.streamPracticeOnly(clientConn, session)
		return
	}

	h.streamWithTranscription(c.Request.Context(), clientConn, session)
}

func (h *SessionStreamHandler) streamWithTranscription(
	ctx context.Context,
	clientConn *websocket.Conn,
	session *models.SpeakingSession,
) {
	dgClient, err := service.ConnectDeepgramStream(
		ctx,
		h.cfg.Deepgram,
	)
	if err != nil {
		log.Printf(
			"session stream: deepgram connect failed: %v",
			err,
		)

		_ = clientConn.WriteJSON(gin.H{
			"type":    "error",
			"message": "Could not start transcription.",
		})

		return
	}

	var audioBuf bytes.Buffer
	var transcriptBuf strings.Builder
	var allWords []service.DeepgramWord

	var confidenceSum float64
	var confidenceCount int

	done := make(chan struct{})

	go func() {
		defer close(done)

		for result := range dgClient.Results() {
			if !result.IsFinal {
				if err := clientConn.WriteJSON(gin.H{
					"type":       "interim",
					"transcript": result.Transcript,
				}); err != nil {
					return
				}

				continue
			}

			transcriptBuf.WriteString(result.Transcript)
			transcriptBuf.WriteByte(' ')

			allWords = append(allWords, result.Words...)

			fillerCount := 0

			for _, word := range result.Words {
				confidenceSum += word.Confidence
				confidenceCount++

				if service.IsFillerWord(word.Word) {
					fillerCount++
				}
			}

			if err := clientConn.WriteJSON(gin.H{
				"type":        "final",
				"transcript":  result.Transcript,
				"fillerCount": fillerCount,
			}); err != nil {
				return
			}
		}
	}()

	for {
		messageType, data, err := clientConn.ReadMessage()
		if err != nil {
			break
		}

		if messageType != websocket.BinaryMessage {
			continue
		}

		if len(data) == 0 {
			continue
		}

		if _, err := audioBuf.Write(data); err != nil {
			log.Printf(
				"session stream: buffer audio failed: %v",
				err,
			)
			break
		}

		if err := dgClient.SendAudio(data); err != nil {
			log.Printf(
				"session stream: forward to deepgram failed: %v",
				err,
			)
			break
		}
	}

	if err := dgClient.Close(); err != nil {
		log.Printf(
			"session stream: close deepgram failed: %v",
			err,
		)
	}

	<-done

	rawText := strings.TrimSpace(
		transcriptBuf.String(),
	)

	if rawText == "" {
		h.markSessionFailed(
			session.ID,
			"No speech detected.",
		)
		return
	}

	avgConfidence := 0.0

	if confidenceCount > 0 {
		avgConfidence =
			confidenceSum / float64(confidenceCount)
	}

	finalizedTranscript := service.FinalizedTranscript{
		RawText:       rawText,
		WordCount:     len(strings.Fields(rawText)),
		AvgConfidence: avgConfidence,
		Words:         allWords,
		Language:      h.cfg.Deepgram.Language,
	}

	finalizeCtx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	if err := h.streamSvc.FinalizeStream(
		finalizeCtx,
		session,
		audioBuf.Bytes(),
		finalizedTranscript,
	); err != nil {
		log.Printf(
			"session stream: finalize failed: %v",
			err,
		)

		h.markSessionFailed(
			session.ID,
			"Could not finalize session.",
		)

		return
	}

	go h.analyzeSession(session.ID)
}

func (h *SessionStreamHandler) streamPracticeOnly(
	clientConn *websocket.Conn,
	session *models.SpeakingSession,
) {
	_ = clientConn.WriteJSON(gin.H{
		"type":    "info",
		"message": "practice_mode",
	})

	var audioBuf bytes.Buffer

	for {
		messageType, data, err := clientConn.ReadMessage()
		if err != nil {
			break
		}

		if messageType != websocket.BinaryMessage {
			continue
		}

		if len(data) == 0 {
			continue
		}

		if _, err := audioBuf.Write(data); err != nil {
			log.Printf(
				"session stream: buffer audio failed: %v",
				err,
			)
			break
		}
	}

	if audioBuf.Len() == 0 {
		h.markSessionFailed(
			session.ID,
			"No audio recorded.",
		)
		return
	}

	finalizeCtx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	if err := h.streamSvc.FinalizePracticeStream(
		finalizeCtx,
		session,
		audioBuf.Bytes(),
	); err != nil {
		log.Printf(
			"session stream: finalize practice stream failed: %v",
			err,
		)

		h.markSessionFailed(
			session.ID,
			"Could not finalize session.",
		)

		return
	}
}

func (h *SessionStreamHandler) markSessionFailed(
	sessionID uuid.UUID,
	reason string,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := h.sessionSvc.MarkSessionFailed(
		ctx,
		sessionID,
		reason,
	); err != nil {
		log.Printf(
			"session stream: mark session %s failed: %v",
			sessionID,
			err,
		)
	}
}

func (h *SessionStreamHandler) analyzeSession(
	sessionID uuid.UUID,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		90*time.Second,
	)
	defer cancel()

	if err := h.analysisSvc.AnalyzeSession(
		ctx,
		sessionID,
	); err != nil {
		log.Printf(
			"session analysis: failed for session %s: %v",
			sessionID,
			err,
		)
	}
}
