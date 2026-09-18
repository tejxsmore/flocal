package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"flocal/internal/config"
)

const oauthProviderRequestTimeout = 10 * time.Second

type OAuthProfile struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Name           string
	AvatarURL      string
}

type OAuthService struct {
	google *oauth2.Config
	github *oauth2.Config
}

func NewOAuthService(cfg *config.Config) *OAuthService {
	svc := &OAuthService{}

	if cfg.Google.Enabled {
		svc.google = &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes: []string{
				"openid",
				"email",
				"profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	if cfg.GitHub.Enabled {
		svc.github = &oauth2.Config{
			ClientID:     cfg.GitHub.ClientID,
			ClientSecret: cfg.GitHub.ClientSecret,
			RedirectURL:  cfg.GitHub.RedirectURL,
			Scopes: []string{
				"read:user",
				"user:email",
			},
			Endpoint: github.Endpoint,
		}
	}

	return svc
}

func (s *OAuthService) GoogleEnabled() bool {
	return s.google != nil
}

func (s *OAuthService) GitHubEnabled() bool {
	return s.github != nil
}

func (s *OAuthService) AuthCodeURL(provider, state string) (string, error) {
	cfg, err := s.configFor(provider)
	if err != nil {
		return "", err
	}

	switch provider {
	case "google":
		return cfg.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("prompt", "select_account"),
		), nil

	case "github":
		return cfg.AuthCodeURL(state), nil

	default:
		return "", fmt.Errorf("service: unknown oauth provider %q", provider)
	}
}

func (s *OAuthService) Exchange(
	ctx context.Context,
	provider string,
	code string,
) (*oauth2.Token, error) {
	cfg, err := s.configFor(provider)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, oauthProviderRequestTimeout)
	defer cancel()

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf(
			"service: exchange %s code: %w",
			provider,
			err,
		)
	}

	return token, nil
}

func (s *OAuthService) FetchProfile(
	ctx context.Context,
	provider string,
	token *oauth2.Token,
) (*OAuthProfile, error) {
	switch provider {
	case "google":
		return s.fetchGoogleProfile(ctx, token)

	case "github":
		return s.fetchGitHubProfile(ctx, token)

	default:
		return nil, fmt.Errorf(
			"service: unknown oauth provider %q",
			provider,
		)
	}
}

func (s *OAuthService) configFor(provider string) (*oauth2.Config, error) {
	switch provider {
	case "google":
		if s.google == nil {
			return nil, fmt.Errorf(
				"service: google oauth is not configured",
			)
		}

		return s.google, nil

	case "github":
		if s.github == nil {
			return nil, fmt.Errorf(
				"service: github oauth is not configured",
			)
		}

		return s.github, nil

	default:
		return nil, fmt.Errorf(
			"service: unknown oauth provider %q",
			provider,
		)
	}
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *OAuthService) fetchGoogleProfile(
	ctx context.Context,
	token *oauth2.Token,
) (*OAuthProfile, error) {
	client := s.google.Client(ctx, token)

	body, err := getJSON(
		ctx,
		client,
		"https://www.googleapis.com/oauth2/v3/userinfo",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"service: fetch google profile: %w",
			err,
		)
	}

	var info googleUserInfo

	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf(
			"service: decode google profile: %w",
			err,
		)
	}

	email := strings.ToLower(strings.TrimSpace(info.Email))

	if email == "" {
		return nil, fmt.Errorf(
			"service: google profile has no email",
		)
	}

	if info.Sub == "" {
		return nil, fmt.Errorf(
			"service: google profile has no user id",
		)
	}

	return &OAuthProfile{
		ProviderUserID: info.Sub,
		Email:          email,
		EmailVerified:  info.EmailVerified,
		Name:           strings.TrimSpace(info.Name),
		AvatarURL:      strings.TrimSpace(info.Picture),
	}, nil
}

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *OAuthService) fetchGitHubProfile(
	ctx context.Context,
	token *oauth2.Token,
) (*OAuthProfile, error) {
	client := s.github.Client(ctx, token)

	userBody, err := getJSON(
		ctx,
		client,
		"https://api.github.com/user",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"service: fetch github profile: %w",
			err,
		)
	}

	var user githubUser

	if err := json.Unmarshal(userBody, &user); err != nil {
		return nil, fmt.Errorf(
			"service: decode github profile: %w",
			err,
		)
	}

	if user.ID == 0 {
		return nil, fmt.Errorf(
			"service: github profile has no user id",
		)
	}

	name := strings.TrimSpace(user.Name)
	if name == "" {
		name = strings.TrimSpace(user.Login)
	}

	profile := &OAuthProfile{
		ProviderUserID: fmt.Sprintf("%d", user.ID),
		Name:           name,
		AvatarURL:      strings.TrimSpace(user.AvatarURL),
	}

	emailsBody, err := getJSON(
		ctx,
		client,
		"https://api.github.com/user/emails",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"service: fetch github emails: %w",
			err,
		)
	}

	var emails []githubEmail

	if err := json.Unmarshal(emailsBody, &emails); err != nil {
		return nil, fmt.Errorf(
			"service: decode github emails: %w",
			err,
		)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			profile.Email = strings.ToLower(
				strings.TrimSpace(email.Email),
			)
			profile.EmailVerified = true

			return profile, nil
		}
	}

	for _, email := range emails {
		if email.Verified {
			profile.Email = strings.ToLower(
				strings.TrimSpace(email.Email),
			)
			profile.EmailVerified = true

			return profile, nil
		}
	}

	return nil, fmt.Errorf(
		"service: no verified email found on github account",
	)
}

func getJSON(
	ctx context.Context,
	client *http.Client,
	url string,
) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, oauthProviderRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build request: %w",
			err,
		)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"send request: %w",
			err,
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"read response: %w",
			err,
		)
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"unexpected status %d from %s: %s",
			resp.StatusCode,
			url,
			string(body),
		)
	}

	return body, nil
}
