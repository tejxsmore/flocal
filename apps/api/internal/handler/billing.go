package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"flocal/internal/middleware"
	"flocal/internal/service"
)

type BillingHandler struct {
	svc *service.BillingService
}

func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

func (h *BillingHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("/plans", h.ListPlans)
	rg.GET("/subscription", middleware.RequireAuth(authValidator, cookies), h.GetMySubscription)
	rg.GET("/payments", middleware.RequireAuth(authValidator, cookies), h.ListMyPayments)
	rg.POST("/checkout", middleware.RequireAuth(authValidator, cookies), h.CreateCheckout)
	rg.POST("/subscription/cancel", middleware.RequireAuth(authValidator, cookies), h.CancelSubscription)
	rg.POST("/subscription/resume", middleware.RequireAuth(authValidator, cookies), h.ResumeSubscription)
	rg.POST("/subscription/change-plan", middleware.RequireAuth(authValidator, cookies), h.ChangePlan)
	rg.POST("/subscription/update-payment-method", middleware.RequireAuth(authValidator, cookies), h.UpdatePaymentMethod)
	rg.POST("/webhook", h.HandleWebhook)
}

func (h *BillingHandler) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		log.Printf("billing: list plans failed: %v", err)
		middleware.Internal(c, "Could not load plans.")
		return
	}
	middleware.OK(c, plans)
}

func (h *BillingHandler) GetMySubscription(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	sub, err := h.svc.GetMySubscription(c.Request.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoActiveSubscription):
			middleware.NotFound(c, "No active subscription found.")
		default:
			log.Printf("billing: get my subscription failed: %v", err)
			middleware.Internal(c, "Could not load subscription.")
		}
		return
	}

	middleware.OK(c, sub)
}

func (h *BillingHandler) ListMyPayments(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	payments, err := h.svc.ListMyPayments(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		log.Printf("billing: list my payments failed: %v", err)
		middleware.Internal(c, "Could not load payments.")
		return
	}

	middleware.OK(c, payments)
}

type createCheckoutRequest struct {
	PlanSlug string `json:"planSlug" binding:"required"`
	Country  string `json:"country"`
}

func (h *BillingHandler) CreateCheckout(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req createCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "planSlug is required.")
		return
	}

	session, err := h.svc.CreateCheckoutSession(c.Request.Context(), user.ID, user.Email, user.Name, req.PlanSlug, req.Country)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlanNotFound):
			middleware.NotFound(c, "Plan not found.")
		case errors.Is(err, service.ErrCannotCheckoutFree):
			middleware.BadRequest(c, "The free plan does not require checkout.")
		case errors.Is(err, service.ErrPlanNotPurchasable):
			middleware.BadRequest(c, "This plan is not currently purchasable.")
		case errors.Is(err, service.ErrAlreadySubscribed):
			middleware.BadRequest(c, "You already have an active subscription.")
		default:
			log.Printf("billing: create checkout failed: %v", err)
			middleware.Internal(c, "Could not start checkout.")
		}
		return
	}

	middleware.OK(c, session)
}

func (h *BillingHandler) CancelSubscription(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if err := h.svc.CancelMySubscription(c.Request.Context(), user.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrNoActiveSubscription):
			middleware.NotFound(c, "No active subscription found.")
		default:
			log.Printf("billing: cancel subscription failed: %v", err)
			middleware.Internal(c, "Could not cancel subscription.")
		}
		return
	}

	middleware.OK(c, gin.H{"canceled": true})
}

func (h *BillingHandler) ResumeSubscription(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if err := h.svc.ResumeMySubscription(c.Request.Context(), user.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrNoActiveSubscription):
			middleware.NotFound(c, "No active subscription found.")
		default:
			log.Printf("billing: resume subscription failed: %v", err)
			middleware.Internal(c, "Could not resume subscription.")
		}
		return
	}

	middleware.OK(c, gin.H{"resumed": true})
}

type changePlanRequest struct {
	PlanSlug string `json:"planSlug" binding:"required"`
}

func (h *BillingHandler) ChangePlan(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req changePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "planSlug is required.")
		return
	}

	if err := h.svc.ChangeMyPlan(c.Request.Context(), user.ID, req.PlanSlug); err != nil {
		switch {
		case errors.Is(err, service.ErrNoActiveSubscription):
			middleware.NotFound(c, "No active subscription found.")
		case errors.Is(err, service.ErrPlanNotFound):
			middleware.NotFound(c, "Plan not found.")
		case errors.Is(err, service.ErrPlanNotPurchasable):
			middleware.BadRequest(c, "This plan is not currently purchasable.")
		case errors.Is(err, service.ErrSamePlan):
			middleware.BadRequest(c, "You're already on this plan.")
		case errors.Is(err, service.ErrSubscriptionNotActive):
			middleware.BadRequest(c, "This subscription can't be changed right now.")
		case errors.Is(err, service.ErrPendingPlanChange):
			middleware.BadRequest(c, "A plan change is already in progress.")
		default:
			log.Printf("billing: change plan failed: %v", err)
			middleware.Internal(c, "Could not change plan.")
		}
		return
	}

	middleware.OK(c, gin.H{"changing": true})
}

func (h *BillingHandler) UpdatePaymentMethod(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	result, err := h.svc.UpdateMyPaymentMethod(c.Request.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoActiveSubscription):
			middleware.NotFound(c, "No active subscription found.")
		default:
			log.Printf("billing: update payment method failed: %v", err)
			middleware.Internal(c, "Could not start payment method update.")
		}
		return
	}

	middleware.OK(c, result)
}

func (h *BillingHandler) HandleWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		middleware.BadRequest(c, "Could not read webhook body.")
		return
	}

	webhookID := c.GetHeader("webhook-id")
	timestamp := c.GetHeader("webhook-timestamp")
	signature := c.GetHeader("webhook-signature")

	if webhookID == "" || timestamp == "" || signature == "" {
		middleware.BadRequest(c, "Missing webhook signature headers.")
		return
	}

	if err := h.svc.HandleWebhook(c.Request.Context(), webhookID, timestamp, signature, body); err != nil {
		if errors.Is(err, service.ErrWebhookSignature) {
			middleware.BadRequest(c, "Invalid webhook signature.")
			return
		}
		log.Printf("billing: handle webhook failed: %v", err)
		middleware.Internal(c, "Could not process webhook.")
		return
	}

	c.Status(200)
}
