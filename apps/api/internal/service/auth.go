package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/oauth2"

	"flocal/internal/config"
	"flocal/internal/models"
	"flocal/internal/repository"
	"flocal/internal/utils"
)

var (
	ErrInvalidOrExpiredLink       = errors.New("service: magic link is invalid or has expired")
	ErrInvalidSession             = errors.New("service: session is invalid or has expired")
	ErrAccountInactive            = errors.New("service: account is inactive")
	ErrProviderNotConfigured      = errors.New("service: oauth provider is not configured")
	ErrProviderAlreadyLinked      = errors.New("service: provider is already linked to a different account")
	ErrUsernameTaken              = errors.New("service: username is already taken")
	ErrEmailTaken                 = errors.New("service: email is already in use")
	ErrAvatarTooLarge             = errors.New("service: avatar file is too large")
	ErrAvatarInvalidType          = errors.New("service: avatar must be a jpeg, png, or webp image")
	ErrSessionNotFoundForUser     = errors.New("service: session not found for user")
	ErrEmailNotVerifiedByProvider = errors.New("service: oauth provider did not return a verified email")
)

const (
	providerGoogle     = "google"
	providerGitHub     = "github"
	providerCredential = "credential"
	maxAvatarBytes     = 5 * 1024 * 1024
)

var allowedAvatarTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

type AuthService struct {
	cfg              *config.Config
	repo             repository.AuthRepository
	sessionRepo      repository.SessionRepository
	gamificationRepo repository.GamificationRepository
	billingRepo      repository.BillingRepository
	transcriptRepo   repository.TranscriptRepository
	analysisRepo     repository.AnalysisRepository
	vocabularyRepo   repository.VocabularyRepository
	notificationRepo repository.NotificationRepository
	challengeRepo    repository.ChallengeRepository
	auditSvc         *AuditService
	email            EmailService
	oauth            *OAuthService
	s3Client         *s3.Client
	s3Bucket         string
	s3Region         string
}

func NewAuthService(
	cfg *config.Config,
	repo repository.AuthRepository,
	sessionRepo repository.SessionRepository,
	gamificationRepo repository.GamificationRepository,
	billingRepo repository.BillingRepository,
	transcriptRepo repository.TranscriptRepository,
	analysisRepo repository.AnalysisRepository,
	vocabularyRepo repository.VocabularyRepository,
	notificationRepo repository.NotificationRepository,
	challengeRepo repository.ChallengeRepository,
	auditSvc *AuditService,
	email EmailService,
	oauth *OAuthService,
	s3Client *s3.Client,
	s3Bucket string,
	s3Region string,
) *AuthService {
	return &AuthService{
		cfg:              cfg,
		repo:             repo,
		sessionRepo:      sessionRepo,
		gamificationRepo: gamificationRepo,
		billingRepo:      billingRepo,
		transcriptRepo:   transcriptRepo,
		analysisRepo:     analysisRepo,
		vocabularyRepo:   vocabularyRepo,
		notificationRepo: notificationRepo,
		challengeRepo:    challengeRepo,
		auditSvc:         auditSvc,
		email:            email,
		oauth:            oauth,
		s3Client:         s3Client,
		s3Bucket:         s3Bucket,
		s3Region:         s3Region,
	}
}

func (s *AuthService) logAudit(ctx context.Context, actorUserID *string, action, entityType, entityID string, oldValue, newValue *models.JSONB) {
	if s.auditSvc == nil {
		return
	}
	source := "api"
	if err := s.auditSvc.Record(ctx, actorUserID, action, entityType, entityID, oldValue, newValue, &source); err != nil {
		log.Printf("auth: audit log failed: %v", err)
	}
}

type IssuedSession struct {
	User      *models.User
	RawToken  string
	ExpiresAt time.Time
}

type LinkedAccount struct {
	Provider  string `json:"provider"`
	Connected bool   `json:"connected"`
}

