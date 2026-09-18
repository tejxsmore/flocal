package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/repository"
	"flocal/internal/service"
)

type CompleteProfileRequest struct {
	Name        string  `json:"name" validate:"required"`
	Username    *string `json:"username"`
	Image       *string `json:"image"`
	CountryCode *string `json:"countryCode" validate:"omitempty,len=2"`
}

type CompletePreferencesRequest struct {
	PrimaryGoal         *models.OnboardingGoal      `json:"primaryGoal" validate:"omitempty"`
	DailyTimeCommitment *models.DailyTimeCommitment `json:"dailyTimeCommitment" validate:"omitempty"`
	FocusAreas          []models.FocusArea          `json:"focusAreas" validate:"omitempty"`
}

type OnboardingHandler struct {
	svc *service.OnboardingService
}

func NewOnboardingHandler(svc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{svc: svc}
}

func (h *OnboardingHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	auth := middleware.RequireAuth(authValidator, cookies)

	rg.POST("/profile", auth, h.CompleteProfile)
	rg.POST("/preferences", auth, h.CompletePreferences)
	rg.GET("/preferences", auth, h.GetPreferences)
}

func (h *OnboardingHandler) CompleteProfile(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req CompleteProfileRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	updated, err := h.svc.CompleteProfile(c.Request.Context(), user.ID, service.CompleteProfileInput{
		Name:        req.Name,
		Username:    req.Username,
		Image:       req.Image,
		CountryCode: req.CountryCode,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			middleware.Fail(c, http.StatusConflict, middleware.CodeBadRequest, "That username is already taken.")
		default:
			log.Printf("onboarding: complete profile failed: %v", err)
			middleware.Internal(c, "Could not complete onboarding.")
		}
		return
	}

	middleware.OK(c, updated)
}

func (h *OnboardingHandler) CompletePreferences(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req CompletePreferencesRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	prefs, areas, err := h.svc.CompletePreferences(c.Request.Context(), user.ID, service.CompletePreferencesInput{
		PrimaryGoal:         req.PrimaryGoal,
		DailyTimeCommitment: req.DailyTimeCommitment,
		FocusAreas:          req.FocusAreas,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoFocusAreas):
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Select at least one focus area.")
		default:
			log.Printf("onboarding: complete preferences failed: %v", err)
			middleware.Internal(c, "Could not save your preferences.")
		}
		return
	}

	middleware.OK(c, gin.H{
		"preferences": prefs,
		"focusAreas":  areas,
	})
}

func (h *OnboardingHandler) GetPreferences(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	prefs, areas, err := h.svc.GetPreferences(c.Request.Context(), user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			middleware.OK(c, gin.H{"preferences": nil, "focusAreas": []models.UserFocusArea{}})
			return
		}
		log.Printf("onboarding: get preferences failed: %v", err)
		middleware.Internal(c, "Could not load your preferences.")
		return
	}

	middleware.OK(c, gin.H{
		"preferences": prefs,
		"focusAreas":  areas,
	})
}
