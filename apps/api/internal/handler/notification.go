package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
)

type UpdateNotificationPreferencesRequest struct {
	DailyReminderEnabled bool    `json:"dailyReminderEnabled"`
	StreakRiskEnabled    bool    `json:"streakRiskEnabled"`
	SessionReadyEnabled  bool    `json:"sessionReadyEnabled"`
	ReminderTime         *string `json:"reminderTime"`
}

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("/preferences", middleware.RequireAuth(authValidator, cookies), h.GetPreferences)
	rg.PATCH("/preferences", middleware.RequireAuth(authValidator, cookies), h.UpdatePreferences)

	rg.GET("", middleware.RequireAuth(authValidator, cookies), h.ListNotifications)
	rg.GET("/unread-count", middleware.RequireAuth(authValidator, cookies), h.UnreadCount)
	rg.POST("/:id/read", middleware.RequireAuth(authValidator, cookies), h.MarkRead)
	rg.POST("/read-all", middleware.RequireAuth(authValidator, cookies), h.MarkAllRead)
}

func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	prefs, err := h.svc.GetMyPreferences(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("notification: get my preferences failed: %v", err)
		middleware.Internal(c, "Could not load notification preferences.")
		return
	}

	middleware.OK(c, prefs)
}

func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req UpdateNotificationPreferencesRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	var reminderTime *models.TimeOfDay
	if req.ReminderTime != nil {
		parsed, err := parseReminderTime(*req.ReminderTime)
		if err != nil {
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Invalid reminder time, expected HH:MM.")
			return
		}
		reminderTime = &parsed
	}

	updated, err := h.svc.UpdateMyPreferences(c.Request.Context(), user.ID, service.UpdatePreferencesInput{
		DailyReminderEnabled: req.DailyReminderEnabled,
		StreakRiskEnabled:    req.StreakRiskEnabled,
		SessionReadyEnabled:  req.SessionReadyEnabled,
		ReminderTime:         reminderTime,
	})
	if err != nil {
		log.Printf("notification: update my preferences failed: %v", err)
		middleware.Internal(c, "Could not update notification preferences.")
		return
	}

	middleware.OK(c, updated)
}

func parseReminderTime(value string) (models.TimeOfDay, error) {
	layouts := []string{"15:04", "15:04:05"}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return models.NewTimeOfDay(parsed), nil
		}
	}

	return models.TimeOfDay{}, fmt.Errorf("invalid reminder time %q", value)
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	unreadOnly, _ := strconv.ParseBool(c.Query("unreadOnly"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	notifications, err := h.svc.ListMyNotifications(c.Request.Context(), user.ID, unreadOnly, limit, offset)
	if err != nil {
		log.Printf("notification: list my notifications failed: %v", err)
		middleware.Internal(c, "Could not load notifications.")
		return
	}

	middleware.OK(c, notifications)
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	count, err := h.svc.GetUnreadCount(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("notification: get unread count failed: %v", err)
		middleware.Internal(c, "Could not load unread count.")
		return
	}

	middleware.OK(c, gin.H{"unreadCount": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Invalid notification id.")
		return
	}

	if err := h.svc.MarkRead(c.Request.Context(), user.ID, id); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationNotFound):
			middleware.NotFound(c, "That notification does not exist.")
		default:
			log.Printf("notification: mark read failed: %v", err)
			middleware.Internal(c, "Could not mark notification as read.")
		}
		return
	}

	middleware.OK(c, gin.H{"message": "Notification marked as read."})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if err := h.svc.MarkAllRead(c.Request.Context(), user.ID); err != nil {
		log.Printf("notification: mark all read failed: %v", err)
		middleware.Internal(c, "Could not mark notifications as read.")
		return
	}

	middleware.OK(c, gin.H{"message": "All notifications marked as read."})
}