func (s *AuthService) GoogleEnabled() bool { return s.oauth.GoogleEnabled() }
func (s *AuthService) GitHubEnabled() bool { return s.oauth.GitHubEnabled() }

func (s *AuthService) GoogleAuthURL(state string) (string, error) {
	if !s.oauth.GoogleEnabled() {
		return "", ErrProviderNotConfigured
	}
	return s.oauth.AuthCodeURL(providerGoogle, state)
}

func (s *AuthService) GitHubAuthURL(state string) (string, error) {
	if !s.oauth.GitHubEnabled() {
		return "", ErrProviderNotConfigured
	}
	return s.oauth.AuthCodeURL(providerGitHub, state)
}

func (s *AuthService) CompleteGoogleLogin(ctx context.Context, code, ip, userAgent string, countryCode *string) (*IssuedSession, error) {
	return s.completeOAuthLogin(ctx, providerGoogle, code, ip, userAgent, countryCode)
}

func (s *AuthService) CompleteGitHubLogin(ctx context.Context, code, ip, userAgent string, countryCode *string) (*IssuedSession, error) {
	return s.completeOAuthLogin(ctx, providerGitHub, code, ip, userAgent, countryCode)
}

func (s *AuthService) completeOAuthLogin(ctx context.Context, provider, code, ip, userAgent string, countryCode *string) (*IssuedSession, error) {
	tok, err := s.oauth.Exchange(ctx, provider, code)
	if err != nil {
		return nil, fmt.Errorf("service: %s oauth exchange: %w", provider, err)
	}

	profile, err := s.oauth.FetchProfile(ctx, provider, tok)
	if err != nil {
		return nil, fmt.Errorf("service: %s fetch profile: %w", provider, err)
	}

	user, err := s.findOrCreateUserForOAuth(ctx, provider, profile, tok)
	if err != nil {
		return nil, fmt.Errorf("service: resolve %s user: %w", provider, err)
	}
	if !user.Active() {
		return nil, ErrAccountInactive
	}

	return s.issueSession(ctx, user, ip, userAgent, countryCode)
}

func (s *AuthService) ListLinkedAccounts(ctx context.Context, userID string) ([]LinkedAccount, error) {
	accounts, err := s.repo.ListAccountsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: list linked accounts: %w", err)
	}

	linked := make(map[string]bool, len(accounts))
	for _, a := range accounts {
		linked[a.ProviderID] = true
	}

	return []LinkedAccount{
		{Provider: providerGoogle, Connected: linked[providerGoogle]},
		{Provider: providerGitHub, Connected: linked[providerGitHub]},
	}, nil
}

func (s *AuthService) LinkGoogleAccount(ctx context.Context, userID, code string) (*models.User, error) {
	return s.linkOAuthAccount(ctx, userID, providerGoogle, code)
}

func (s *AuthService) LinkGitHubAccount(ctx context.Context, userID, code string) (*models.User, error) {
	return s.linkOAuthAccount(ctx, userID, providerGitHub, code)
}

