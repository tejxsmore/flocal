package handler

import (
	"errors"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
)

type TopicHandler struct {
	svc *service.TopicService
}

func NewTopicHandler(svc *service.TopicService) *TopicHandler {
	return &TopicHandler{svc: svc}
}

func (h *TopicHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("/spin", middleware.RequireAuth(authValidator, cookies), h.SpinRandomTopic)
	rg.GET("/categories", middleware.RequireAuth(authValidator, cookies), h.ListCategories)
}

func (h *TopicHandler) SpinRandomTopic(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var categoryID *uuid.UUID
	if raw := strings.TrimSpace(c.Query("categoryId")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			middleware.Fail(c, 400, middleware.CodeBadRequest, "categoryId must be a valid UUID.")
			return
		}
		categoryID = &id
	}

	var format *models.TopicFormat
	if raw := strings.TrimSpace(c.Query("format")); raw != "" {
		f := models.TopicFormat(raw)
		if !f.Valid() {
			middleware.Fail(c, 400, middleware.CodeBadRequest, "format must be one of word, quote, debate, situation, story_starter, image.")
			return
		}
		format = &f
	}

	topic, err := h.svc.SpinRandomTopic(c.Request.Context(), user.ID, categoryID, format)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCategoryNotFound):
			middleware.NotFound(c, "That category does not exist.")
		case errors.Is(err, service.ErrNoTopicsAvailable):
			middleware.NotFound(c, "No topics are available right now.")
		default:
			log.Printf("topic: spin failed: %v", err)
			middleware.Internal(c, "Could not select a topic.")
		}
		return
	}

	middleware.OK(c, topic)
}

func (h *TopicHandler) ListCategories(c *gin.Context) {
	categories, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		log.Printf("topic: list categories failed: %v", err)
		middleware.Internal(c, "Could not load categories.")
		return
	}

	middleware.OK(c, categories)
}
