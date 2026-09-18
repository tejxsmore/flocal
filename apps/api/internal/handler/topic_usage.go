package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/service"
)

type TopicUsageHandler struct {
	svc *service.TopicUsageService
}

func NewTopicUsageHandler(svc *service.TopicUsageService) *TopicUsageHandler {
	return &TopicUsageHandler{svc: svc}
}

func (h *TopicUsageHandler) RegisterRoutes(topicGroup *gin.RouterGroup, statsGroup *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	topicGroup.GET("/:id/usage", middleware.RequireAuth(authValidator, cookies), h.GetTopicUsage)
	statsGroup.GET("/top", middleware.RequireAuth(authValidator, cookies), h.ListTopTopics)
}

func (h *TopicUsageHandler) GetTopicUsage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, 400, middleware.CodeBadRequest, "id must be a valid UUID.")
		return
	}

	stats, err := h.svc.GetTopicUsage(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTopicNotFound):
			middleware.NotFound(c, "That topic does not exist.")
		default:
			log.Printf("topic usage: get failed: %v", err)
			middleware.Internal(c, "Could not load topic usage.")
		}
		return
	}

	middleware.OK(c, stats)
}

func (h *TopicUsageHandler) ListTopTopics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))

	stats, err := h.svc.ListTopTopics(c.Request.Context(), limit)
	if err != nil {
		log.Printf("topic usage: list failed: %v", err)
		middleware.Internal(c, "Could not load topic usage.")
		return
	}

	middleware.OK(c, stats)
}