func (s *AuthService) linkOAuthAccount(ctx context.Context, userID, provider, code string) (*models.User, error) {
	tok, err := s.oauth.Exchange(ctx, provider, code)
	if err != nil {
		return nil, fmt.Errorf("service: %s oauth exchange: %w", provider, err)
	}

	profile, err := s.oauth.FetchProfile(ctx, provider, tok)
	if err != nil {
		return nil, fmt.Errorf("service: %s fetch profile: %w", provider, err)
	}

	existing, err := s.repo.GetAccountByProvider(ctx, provider, profile.ProviderUserID)
	switch {
	case err == nil:
		if existing.UserID != userID {
			return nil, ErrProviderAlreadyLinked
		}
		existing.AccessToken = &tok.AccessToken
		if tok.RefreshToken != "" {
			existing.RefreshToken = &tok.RefreshToken
		}
		if !tok.Expiry.IsZero() {
			existing.AccessTokenExpiresAt = &tok.Expiry
		}
		if err := s.repo.UpdateAccountTokens(ctx, existing); err != nil {
			return nil, fmt.Errorf("service: update %s account tokens: %w", provider, err)
		}

	case errors.Is(err, repository.ErrNotFound):
		accountID, err := utils.NewID()
		if err != nil {
			return nil, err
		}
		newAccount := s.buildOAuthAccount(accountID, provider, userID, profile, tok)
		if err := s.repo.CreateAccount(ctx, newAccount); err != nil {
			if errors.Is(err, repository.ErrConflict) {
				return nil, ErrProviderAlreadyLinked
			}
			return nil, fmt.Errorf("service: link %s account: %w", provider, err)
		}
		s.logAudit(ctx, &userID, "account.linked", "user", userID, nil, toJSONB(map[string]any{"provider": provider}))

	default:
		return nil, fmt.Errorf("service: lookup linked account: %w", err)
	}

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) buildOAuthAccount(accountID, provider, userID string, profile *OAuthProfile, tok *oauth2.Token) *models.Account {
	account := &models.Account{
		ID:         accountID,
		AccountID:  profile.ProviderUserID,
		ProviderID: provider,
		UserID:     userID,
	}

	if tok.AccessToken != "" {
		account.AccessToken = &tok.AccessToken
	}
	if tok.RefreshToken != "" {
		account.RefreshToken = &tok.RefreshToken
	}
	if !tok.Expiry.IsZero() {
		account.AccessTokenExpiresAt = &tok.Expiry
	}

	return account
}

func (s *AuthService) findOrCreateUserForOAuth(ctx context.Context, provider string, profile *OAuthProfile, tok *oauth2.Token) (*models.User, error) {
	account, err := s.repo.GetAccountByProvider(ctx, provider, profile.ProviderUserID)
	switch {
	case err == nil:
		user, err := s.repo.GetUserByID(ctx, account.UserID)
		if err != nil {
			return nil, fmt.Errorf("load user for linked account: %w", err)
		}
		if profile.EmailVerified && !user.EmailVerified {
			if err := s.repo.MarkEmailVerified(ctx, user.ID); err != nil {
				return nil, err
			}
			user.EmailVerified = true
		}
		return user, nil

	case errors.Is(err, repository.ErrNotFound):
		email := normalizeEmail(profile.Email)
		user, err := s.repo.GetUserByEmail(ctx, email)
		switch {
		case err == nil:
			if !profile.EmailVerified {
				return nil, ErrEmailNotVerifiedByProvider
			}

		case errors.Is(err, repository.ErrNotFound):
			user, err = s.createUser(ctx, email, profile.Name, profile.AvatarURL, profile.EmailVerified)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("lookup user by email: %w", err)
		}

		accountID, err := utils.NewID()
		if err != nil {
			return nil, err
		}
		newAccount := s.buildOAuthAccount(accountID, provider, user.ID, profile, tok)
		if err := s.repo.CreateAccount(ctx, newAccount); err != nil {
			return nil, fmt.Errorf("link %s account: %w", provider, err)
		}

		if profile.EmailVerified && !user.EmailVerified {
			if err := s.repo.MarkEmailVerified(ctx, user.ID); err != nil {
				return nil, err
			}
			user.EmailVerified = true
		}

		return user, nil

	default:
		return nil, fmt.Errorf("lookup linked account: %w", err)
	}
}

