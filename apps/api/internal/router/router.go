package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"flocal/internal/config"
	"flocal/internal/database"
	"flocal/internal/handler"
	"flocal/internal/middleware"
	"flocal/internal/repository"
	"flocal/internal/service"
	"flocal/internal/utils"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
)

func colorize(color, s string) string {
	return color + s + ansiReset
}

func New(cfg *config.Config, db *database.DB, redisClient *redis.Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(requestLogger())
	r.Use(middleware.ErrorHandler())

	if len(cfg.App.TrustedProxies) > 0 {
		_ = r.SetTrustedProxies(cfg.App.TrustedProxies)
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	cookies := middleware.CookieConfig{
		Domain: cfg.Auth.CookieDomain,
		Secure: cfg.Auth.CookieSecure,
	}

	r.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		redisCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := redisClient.Ping(redisCtx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.Env,
			"time":   time.Now().UTC(),
		})
	})

	log.Printf(
		"aws s3 config: region=%q bucket=%q uploadPrefix=%q",
		cfg.AWS.Region,
		cfg.AWS.S3Bucket,
		cfg.AWS.S3UploadPrefix,
	)

	s3Ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s3Client, err := utils.NewS3Client(s3Ctx, cfg.AWS)
	if err != nil {
		log.Fatalf("router: init s3 client: %v", err)
	}

	var geoIP *utils.GeoIP
	if cfg.GeoIP.Enabled {
		geoIP, err = utils.NewGeoIP(cfg.GeoIP.DBPath)
		if err != nil {
			log.Printf("router: geoip init failed, continuing without country resolution: %v", err)
			geoIP = nil
		}
	}

	authRepo := repository.NewAuthRepository(db.Pool)
	onboardingRepo := repository.NewOnboardingRepository(db.Pool)
	auditRepo := repository.NewAuditRepository(db.Pool)
	topicRepo := repository.NewTopicRepository(db.Pool)
	sessionRepo := repository.NewSessionRepository(db.Pool)
	transcriptRepo := repository.NewTranscriptRepository(db.Pool)
	analysisRepo := repository.NewAnalysisRepository(db.Pool)
	viewRepo := repository.NewViewRepository(db.Pool)
	billingRepo := repository.NewBillingRepository(db.Pool)
	gamificationRepo := repository.NewGamificationRepository(db.Pool)
	vocabularyRepo := repository.NewVocabularyRepository(db.Pool)
	notificationRepo := repository.NewNotificationRepository(db.Pool)
	challengeRepo := repository.NewChallengeRepository(db.Pool)
	sessionStreamSvc := service.NewSessionStreamService(sessionRepo, transcriptRepo, s3Client, cfg.AWS.S3Bucket, cfg.AWS.S3UploadPrefix)

	emailSvc := service.NewResendEmailService(cfg.Resend)
	oauthSvc := service.NewOAuthService(cfg)
	auditSvc := service.NewAuditService(auditRepo)
	vocabularySvc := service.NewVocabularyService(vocabularyRepo)
	notificationSvc := service.NewNotificationService(notificationRepo)
	challengeSvc := service.NewChallengeService(challengeRepo)
	authSvc := service.NewAuthService(cfg, authRepo, sessionRepo, gamificationRepo, billingRepo, transcriptRepo, analysisRepo, vocabularyRepo, notificationRepo, challengeRepo, auditSvc, emailSvc, oauthSvc, s3Client, cfg.AWS.S3Bucket, cfg.AWS.Region)
	onboardingSvc := service.NewOnboardingService(onboardingRepo, authRepo, auditSvc)
	topicSvc := service.NewTopicService(topicRepo)
	sessionSvc := service.NewSessionService(sessionRepo, topicRepo)
	// gamificationSvc is constructed before sessionAnalysisSvc because
	// AnalyzeSession awards XP and evaluates badges as part of completing
	// a session, so sessionAnalysisSvc depends on it.
	gamificationSvc := service.NewGamificationService(gamificationRepo, viewRepo, notificationSvc)
	sessionAnalysisSvc := service.NewSessionAnalysisService(cfg.OpenAI, sessionRepo, transcriptRepo, analysisRepo, topicRepo, viewRepo, vocabularySvc, gamificationSvc)
	sessionReportSvc := service.NewSessionReportService(viewRepo, analysisRepo, s3Client, cfg.AWS.S3Bucket)
	planSvc := service.NewPlanService(viewRepo)
	topicUsageSvc := service.NewTopicUsageService(viewRepo)
	dodoClient := service.NewDodoClient(cfg.Dodo.APIKey, cfg.Dodo.WebhookSecret, cfg.Dodo.TestMode)
	billingSvc := service.NewBillingService(billingRepo, dodoClient, cfg.Dodo.ReturnURL)

	authLimiter := middleware.PerRoute(redisClient, 10, 5, 10*time.Minute, true)
	emailLimiter := middleware.NewIPRateLimiter(redisClient, 3, 2, 30*time.Minute, true)

	authHandler := handler.NewAuthHandler(cfg, authSvc, emailLimiter, geoIP)
	onboardingHandler := handler.NewOnboardingHandler(onboardingSvc)
	auditHandler := handler.NewAuditHandler(auditSvc)
	topicHandler := handler.NewTopicHandler(topicSvc)
	sessionHandler := handler.NewSpeakingSessionHandler(sessionSvc, sessionAnalysisSvc)
	sessionStreamHandler := handler.NewSessionStreamHandler(cfg, sessionSvc, sessionStreamSvc, sessionAnalysisSvc)
	sessionReportHandler := handler.NewSessionReportHandler(sessionReportSvc)
	planHandler := handler.NewPlanHandler(planSvc)
	topicUsageHandler := handler.NewTopicUsageHandler(topicUsageSvc)
	billingHandler := handler.NewBillingHandler(billingSvc)
	gamificationHandler := handler.NewGamificationHandler(gamificationSvc)
	vocabularyHandler := handler.NewVocabularyHandler(vocabularySvc)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)
	challengeHandler := handler.NewChallengeHandler(challengeSvc)

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		authHandler.RegisterRoutes(authGroup, authSvc, authLimiter, cookies)

		onboardingGroup := v1.Group("/onboarding")
		onboardingHandler.RegisterRoutes(onboardingGroup, authSvc, cookies)

		auditGroup := v1.Group("/audit-logs")
		auditHandler.RegisterRoutes(auditGroup, authSvc, cookies)

		topicGroup := v1.Group("/topics")
		topicHandler.RegisterRoutes(topicGroup, authSvc, cookies)

		topicStatsGroup := v1.Group("/topic-stats")
		topicUsageHandler.RegisterRoutes(topicGroup, topicStatsGroup, authSvc, cookies)

		sessionGroup := v1.Group("/sessions")
		sessionHandler.RegisterRoutes(sessionGroup, authSvc, cookies)
		sessionReportHandler.RegisterRoutes(sessionGroup, authSvc, cookies)

		sessionStreamHandler.RegisterRoutes(sessionGroup, authSvc, cookies)

		planGroup := v1.Group("/plan")
		planHandler.RegisterRoutes(planGroup, authSvc, cookies)

		billingGroup := v1.Group("/billing")
		billingHandler.RegisterRoutes(billingGroup, authSvc, cookies)

		gamificationGroup := v1.Group("/gamification")
		gamificationHandler.RegisterRoutes(gamificationGroup, authSvc, cookies)

		vocabularyGroup := v1.Group("/vocabulary")
		vocabularyHandler.RegisterRoutes(vocabularyGroup, authSvc, cookies)

		notificationGroup := v1.Group("/notifications")
		notificationHandler.RegisterRoutes(notificationGroup, authSvc, cookies)

		dailyChallengeGroup := v1.Group("/daily-challenge")
		challengeHandler.RegisterRoutes(dailyChallengeGroup, authSvc, cookies)
	}

	return r
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path += "?" + raw
		}

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		statusColor := ansiGreen

		switch {
		case status >= 500:
			statusColor = ansiRed
		case status >= 400:
			statusColor = ansiYellow
		case status >= 300:
			statusColor = ansiCyan
		}

		line := fmt.Sprintf(
			"%s %-6s %s %s",
			colorize(statusColor+ansiBold, fmt.Sprintf("%d", status)),
			colorize(ansiBlue, c.Request.Method),
			path,
			colorize(ansiDim, latency.String()),
		)

		if len(c.Errors) > 0 {
			fmt.Printf(
				"%s %s | %s\n",
				colorize(ansiRed+ansiBold, "[FAIL]"),
				line,
				c.Errors.String(),
			)
			return
		}

		fmt.Println(line)
	}
}
