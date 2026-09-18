package models

import (
	"errors"
	"time"
)

type User struct {
	ID                     string     `db:"id" json:"id"`
	Name                   string     `db:"name" json:"name"`
	Email                  string     `db:"email" json:"email"`
	EmailVerified          bool       `db:"email_verified" json:"emailVerified"`
	EmailVerifiedAt        *time.Time `db:"email_verified_at" json:"emailVerifiedAt,omitempty"`
	Image                  *string    `db:"image" json:"image,omitempty"`
	Username               *string    `db:"username" json:"username,omitempty"`
	DefaultPrepTimeSeconds int        `db:"default_prep_time_seconds" json:"defaultPrepTimeSeconds"`
	Timezone               *string    `db:"timezone" json:"timezone,omitempty"`
	CountryCode            *string    `db:"country_code" json:"countryCode,omitempty"` // NEW: ISO 3166-1 alpha-2, drives global/country leaderboards
	Role                   UserRole   `db:"role" json:"role"`
	IsActive               bool       `db:"is_active" json:"isActive"`
	OnboardingCompleted    bool       `db:"onboarding_completed" json:"onboardingCompleted"`
	LastLoginAt            *time.Time `db:"last_login_at" json:"lastLoginAt,omitempty"`
	DeletedAt              *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updatedAt"`
}

func (User) TableName() string {
	return "user"
}

func (u User) Validate() error {
	var errs []error

	if u.ID == "" {
		errs = append(errs, errors.New("id is required"))
	}

	if u.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	if u.Email == "" {
		errs = append(errs, errors.New("email is required"))
	}

	if !u.Role.Valid() {
		errs = append(errs, errors.New("invalid role"))
	}

	if u.DefaultPrepTimeSeconds < 0 {
		errs = append(errs, errors.New("defaultPrepTimeSeconds must be >= 0"))
	}

	if u.EmailVerified && u.EmailVerifiedAt == nil {
		errs = append(errs, errors.New("emailVerifiedAt is required when emailVerified is true"))
	}

	// NEW: matches the db-level `country_code ~ '^[A-Z]{2}$'` check constraint.
	if u.CountryCode != nil && !isValidCountryCode(*u.CountryCode) {
		errs = append(errs, errors.New("countryCode must be a 2-letter uppercase ISO code"))
	}

	return errors.Join(errs...)
}

func isValidCountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func (u User) Active() bool {
	return u.IsActive && u.DeletedAt == nil
}

func (u User) Deleted() bool {
	return u.DeletedAt != nil
}

func (u User) Onboarded() bool {
	return u.OnboardingCompleted
}

// NEW
func (u User) HasCountry() bool {
	return u.CountryCode != nil && *u.CountryCode != ""
}

type Session struct {
	ID        string    `db:"id" json:"id"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	Token     string    `db:"token" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
	IPAddress *string   `db:"ip_address" json:"ipAddress,omitempty"`
	UserAgent *string   `db:"user_agent" json:"userAgent,omitempty"`
	UserID    string    `db:"user_id" json:"userId"`
}

func (Session) TableName() string {
	return "session"
}

func (s Session) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s Session) Valid() bool {
	return s.ID != "" &&
		s.UserID != "" &&
		s.Token != "" &&
		!s.Expired()
}

func (s Session) ExpiresSoon(within time.Duration) bool {
	return time.Now().Add(within).After(s.ExpiresAt)
}

func (s Session) Validate() error {
	var errs []error

	if s.ID == "" {
		errs = append(errs, errors.New("id is required"))
	}

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if s.Token == "" {
		errs = append(errs, errors.New("token is required"))
	}

	if s.ExpiresAt.IsZero() {
		errs = append(errs, errors.New("expiresAt is required"))
	}

	return errors.Join(errs...)
}

type Account struct {
	ID         string `db:"id" json:"id"`
	AccountID  string `db:"account_id" json:"accountId"`
	ProviderID string `db:"provider_id" json:"providerId"`
	UserID     string `db:"user_id" json:"userId"`

	AccessToken           *string    `db:"access_token" json:"-"`
	RefreshToken          *string    `db:"refresh_token" json:"-"`
	IDToken               *string    `db:"id_token" json:"-"`
	AccessTokenExpiresAt  *time.Time `db:"access_token_expires_at" json:"accessTokenExpiresAt,omitempty"`
	RefreshTokenExpiresAt *time.Time `db:"refresh_token_expires_at" json:"refreshTokenExpiresAt,omitempty"`
	Scope                 *string    `db:"scope" json:"scope,omitempty"`
	PasswordHash          *string    `db:"password_hash" json:"-"`
	FailedLoginAttempts   int        `db:"failed_login_attempts" json:"-"`
	LockedUntil           *time.Time `db:"locked_until" json:"-"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

func (Account) TableName() string {
	return "account"
}

func (a Account) IsCredentialAccount() bool {
	return a.ProviderID == "credential"
}

func (a Account) IsOAuthAccount() bool {
	return a.ProviderID != "credential"
}

func (a Account) AccessTokenExpired() bool {
	if a.AccessTokenExpiresAt == nil {
		return false
	}

	return time.Now().After(*a.AccessTokenExpiresAt)
}

func (a Account) RefreshTokenExpired() bool {
	if a.RefreshTokenExpiresAt == nil {
		return false
	}

	return time.Now().After(*a.RefreshTokenExpiresAt)
}

func (a Account) Locked() bool {
	return a.LockedUntil != nil && time.Now().Before(*a.LockedUntil)
}

func (a Account) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("id is required"))
	}

	if a.AccountID == "" {
		errs = append(errs, errors.New("accountId is required"))
	}

	if a.ProviderID == "" {
		errs = append(errs, errors.New("providerId is required"))
	}

	if a.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if a.IsCredentialAccount() && a.PasswordHash == nil {
		errs = append(errs, errors.New("passwordHash is required for credential accounts"))
	}

	if a.FailedLoginAttempts < 0 {
		errs = append(errs, errors.New("failedLoginAttempts must be >= 0"))
	}

	return errors.Join(errs...)
}

type Verification struct {
	ID         string     `db:"id" json:"id"`
	Identifier string     `db:"identifier" json:"identifier"`
	ValueHash  string     `db:"value_hash" json:"-"`
	UsedAt     *time.Time `db:"used_at" json:"-"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expiresAt"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
}

func (Verification) TableName() string {
	return "verification"
}

func (v Verification) Expired() bool {
	return time.Now().After(v.ExpiresAt)
}

func (v Verification) Used() bool {
	return v.UsedAt != nil
}

func (v Verification) Valid() bool {
	return v.ID != "" &&
		v.Identifier != "" &&
		v.ValueHash != "" &&
		!v.Expired() &&
		!v.Used()
}

func (v Verification) Validate() error {
	var errs []error

	if v.ID == "" {
		errs = append(errs, errors.New("id is required"))
	}

	if v.Identifier == "" {
		errs = append(errs, errors.New("identifier is required"))
	}

	if v.ValueHash == "" {
		errs = append(errs, errors.New("valueHash is required"))
	}

	if v.ExpiresAt.IsZero() {
		errs = append(errs, errors.New("expiresAt is required"))
	}

	return errors.Join(errs...)
}
