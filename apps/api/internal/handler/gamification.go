package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"flocal/internal/middleware"
	"flocal/internal/service"
)

type GamificationHandler struct {
	svc *service.GamificationService
}

func NewGamificationHandler(svc *service.GamificationService) *GamificationHandler {
	return &GamificationHandler{svc: svc}
}

func (h *GamificationHandler) RegisterRoutes(rg *gin.RouterGroup, authValidator middleware.SessionValidator, cookies middleware.CookieConfig) {
	rg.GET("/stats", middleware.RequireAuth(authValidator, cookies), h.GetMyStats)
	rg.GET("/skills", middleware.RequireAuth(authValidator, cookies), h.GetMySkillStats)
	rg.GET("/activity", middleware.RequireAuth(authValidator, cookies), h.ListMyDailyActivity)
	rg.GET("/xp", middleware.RequireAuth(authValidator, cookies), h.ListMyXPTransactions)
	rg.GET("/badges", middleware.RequireAuth(authValidator, cookies), h.ListBadges)
	rg.GET("/badges/me", middleware.RequireAuth(authValidator, cookies), h.ListMyBadges)
	rg.GET("/leaderboard", middleware.RequireAuth(authValidator, cookies), h.GetLeaderboard)
	rg.GET("/leaderboard/country/:code", middleware.RequireAuth(authValidator, cookies), h.GetCountryLeaderboard)
	rg.GET("/leaderboard/me", middleware.RequireAuth(authValidator, cookies), h.GetMyLeaderboardRank)
}

func (h *GamificationHandler) GetMyStats(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	stats, err := h.svc.GetMyStats(c.Request.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserStatsNotFound):
			middleware.NotFound(c, "No stats found yet.")
		default:
			log.Printf("gamification: get my stats failed: %v", err)
			middleware.Internal(c, "Could not load stats.")
		}
		return
	}

	middleware.OK(c, stats)
}

func (h *GamificationHandler) GetMySkillStats(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	stats, err := h.svc.GetMySkillStats(c.Request.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserSkillStatsNotFound):
			middleware.NotFound(c, "No skill stats found yet.")
		default:
			log.Printf("gamification: get my skill stats failed: %v", err)
			middleware.Internal(c, "Could not load skill stats.")
		}
		return
	}

	middleware.OK(c, stats)
}

func (h *GamificationHandler) ListMyDailyActivity(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))

	activity, err := h.svc.ListMyDailyActivity(c.Request.Context(), user.ID, limit)
	if err != nil {
		log.Printf("gamification: list my daily activity failed: %v", err)
		middleware.Internal(c, "Could not load activity.")
		return
	}

	middleware.OK(c, activity)
}

func (h *GamificationHandler) ListMyXPTransactions(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	txns, err := h.svc.ListMyXPTransactions(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		log.Printf("gamification: list my xp transactions failed: %v", err)
		middleware.Internal(c, "Could not load XP transactions.")
		return
	}

	middleware.OK(c, txns)
}

func (h *GamificationHandler) ListBadges(c *gin.Context) {
	badges, err := h.svc.ListBadges(c.Request.Context())
	if err != nil {
		log.Printf("gamification: list badges failed: %v", err)
		middleware.Internal(c, "Could not load badges.")
		return
	}

	middleware.OK(c, badges)
}

func (h *GamificationHandler) ListMyBadges(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	badges, err := h.svc.ListMyBadges(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("gamification: list my badges failed: %v", err)
		middleware.Internal(c, "Could not load your badges.")
		return
	}

	middleware.OK(c, badges)
}

func (h *GamificationHandler) GetLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	entries, err := h.svc.GetLeaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		log.Printf("gamification: get leaderboard failed: %v", err)
		middleware.Internal(c, "Could not load leaderboard.")
		return
	}

	middleware.OK(c, entries)
}

func (h *GamificationHandler) GetCountryLeaderboard(c *gin.Context) {
	countryCode := c.Param("code")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	entries, err := h.svc.GetCountryLeaderboard(c.Request.Context(), countryCode, limit, offset)
	if err != nil {
		log.Printf("gamification: get country leaderboard failed: %v", err)
		middleware.Internal(c, "Could not load leaderboard.")
		return
	}

	middleware.OK(c, entries)
}

func (h *GamificationHandler) GetMyLeaderboardRank(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	entry, err := h.svc.GetMyLeaderboardRank(c.Request.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotOnLeaderboard):
			middleware.NotFound(c, "You're not on the leaderboard yet.")
		default:
			log.Printf("gamification: get my leaderboard rank failed: %v", err)
			middleware.Internal(c, "Could not load your rank.")
		}
		return
	}

	middleware.OK(c, entry)
}
