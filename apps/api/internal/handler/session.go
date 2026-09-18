package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
)

type CreateSpeakingSessionRequest struct {
	TopicID         string  `json:"topicId" validate:"required"`
	PrepTimeSeconds *int    `json:"prepTimeSeconds"`
	DebateStance    *string `json:"debateStance" validate:"omitempty,oneof=for against"`
}

type SpeakingSessionHandler struct {
	svc         *service.SessionService
	analysisSvc *service.SessionAnalysisService
}

func NewSpeakingSessionHandler(svc *service.SessionService, analysisSvc *service.SessionAnalysisService) *SpeakingSessionHandler {
	return &SpeakingSessionHandler{svc: svc, analysisSvc: analysisSvc}
}

func (h *SpeakingSessionHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.POST("", middleware.RequireAuth(authValidator, cookies), h.CreateSession)
	rg.GET("/:id", middleware.RequireAuth(authValidator, cookies), h.GetSession)
	rg.DELETE("/:id", middleware.RequireAuth(authValidator, cookies), h.DeleteSession)
	rg.GET("", middleware.RequireAuth(authValidator, cookies), h.ListSessions)
	rg.GET("/history", middleware.RequireAuth(authValidator, cookies), h.ListSessionHistory)
	rg.GET("/quota", middleware.RequireAuth(authValidator, cookies), h.GetAIQuota)
	rg.POST("/:id/reanalyze", middleware.RequireAuth(authValidator, cookies), h.ReanalyzeSession)
}

func (h *SpeakingSessionHandler) CreateSession(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req CreateSpeakingSessionRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	topicID, err := uuid.Parse(req.TopicID)
	if err != nil {
		middleware.Fail(c, 400, middleware.CodeBadRequest, "topicId must be a valid UUID.")
		return
	}

	session, err := h.svc.CreateSession(c.Request.Context(), user, topicID, req.PrepTimeSeconds, req.DebateStance)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTopicNotFound):
			middleware.NotFound(c, "That topic does not exist.")
		case errors.Is(err, service.ErrInvalidPrepTime):
			middleware.Fail(c, 400, middleware.CodeBadRequest, "prepTimeSeconds must be one of 0, 300, 600, 900.")
		case errors.Is(err, service.ErrDebateStanceNotApplicable):
			middleware.Fail(c, 400, middleware.CodeBadRequest, "debateStance can only be set for debate topics.")
		default:
			log.Printf("session: create failed: %v", err)
			middleware.Internal(c, "Could not create session.")
		}
		return
	}

	middleware.OK(c, session)
}

func (h *SpeakingSessionHandler) GetSession(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, 400, middleware.CodeBadRequest, "id must be a valid UUID.")
		return
	}

	session, err := h.svc.GetSession(c.Request.Context(), id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound), errors.Is(err, service.ErrSessionAccessDenied):
			middleware.NotFound(c, "That session does not exist.")
		default:
			log.Printf("session: get failed: %v", err)
			middleware.Internal(c, "Could not load session.")
		}
		return
	}

	middleware.OK(c, session)
}

func (h *SpeakingSessionHandler) ListSessions(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	sessions, err := h.svc.ListSessions(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		log.Printf("session: list failed: %v", err)
		middleware.Internal(c, "Could not load sessions.")
		return
	}

	middleware.OK(c, sessions)
}

func (h *SpeakingSessionHandler) ListSessionHistory(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	filter := models.SessionHistoryFilter{
		Search: c.Query("q"),
		Sort:   c.DefaultQuery("sort", "newest"),
	}

	if status := c.Query("status"); status != "" {
		filter.Status = strings.Split(status, ",")
	}

	if format := c.Query("format"); format != "" {
		filter.Format = strings.Split(format, ",")
	}

	if from := c.Query("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			middleware.Fail(c, 400, middleware.CodeBadRequest, "from must be a valid RFC3339 timestamp.")
			return
		}
		filter.From = &t
	}

	if to := c.Query("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			middleware.Fail(c, 400, middleware.CodeBadRequest, "to must be a valid RFC3339 timestamp.")
			return
		}
		filter.To = &t
	}

	items, err := h.svc.ListMyHistory(c.Request.Context(), user.ID, filter, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFilter):
			middleware.Fail(c, 400, middleware.CodeBadRequest, "Invalid filter parameter.")
		default:
			log.Printf("session: list history failed: %v", err)
			middleware.Internal(c, "Could not load session history.")
		}
		return
	}

	middleware.OK(c, items)
}

func (h *SpeakingSessionHandler) ReanalyzeSession(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, 400, middleware.CodeBadRequest, "id must be a valid UUID.")
		return
	}

	session, err := h.analysisSvc.EnsureReanalyzable(c.Request.Context(), id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound), errors.Is(err, service.ErrSessionAccessDenied):
			middleware.NotFound(c, "That session does not exist.")
		case errors.Is(err, service.ErrSessionNotEligibleForReanalysis):
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "This session cannot be reanalyzed right now.")
		default:
			log.Printf("session: reanalyze check failed: %v", err)
			middleware.Internal(c, "Could not reanalyze session.")
		}
		return
	}

	go h.runReanalysis(session.ID)

	middleware.OK(c, gin.H{"status": "processing"})
}

func (h *SpeakingSessionHandler) runReanalysis(sessionID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := h.analysisSvc.AnalyzeSession(ctx, sessionID); err != nil {
		log.Printf("session reanalysis: failed for session %s: %v", sessionID, err)
	}
}

func (h *SpeakingSessionHandler) GetAIQuota(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	quota, err := h.analysisSvc.AIAnalysisQuota(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("session: get ai quota failed: %v", err)
		middleware.Internal(c, "Could not load AI quota.")
		return
	}

	middleware.OK(c, quota)
}

func (h *SpeakingSessionHandler) DeleteSession(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, 400, middleware.CodeBadRequest, "id must be a valid UUID.")
		return
	}

	if err := h.svc.DeleteSession(c.Request.Context(), id, user.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			middleware.NotFound(c, "That session does not exist.")
		default:
			log.Printf("session: delete failed: %v", err)
			middleware.Internal(c, "Could not delete session.")
		}
		return
	}

	middleware.OK(c, gin.H{"status": "deleted"})
}
