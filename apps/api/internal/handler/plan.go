package handler

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"flocal/internal/middleware"
	"flocal/internal/service"
)

type PlanHandler struct {
	svc *service.PlanService
}

func NewPlanHandler(svc *service.PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("", middleware.RequireAuth(authValidator, cookies), h.GetCurrentPlan)
}

func (h *PlanHandler) GetCurrentPlan(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	plan, err := h.svc.GetCurrentPlan(c.Request.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlanNotFound):
			middleware.NotFound(c, "No active plan found.")
		default:
			log.Printf("plan: get current plan failed: %v", err)
			middleware.Internal(c, "Could not load plan.")
		}
		return
	}

	middleware.OK(c, plan)
}