func (s *AuthService) RequestMagicLink(ctx context.Context, rawEmail string) error {
	email := normalizeEmail(rawEmail)

	rawToken, err := utils.NewToken(32)
	if err != nil {
		return fmt.Errorf("generate magic link token: %w", err)
	}
	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.MagicLinkSigningKey)

	id, err := utils.NewID()
	if err != nil {
		return err
	}

	if err := s.repo.DeleteVerificationsByIdentifier(ctx, email); err != nil {
		return fmt.Errorf("clear previous magic links: %w", err)
	}

	verification := &models.Verification{
		ID:         id,
		Identifier: email,
		ValueHash:  hash,
		ExpiresAt:  time.Now().Add(s.cfg.Auth.MagicLinkTTL),
	}
	if err := s.repo.CreateVerification(ctx, verification); err != nil {
		return fmt.Errorf("store magic link: %w", err)
	}

	link := fmt.Sprintf("%s/api/v1/auth/magic-link/verify?token=%s&email=%s",
		s.cfg.App.APIBaseURL, url.QueryEscape(rawToken), url.QueryEscape(email))

	displayName := strings.Split(email, "@")[0]

	if err := s.email.SendMagicLink(ctx, email, displayName, link); err != nil {
		return fmt.Errorf("send magic link email: %w", err)
	}
	return nil
}

func (s *AuthService) ConsumeMagicLink(ctx context.Context, rawEmail, rawToken, ip, userAgent string, countryCode *string) (*IssuedSession, error) {
	email := normalizeEmail(rawEmail)
	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.MagicLinkSigningKey)

	if _, err := s.repo.ConsumeVerification(ctx, email, hash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidOrExpiredLink
		}
		return nil, fmt.Errorf("consume magic link: %w", err)
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("lookup user by email: %w", err)
		}
		user, err = s.createUser(ctx, email, strings.Split(email, "@")[0], "", true)
		if err != nil {
			return nil, err
		}
	}

	if !user.Active() {
		return nil, ErrAccountInactive
	}
	if !user.EmailVerified {
		if err := s.repo.MarkEmailVerified(ctx, user.ID); err != nil {
			return nil, err
		}
		user.EmailVerified = true
	}

	return s.issueSession(ctx, user, ip, userAgent, countryCode)
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.SessionHashKey)
	return s.repo.DeleteSessionByTokenHash(ctx, hash)
}

func (s *AuthService) LogoutAllSessions(ctx context.Context, userID string) error {
	return s.repo.DeleteAllSessionsForUser(ctx, userID)
}

func (s *AuthService) issueSession(ctx context.Context, user *models.User, ip, userAgent string, countryCode *string) (*IssuedSession, error) {
	id, err := utils.NewID()
	if err != nil {
		return nil, err
	}

	rawToken, expiresAt, err := utils.GenerateSessionToken(s.cfg.Auth.JWTSecret, user.ID, id, s.cfg.Auth.SessionTTL)
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}
	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.SessionHashKey)

	var ipPtr, uaPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	session := &models.Session{
		ID:        id,
		ExpiresAt: expiresAt,
		Token:     hash,
		IPAddress: ipPtr,
		UserAgent: uaPtr,
		UserID:    user.ID,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	now := time.Now()
	_ = s.repo.TouchLastLogin(ctx, user.ID, now)
	user.LastLoginAt = &now

	if cc := normalizeCountryCode(countryCode); cc != nil && (user.CountryCode == nil || *user.CountryCode != *cc) {
		if err := s.repo.UpdateUserCountryCode(ctx, user.ID, *cc); err != nil {
			log.Printf("auth: update user country code failed: %v", err)
		} else {
			user.CountryCode = cc
		}
	}

	return &IssuedSession{User: user, RawToken: rawToken, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, rawToken string) (*models.User, *models.Session, error) {
	if rawToken == "" {
		return nil, nil, ErrInvalidSession
	}

	claims, err := utils.ParseSessionToken(s.cfg.Auth.JWTSecret, rawToken)
	if err != nil {
		return nil, nil, ErrInvalidSession
	}

	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.SessionHashKey)

	session, err := s.repo.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrInvalidSession
		}
		return nil, nil, fmt.Errorf("lookup session: %w", err)
	}

	if session.ID != claims.SessionID || session.UserID != claims.Subject || session.Expired() {
		_ = s.repo.DeleteSessionByTokenHash(ctx, hash)
		return nil, nil, ErrInvalidSession
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrInvalidSession
		}
		return nil, nil, fmt.Errorf("lookup session user: %w", err)
	}
	if !user.Active() {
		return nil, nil, ErrAccountInactive
	}

	return user, session, nil
}

