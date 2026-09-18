package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("", middleware.RequireAuth(authValidator, cookies), h.List)
}

func (h *AuditHandler) List(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if user.Role != models.UserRoleAdmin && user.Role != models.UserRoleModerator {
		middleware.Fail(c, http.StatusForbidden, middleware.CodeForbidden, "You do not have access to audit logs.")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	entityType := c.Query("entityType")
	entityID := c.Query("entityId")
	actorUserID := c.Query("actorUserId")

	var (
		logs []models.AuditLog
		err  error
	)

	switch {
	case entityType != "" && entityID != "":
		logs, err = h.svc.ListForEntity(c.Request.Context(), entityType, entityID, limit, offset)
	case actorUserID != "":
		logs, err = h.svc.ListForActor(c.Request.Context(), actorUserID, limit, offset)
	default:
		logs, err = h.svc.List(c.Request.Context(), limit, offset)
	}

	if err != nil {
		log.Printf("audit: list failed: %v", err)
		middleware.Internal(c, "Could not load audit logs.")
		return
	}

	middleware.OK(c, logs)
}
