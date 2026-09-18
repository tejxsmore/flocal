package handler

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flocal/internal/middleware"
	"flocal/internal/service"
)

type SessionReportHandler struct {
	svc *service.SessionReportService
}

func NewSessionReportHandler(svc *service.SessionReportService) *SessionReportHandler {
	return &SessionReportHandler{svc: svc}
}

func (h *SessionReportHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("/:id/report", middleware.RequireAuth(authValidator, cookies), h.GetReport)
}

func (h *SessionReportHandler) GetReport(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.Fail(c, 400, middleware.CodeBadRequest, "id must be a valid UUID.")
		return
	}

	report, err := h.svc.GetSessionReport(c.Request.Context(), id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionReportNotFound):
			middleware.NotFound(c, "That session does not exist.")
		default:
			log.Printf("session report: get failed: %v", err)
			middleware.Internal(c, "Could not load session report.")
		}
		return
	}

	middleware.OK(c, report)
}