func (s *AuthService) createUser(ctx context.Context, email, name, avatarURL string, emailVerified bool) (*models.User, error) {
	id, err := utils.NewID()
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	var image *string
	if avatarURL != "" {
		image = &avatarURL
	}

	var emailVerifiedAt *time.Time
	if emailVerified {
		now := time.Now()
		emailVerifiedAt = &now
	}

	timezone := "UTC"

	user := &models.User{
		ID:                     id,
		Name:                   name,
		Email:                  email,
		EmailVerified:          emailVerified,
		EmailVerifiedAt:        emailVerifiedAt,
		Image:                  image,
		DefaultPrepTimeSeconds: 0,
		Timezone:               &timezone,
		Role:                   models.UserRoleUser,
		IsActive:               true,
		OnboardingCompleted:    false,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			existing, getErr := s.repo.GetUserByEmail(ctx, email)
			if getErr == nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logAudit(ctx, nil, "user.created", "user", user.ID, nil, nil)

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.email.SendWelcome(bgCtx, user.Email, user.Name)
	}()

	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username *string) *string {
	if username == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*username)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeCountryCode(code *string) *string {
	if code == nil {
		return nil
	}
	trimmed := strings.ToUpper(strings.TrimSpace(*code))
	if len(trimmed) != 2 {
		return nil
	}
	return &trimmed
}

type UpdateProfileInput struct {
	Name                   string
	Username               *string
	Image                  *string
	DefaultPrepTimeSeconds int
	Timezone               *string
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*models.User, error) {
	username := normalizeUsername(input.Username)

	if err := s.repo.UpdateUserProfile(
		ctx, userID, input.Name, username, input.Image, input.DefaultPrepTimeSeconds, input.Timezone,
	); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("service: update profile: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: reload user after profile update: %w", err)
	}

	return user, nil
}

func emailChangeIdentifier(userID, newEmail string) string {
	return fmt.Sprintf("email-change:%s:%s", userID, newEmail)
}

func (s *AuthService) RequestEmailChange(ctx context.Context, userID, rawNewEmail string) error {
	newEmail := normalizeEmail(rawNewEmail)

	if _, err := s.repo.GetUserByEmail(ctx, newEmail); err == nil {
		return ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("check new email availability: %w", err)
	}

	rawToken, err := utils.NewToken(32)
	if err != nil {
		return fmt.Errorf("generate email change token: %w", err)
	}
	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.MagicLinkSigningKey)

	id, err := utils.NewID()
	if err != nil {
		return err
	}

	identifier := emailChangeIdentifier(userID, newEmail)

	if err := s.repo.DeleteVerificationsByIdentifier(ctx, identifier); err != nil {
		return fmt.Errorf("clear previous email change requests: %w", err)
	}

	verification := &models.Verification{
		ID:         id,
		Identifier: identifier,
		ValueHash:  hash,
		ExpiresAt:  time.Now().Add(s.cfg.Auth.MagicLinkTTL),
	}
	if err := s.repo.CreateVerification(ctx, verification); err != nil {
		return fmt.Errorf("store email change verification: %w", err)
	}

	link := fmt.Sprintf("%s/api/v1/auth/email/verify?token=%s&email=%s",
		s.cfg.App.APIBaseURL, url.QueryEscape(rawToken), url.QueryEscape(newEmail))

	if err := s.email.SendMagicLink(ctx, newEmail, strings.Split(newEmail, "@")[0], link); err != nil {
		return fmt.Errorf("send email change verification: %w", err)
	}

	return nil
}

