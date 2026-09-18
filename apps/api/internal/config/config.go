package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	App      AppConfig
	DB       DatabaseConfig
	AWS      AWSConfig
	Deepgram DeepgramConfig
	OpenAI   OpenAIConfig
	Resend   ResendConfig
	Google   OAuthProviderConfig
	GitHub   OAuthProviderConfig
	Auth     AuthConfig
	Redis    RedisConfig
	Dodo     DodoConfig
	GeoIP    GeoIPConfig
}

type AppConfig struct {
	Name            string
	Port            string
	APIBaseURL      string
	FrontendURL     string
	AllowedOrigins  []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
}

type DatabaseConfig struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
	S3UploadPrefix  string
	PresignExpiry   time.Duration
}

type DeepgramConfig struct {
	APIKey   string
	Model    string
	Language string
}

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type ResendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

type AuthConfig struct {
	JWTSecret           string
	SessionHashKey      string
	SessionTTL          time.Duration
	MagicLinkTTL        time.Duration
	MagicLinkSigningKey string
	CookieDomain        string
	CookieSecure        bool
}

type RedisConfig struct {
	URL string
}

type DodoConfig struct {
	APIKey        string
	WebhookSecret string
	TestMode      bool
	ReturnURL     string
}

type GeoIPConfig struct {
	DBPath  string
	Enabled bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	env := getEnvDefault("APP_ENV", "development")

	cfg := &Config{
		Env: env,
		App: AppConfig{
			Name:            getEnvDefault("APP_NAME", "flocal-api"),
			Port:            getEnvDefault("PORT", "8080"),
			APIBaseURL:      strings.TrimSuffix(getEnvDefault("API_BASE_URL", "http://localhost:8080"), "/"),
			FrontendURL:     strings.TrimSuffix(getEnvDefault("FRONTEND_URL", "http://localhost:3000"), "/"),
			AllowedOrigins:  splitAndTrim(getEnvDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
			ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			TrustedProxies:  splitAndTrim(getEnvDefault("TRUSTED_PROXIES", "")),
		},
		DB: DatabaseConfig{
			URL:               os.Getenv("DATABASE_URL"),
			MaxConns:          int32(getEnvInt("DB_MAX_CONNS", 10)),
			MinConns:          int32(getEnvInt("DB_MIN_CONNS", 2)),
			MaxConnLifetime:   getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:   getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvDuration("DB_HEALTH_CHECK_PERIOD", time.Minute),
			ConnectTimeout:    getEnvDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
		},
		AWS: AWSConfig{
			Region:          os.Getenv("AWS_REGION"),
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			S3Bucket:        os.Getenv("AWS_S3_BUCKET"),
			S3UploadPrefix:  getEnvDefault("AWS_S3_UPLOAD_PREFIX", "sessions/"),
			PresignExpiry:   getEnvDuration("AWS_S3_PRESIGN_EXPIRY", 15*time.Minute),
		},
		Deepgram: DeepgramConfig{
			APIKey:   os.Getenv("DEEPGRAM_API_KEY"),
			Model:    getEnvDefault("DEEPGRAM_MODEL", "nova-3"),
			Language: getEnvDefault("DEEPGRAM_LANGUAGE", "en"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Model:   getEnvDefault("OPENAI_MODEL", "gpt-5-mini"),
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
		},
		Resend: ResendConfig{
			APIKey:    strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
			FromEmail: strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")),
			FromName:  getEnvDefault("RESEND_FROM_NAME", "Flocal"),
		},
		Google: OAuthProviderConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
		GitHub: OAuthProviderConfig{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		},
		Auth: AuthConfig{
			JWTSecret:           os.Getenv("JWT_SECRET"),
			SessionHashKey:      os.Getenv("SESSION_HASH_KEY"),
			SessionTTL:          getEnvDuration("SESSION_TTL", 30*24*time.Hour),
			MagicLinkTTL:        getEnvDuration("MAGIC_LINK_TTL", 15*time.Minute),
			MagicLinkSigningKey: os.Getenv("MAGIC_LINK_SIGNING_KEY"),
			CookieDomain:        os.Getenv("COOKIE_DOMAIN"),
			CookieSecure:        getEnvBool("COOKIE_SECURE", env == "production"),
		},
		Redis: RedisConfig{
			URL: getEnvDefault("REDIS_URL", "redis://localhost:6379/0"),
		},
		Dodo: DodoConfig{
			APIKey:        os.Getenv("DODO_API_KEY"),
			WebhookSecret: os.Getenv("DODO_WEBHOOK_SECRET"),
			TestMode:      getEnvBool("DODO_TEST_MODE", env != "production"),
			ReturnURL:     getEnvDefault("DODO_RETURN_URL", strings.TrimSuffix(getEnvDefault("FRONTEND_URL", "http://localhost:3000"), "/")+"/billing/return"),
		},
		GeoIP: GeoIPConfig{
			DBPath: getEnvDefault("GEOIP_DB_PATH", "internal/utils/GeoLite2-Country.mmdb"),
		},
	}

	cfg.Google.Enabled = cfg.Google.ClientID != "" && cfg.Google.ClientSecret != ""
	cfg.GitHub.Enabled = cfg.GitHub.ClientID != "" && cfg.GitHub.ClientSecret != ""
	cfg.GeoIP.Enabled = strings.TrimSpace(cfg.GeoIP.DBPath) != ""

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func (c *Config) validate() error {
	var missing []string

	require := func(name, value string) {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}

	require("DATABASE_URL", c.DB.URL)
	require("JWT_SECRET", c.Auth.JWTSecret)
	require("SESSION_HASH_KEY", c.Auth.SessionHashKey)
	require("MAGIC_LINK_SIGNING_KEY", c.Auth.MagicLinkSigningKey)
	require("RESEND_API_KEY", c.Resend.APIKey)
	require("RESEND_FROM_EMAIL", c.Resend.FromEmail)
	require("AWS_REGION", c.AWS.Region)
	require("AWS_ACCESS_KEY_ID", c.AWS.AccessKeyID)
	require("AWS_SECRET_ACCESS_KEY", c.AWS.SecretAccessKey)
	require("AWS_S3_BUCKET", c.AWS.S3Bucket)
	require("DEEPGRAM_API_KEY", c.Deepgram.APIKey)
	require("OPENAI_API_KEY", c.OpenAI.APIKey)
	require("DODO_API_KEY", c.Dodo.APIKey)
	require("DODO_WEBHOOK_SECRET", c.Dodo.WebhookSecret)

	if len(c.Auth.JWTSecret) > 0 && len(c.Auth.JWTSecret) < 32 {
		missing = append(missing, "JWT_SECRET (must be at least 32 characters)")
	}

	if len(c.Auth.SessionHashKey) > 0 && len(c.Auth.SessionHashKey) < 32 {
		missing = append(missing, "SESSION_HASH_KEY (must be at least 32 characters)")
	}

	if len(c.Auth.MagicLinkSigningKey) > 0 && len(c.Auth.MagicLinkSigningKey) < 32 {
		missing = append(missing, "MAGIC_LINK_SIGNING_KEY (must be at least 32 characters)")
	}

	if c.IsProduction() {
		if strings.TrimSpace(c.Auth.CookieDomain) == "" {
			missing = append(missing, "COOKIE_DOMAIN")
		}
		if len(c.App.AllowedOrigins) == 1 && c.App.AllowedOrigins[0] == "http://localhost:3000" {
			missing = append(missing, "CORS_ALLOWED_ORIGINS (still default localhost value)")
		}
		if c.Redis.URL == "redis://localhost:6379/0" {
			missing = append(missing, "REDIS_URL (still default localhost value)")
		}
		if c.Dodo.TestMode {
			missing = append(missing, "DODO_TEST_MODE (must be false in production)")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("config: missing/invalid required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func splitAndTrim(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
