package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
)

var ErrUserStatsNotFound = errors.New("service: user stats not found")
var ErrUserSkillStatsNotFound = errors.New("service: user skill stats not found")
var ErrNotOnLeaderboard = errors.New("service: user not on leaderboard")

const (
	defaultLeaderboardLimit = 50
	maxLeaderboardLimit     = 100
	badgeAwardXP            = 25
	nonAnalyzedSessionXP    = 10
)

type GamificationService struct {
	gamificationRepo repository.GamificationRepository
	viewRepo         repository.ViewRepository
	notificationSvc  *NotificationService
}

func NewGamificationService(gamificationRepo repository.GamificationRepository, viewRepo repository.ViewRepository, notificationSvc *NotificationService) *GamificationService {
	return &GamificationService{gamificationRepo: gamificationRepo, viewRepo: viewRepo, notificationSvc: notificationSvc}
}

func (s *GamificationService) GetMyStats(ctx context.Context, userID string) (*models.UserStats, error) {
	stats, err := s.gamificationRepo.GetUserStats(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserStatsNotFound
		}
		return nil, fmt.Errorf("service: get my stats: %w", err)
	}
	return stats, nil
}

func (s *GamificationService) ListMyDailyActivity(ctx context.Context, userID string, limit int) ([]models.DailyActivity, error) {
	if limit <= 0 || limit > 365 {
		limit = 30
	}
	activity, err := s.gamificationRepo.ListDailyActivityForUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("service: list my daily activity: %w", err)
	}
	return activity, nil
}

func (s *GamificationService) ListMyXPTransactions(ctx context.Context, userID string, limit, offset int) ([]models.XPTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	txns, err := s.gamificationRepo.ListXPTransactionsForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: list my xp transactions: %w", err)
	}
	return txns, nil
}

func (s *GamificationService) GetMySkillStats(ctx context.Context, userID string) (*models.UserSkillStats, error) {
	stats, err := s.gamificationRepo.GetUserSkillStats(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserSkillStatsNotFound
		}
		return nil, fmt.Errorf("service: get my skill stats: %w", err)
	}
	return stats, nil
}

func (s *GamificationService) ListBadges(ctx context.Context) ([]models.Badge, error) {
	badges, err := s.gamificationRepo.ListActiveBadges(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list badges: %w", err)
	}
	return badges, nil
}

func (s *GamificationService) ListMyBadges(ctx context.Context, userID string) ([]models.UserBadge, error) {
	badges, err := s.gamificationRepo.ListUserBadges(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: list my badges: %w", err)
	}
	return badges, nil
}

// AwardSessionXP grants XP for a completed speaking session and updates the
// user's aggregate stats (total sessions, streak, average/best score, total
// XP) and today's daily activity row in the same logical operation. overallScore
// should be nil for sessions that skipped AI analysis (over quota) - those
// still earn a flat XP amount and still count toward streak/session totals,
// just without moving average_score/best_score.
//
// Idempotent per session: safe to call multiple times for the same sessionID
// (e.g. on reanalysis) - only the first call has any effect, since it's keyed
// off the same idempotency key as the XP transaction itself.
func (s *GamificationService) AwardSessionXP(ctx context.Context, userID string, sessionID uuid.UUID, overallScore *float64) (*models.UserStats, error) {
	xp := nonAnalyzedSessionXP
	if overallScore != nil {
		xp = int(math.Round(*overallScore))
		if xp < 1 {
			xp = 1
		}
	}

	txn := &models.XPTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		SessionID:      &sessionID,
		Amount:         xp,
		Reason:         models.XPReasonSessionComplete,
		IdempotencyKey: fmt.Sprintf("session_complete:%s", sessionID.String()),
	}

	created, err := s.gamificationRepo.CreateXPTransaction(ctx, txn)
	if err != nil {
		return nil, fmt.Errorf("service: award session xp: create xp transaction: %w", err)
	}

	if !created {
		// Already awarded for this session (e.g. a reanalysis) - don't
		// double-count stats or daily activity.
		stats, err := s.gamificationRepo.GetUserStats(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, nil
			}
			return nil, fmt.Errorf("service: award session xp: get user stats: %w", err)
		}
		return stats, nil
	}

	if err := s.gamificationRepo.UpsertDailyActivity(ctx, userID, txn.Amount); err != nil {
		return nil, fmt.Errorf("service: award session xp: upsert daily activity: %w", err)
	}

	stats, err := s.gamificationRepo.UpsertUserStatsForSession(ctx, userID, overallScore, txn.Amount)
	if err != nil {
		return nil, fmt.Errorf("service: award session xp: upsert user stats: %w", err)
	}

	return stats, nil
}