func (s *AuthService) UploadAvatar(ctx context.Context, userID string, data []byte, contentType string) (*models.User, error) {
	if len(data) > maxAvatarBytes {
		return nil, ErrAvatarTooLarge
	}

	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		return nil, ErrAvatarInvalidType
	}

	sniffLen := 512
	if len(data) < sniffLen {
		sniffLen = len(data)
	}
	detected := http.DetectContentType(data[:sniffLen])
	if _, ok := allowedAvatarTypes[detected]; !ok {
		return nil, ErrAvatarInvalidType
	}

	id, err := utils.NewID()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("avatars/%s/%s.%s", userID, id, ext)

	if err := utils.UploadPublicObject(ctx, s.s3Client, s.s3Bucket, key, data, contentType); err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	imageURL := utils.PublicS3URL(s.s3Region, s.s3Bucket, key)

	if err := s.repo.UpdateUserImage(ctx, userID, imageURL); err != nil {
		return nil, fmt.Errorf("update user image: %w", err)
	}

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) ConsumeEmailChange(ctx context.Context, userID, rawNewEmail, rawToken string) (*models.User, error) {
	newEmail := normalizeEmail(rawNewEmail)
	hash := utils.HMACSHA256Hex(rawToken, s.cfg.Auth.MagicLinkSigningKey)
	identifier := emailChangeIdentifier(userID, newEmail)

	oldUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user before email change: %w", err)
	}
	oldEmail := oldUser.Email

	if _, err := s.repo.ConsumeVerification(ctx, identifier, hash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidOrExpiredLink
		}
		return nil, fmt.Errorf("consume email change verification: %w", err)
	}

	if _, err := s.repo.GetUserByEmail(ctx, newEmail); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check new email availability: %w", err)
	}

	if err := s.repo.UpdateUserEmail(ctx, userID, newEmail); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("update user email: %w", err)
	}

	s.logAudit(ctx, &userID, "user.email_changed", "user", userID,
		toJSONB(map[string]any{"email": oldEmail}),
		toJSONB(map[string]any{"email": newEmail}),
	)

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) ListSessions(ctx context.Context, userID string) ([]models.Session, error) {
	sessions, err := s.repo.ListSessionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: list sessions: %w", err)
	}
	return sessions, nil
}

func (s *AuthService) RevokeSession(ctx context.Context, id, userID string) error {
	if err := s.repo.DeleteSessionByID(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFoundForUser
		}
		return fmt.Errorf("service: revoke session: %w", err)
	}
	s.logAudit(ctx, &userID, "session.revoked", "session", id, nil, nil)
	return nil
}

func (s *AuthService) DeleteAccount(ctx context.Context, userID string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("load user before delete: %w", err)
	}

	s.logAudit(ctx, &userID, "user.deleted", "user", userID, toJSONB(map[string]any{"email": user.Email}), nil)

	if err := s.repo.DeleteAllSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete sessions on account deletion: %w", err)
	}
	if err := s.repo.DeleteAccountsForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete linked accounts on account deletion: %w", err)
	}
	if err := s.repo.DeleteVerificationsByIdentifier(ctx, user.Email); err != nil {
		return fmt.Errorf("delete verifications on account deletion: %w", err)
	}
	if err := s.repo.DeleteVerificationsByIdentifierPrefix(ctx, "email-change:"+userID+":"); err != nil {
		return fmt.Errorf("delete email-change verifications on account deletion: %w", err)
	}
	if err := s.billingRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete billing data on account deletion: %w", err)
	}
	if err := s.transcriptRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete transcripts on account deletion: %w", err)
	}
	if err := s.analysisRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete analysis data on account deletion: %w", err)
	}
	if err := s.gamificationRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete gamification data on account deletion: %w", err)
	}
	if err := s.vocabularyRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete vocabulary data on account deletion: %w", err)
	}
	if err := s.notificationRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete notification data on account deletion: %w", err)
	}
	if err := s.sessionRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete speaking sessions on account deletion: %w", err)
	}
	if err := s.challengeRepo.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete challenge completions on account deletion: %w", err)
	}
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
