package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/service"
)

type SetMasteredRequest struct {
	Mastered bool `json:"mastered"`
}

type VocabularyHandler struct {
	svc *service.VocabularyService
}

func NewVocabularyHandler(svc *service.VocabularyService) *VocabularyHandler {
	return &VocabularyHandler{svc: svc}
}

func (h *VocabularyHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("", middleware.RequireAuth(authValidator, cookies), h.ListMyVocabulary)
	rg.PATCH("/:wordId", middleware.RequireAuth(authValidator, cookies), h.SetMastered)
}

func (h *VocabularyHandler) ListMyVocabulary(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var mastered *bool
	if raw := c.Query("mastered"); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err != nil {
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "mastered must be true or false.")
			return
		}
		mastered = &val
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	entries, total, err := h.svc.ListMyVocabulary(c.Request.Context(), user.ID, mastered, limit, offset)
	if err != nil {
		log.Printf("vocabulary: list my vocabulary failed: %v", err)
		middleware.Internal(c, "Could not load vocabulary.")
		return
	}

	middleware.OK(c, gin.H{
		"words":  entries,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *VocabularyHandler) SetMastered(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	wordID, err := uuid.Parse(c.Param("wordId"))
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Invalid word id.")
		return
	}

	var req SetMasteredRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	updated, err := h.svc.SetMastered(c.Request.Context(), user.ID, wordID, req.Mastered)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserVocabularyNotFound):
			middleware.NotFound(c, "That word is not in your vocabulary yet.")
		default:
			log.Printf("vocabulary: set mastered failed: %v", err)
			middleware.Internal(c, "Could not update word.")
		}
		return
	}

	middleware.OK(c, updated)
}