func (s *GamificationService) EvaluateAndAwardBadges(ctx context.Context, userID string, sessionID *uuid.UUID) ([]models.Badge, error) {
	stats, err := s.gamificationRepo.GetUserStats(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("service: evaluate badges: get user stats: %w", err)
	}

	badges, err := s.gamificationRepo.ListActiveBadges(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: evaluate badges: list badges: %w", err)
	}

	var awarded []models.Badge

	for _, b := range badges {
		if b.Manual() || !badgeCriteriaMet(b, stats) {
			continue
		}

		userBadge := &models.UserBadge{
			ID:        uuid.New(),
			UserID:    userID,
			BadgeID:   b.ID,
			SessionID: sessionID,
		}

		created, err := s.gamificationRepo.AwardBadge(ctx, userBadge)
		if err != nil {
			return awarded, fmt.Errorf("service: evaluate badges: award badge %s: %w", b.Slug, err)
		}
		if !created {
			continue
		}

		xpTxn := &models.XPTransaction{
			ID:             uuid.New(),
			UserID:         userID,
			SessionID:      sessionID,
			Amount:         badgeAwardXP,
			Reason:         models.XPReasonBadgeEarned,
			IdempotencyKey: fmt.Sprintf("badge_earned:%s:%s", userID, b.ID.String()),
		}
		xpCreated, err := s.gamificationRepo.CreateXPTransaction(ctx, xpTxn)
		if err != nil {
			return awarded, fmt.Errorf("service: evaluate badges: award xp for badge %s: %w", b.Slug, err)
		}
		if xpCreated {
			if err := s.gamificationRepo.AddUserXP(ctx, userID, badgeAwardXP); err != nil {
				return awarded, fmt.Errorf("service: evaluate badges: add xp for badge %s: %w", b.Slug, err)
			}
		}

		if _, err := s.notificationSvc.CreateNotification(
			ctx,
			userID,
			models.NotificationTypeBadgeEarned,
			models.NotificationChannelInApp,
			fmt.Sprintf("New badge: %s", b.Name),
			b.Description,
			models.JSONB("{}"),
		); err != nil {
			log.Printf("gamification: create badge notification failed for badge %s: %v", b.Slug, err)
		}

		awarded = append(awarded, b)
	}

	return awarded, nil
}

func badgeCriteriaMet(b models.Badge, stats *models.UserStats) bool {
	if b.CriteriaValue == nil {
		return false
	}
	threshold := *b.CriteriaValue

	switch b.CriteriaType {
	case models.BadgeCriteriaTypeStreakDays:
		return stats.CurrentStreakDays >= threshold
	case models.BadgeCriteriaTypeSessionCount:
		return stats.TotalSessions >= threshold
	case models.BadgeCriteriaTypeXPTotal:
		return stats.TotalXP >= int64(threshold)
	case models.BadgeCriteriaTypeScoreThreshold:
		return stats.BestScore >= float64(threshold)
	default:
		return false
	}
}

func (s *GamificationService) GetLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	limit, offset = normalizeLeaderboardPaging(limit, offset)
	entries, err := s.viewRepo.GetLeaderboard(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: get leaderboard: %w", err)
	}
	return entries, nil
}

func (s *GamificationService) GetCountryLeaderboard(ctx context.Context, countryCode string, limit, offset int) ([]models.CountryLeaderboardEntry, error) {
	limit, offset = normalizeLeaderboardPaging(limit, offset)
	entries, err := s.viewRepo.GetCountryLeaderboard(ctx, countryCode, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: get country leaderboard: %w", err)
	}
	return entries, nil
}

func (s *GamificationService) GetMyLeaderboardRank(ctx context.Context, userID string) (*models.LeaderboardEntry, error) {
	entry, err := s.viewRepo.GetMyLeaderboardRank(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotOnLeaderboard
		}
		return nil, fmt.Errorf("service: get my leaderboard rank: %w", err)
	}
	return entry, nil
}

func normalizeLeaderboardPaging(limit, offset int) (int, int) {
	if limit <= 0 || limit > maxLeaderboardLimit {
		limit = defaultLeaderboardLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
