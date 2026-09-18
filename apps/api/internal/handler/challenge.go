package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
)

type CompleteDailyChallengeRequest struct {
	SessionID *string `json:"sessionId"`
}

type ChallengeHandler struct {
	svc *service.ChallengeService
}

func NewChallengeHandler(svc *service.ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{svc: svc}
}

func (h *ChallengeHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("/today", middleware.RequireAuth(authValidator, cookies), h.GetToday)
	rg.POST("/:id/complete", middleware.RequireAuth(authValidator, cookies), h.Complete)
	rg.GET("/me", middleware.RequireAuth(authValidator, cookies), h.ListMyCompletions)
}

func (h *ChallengeHandler) GetToday(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}

	language := c.DefaultQuery("language", "en")

	var difficulty models.TopicDifficulty
	if raw := c.Query("difficulty"); raw != "" {
		difficulty = models.TopicDifficulty(raw)
		if !difficulty.Valid() {
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "difficulty must be one of easy, medium, hard.")
			return
		}
	} else {
		difficulty = models.DifficultyMedium
	}

	challenge, err := h.svc.GetToday(c.Request.Context(), language, difficulty)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDailyChallengeNotFound):
			middleware.NotFound(c, "No daily challenge is available for today yet.")
		default:
			log.Printf("challenge: get today failed: %v", err)
			middleware.Internal(c, "Could not load today's challenge.")
		}
		return
	}

	middleware.OK(c, challenge)
}

func (h *ChallengeHandler) Complete(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Invalid challenge id.")
		return
	}

	var req CompleteDailyChallengeRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	var sessionID *uuid.UUID
	if req.SessionID != nil {
		id, err := uuid.Parse(*req.SessionID)
		if err != nil {
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Invalid session id.")
			return
		}
		sessionID = &id
	}

	completion, created, err := h.svc.CompleteChallenge(c.Request.Context(), user.ID, challengeID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDailyChallengeNotFound):
			middleware.NotFound(c, "That daily challenge does not exist.")
		default:
			log.Printf("challenge: complete failed: %v", err)
			middleware.Internal(c, "Could not complete challenge.")
		}
		return
	}

	middleware.OK(c, gin.H{
		"completion":     completion,
		"newlyCompleted": created,
	})
}

func (h *ChallengeHandler) ListMyCompletions(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	completions, err := h.svc.ListMyCompletions(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		log.Printf("challenge: list my completions failed: %v", err)
		middleware.Internal(c, "Could not load your completions.")
		return
	}

	middleware.OK(c, completions)
}
