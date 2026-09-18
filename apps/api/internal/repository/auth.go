package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

var (
	ErrNotFound = errors.New("repository: not found")
	ErrConflict = errors.New("repository: conflict")
)

type AuthRepository interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, u *models.User) error
	MarkEmailVerified(ctx context.Context, userID string) error
	TouchLastLogin(ctx context.Context, userID string, at time.Time) error

	GetAccountByProvider(ctx context.Context, providerID, accountID string) (*models.Account, error)
	CreateAccount(ctx context.Context, a *models.Account) error
	UpdateAccountTokens(ctx context.Context, a *models.Account) error
	ListAccountsForUser(ctx context.Context, userID string) ([]models.Account, error)

	CreateSession(ctx context.Context, s *models.Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
	DeleteAllSessionsForUser(ctx context.Context, userID string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)

	CreateVerification(ctx context.Context, v *models.Verification) error
	ConsumeVerification(ctx context.Context, identifier, valueHash string) (*models.Verification, error)
	DeleteVerification(ctx context.Context, id string) error
	DeleteVerificationsByIdentifier(ctx context.Context, identifier string) error
	DeleteVerificationsByIdentifierPrefix(ctx context.Context, prefix string) error

	UpdateUserProfile(ctx context.Context, userID, name string, username, image *string, defaultPrepTimeSeconds int, timezone *string) error
	UpdateUserEmail(ctx context.Context, userID, email string) error
	UpdateUserCountryCode(ctx context.Context, userID, countryCode string) error
	DeleteAccountsForUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	UpdateUserImage(ctx context.Context, userID, imageURL string) error
	ListSessionsForUser(ctx context.Context, userID string) ([]models.Session, error)
	DeleteSessionByID(ctx context.Context, id, userID string) error
}

type pgAuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) AuthRepository {
	return &pgAuthRepository{pool: pool}
}

func (r *pgAuthRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	const q = `
		select id, name, email, email_verified, email_verified_at, image, username,
		       default_prep_time_seconds, timezone, country_code, role, is_active,
		       onboarding_completed, last_login_at, deleted_at, created_at, updated_at
		from "user"
		where id = $1`

	var u models.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.EmailVerifiedAt, &u.Image, &u.Username,
		&u.DefaultPrepTimeSeconds, &u.Timezone, &u.CountryCode, &u.Role, &u.IsActive,
		&u.OnboardingCompleted, &u.LastLoginAt, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user by id: %w", err)
	}
	return &u, nil
}

func (r *pgAuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `
		select id, name, email, email_verified, email_verified_at, image, username,
		       default_prep_time_seconds, timezone, country_code, role, is_active,
		       onboarding_completed, last_login_at, deleted_at, created_at, updated_at
		from "user"
		where email = $1`

	var u models.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.EmailVerified, &u.EmailVerifiedAt, &u.Image, &u.Username,
		&u.DefaultPrepTimeSeconds, &u.Timezone, &u.CountryCode, &u.Role, &u.IsActive,
		&u.OnboardingCompleted, &u.LastLoginAt, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get user by email: %w", err)
	}
	return &u, nil
}

func (r *pgAuthRepository) CreateUser(ctx context.Context, u *models.User) error {
	const q = `
		insert into "user" (id, name, email, email_verified, email_verified_at, image, username,
		                     default_prep_time_seconds, timezone, is_active,
		                     onboarding_completed, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
		returning created_at, updated_at`

	err := r.pool.QueryRow(ctx, q,
		u.ID, u.Name, u.Email, u.EmailVerified, u.EmailVerifiedAt, u.Image, u.Username,
		u.DefaultPrepTimeSeconds, u.Timezone, u.IsActive, u.OnboardingCompleted,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("repository: create user: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	const q = `
		update "user"
		set email_verified = true,
		    email_verified_at = coalesce(email_verified_at, now()),
		    updated_at = now()
		where id = $1`

	ct, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("repository: mark email verified: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgAuthRepository) TouchLastLogin(ctx context.Context, userID string, at time.Time) error {
	const q = `
		update "user"
		set last_login_at = $2, updated_at = now()
		where id = $1`

	ct, err := r.pool.Exec(ctx, q, userID, at)
	if err != nil {
		return fmt.Errorf("repository: touch last login: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgAuthRepository) GetAccountByProvider(ctx context.Context, providerID, accountID string) (*models.Account, error) {
	const q = `
		select id, account_id, provider_id, user_id, access_token, refresh_token,
		       id_token, access_token_expires_at, refresh_token_expires_at,
		       scope, password_hash, failed_login_attempts, locked_until, created_at, updated_at
		from account
		where provider_id = $1 and account_id = $2`

	var a models.Account
	err := r.pool.QueryRow(ctx, q, providerID, accountID).Scan(
		&a.ID, &a.AccountID, &a.ProviderID, &a.UserID, &a.AccessToken, &a.RefreshToken,
		&a.IDToken, &a.AccessTokenExpiresAt, &a.RefreshTokenExpiresAt,
		&a.Scope, &a.PasswordHash, &a.FailedLoginAttempts, &a.LockedUntil, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get account by provider: %w", err)
	}
	return &a, nil
}

func (r *pgAuthRepository) CreateAccount(ctx context.Context, a *models.Account) error {
	const q = `
		insert into account (id, account_id, provider_id, user_id, access_token,
		                      refresh_token, id_token, access_token_expires_at,
		                      refresh_token_expires_at, scope, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		returning created_at, updated_at`

	err := r.pool.QueryRow(ctx, q,
		a.ID, a.AccountID, a.ProviderID, a.UserID, a.AccessToken,
		a.RefreshToken, a.IDToken, a.AccessTokenExpiresAt,
		a.RefreshTokenExpiresAt, a.Scope,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("repository: create account: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) UpdateAccountTokens(ctx context.Context, a *models.Account) error {
	const q = `
		update account
		set access_token = $2,
		    refresh_token = coalesce($3, refresh_token),
		    id_token = $4,
		    access_token_expires_at = $5,
		    refresh_token_expires_at = coalesce($6, refresh_token_expires_at),
		    scope = $7,
		    updated_at = now()
		where id = $1`

	ct, err := r.pool.Exec(
		ctx,
		q,
		a.ID,
		a.AccessToken,
		a.RefreshToken,
		a.IDToken,
		a.AccessTokenExpiresAt,
		a.RefreshTokenExpiresAt,
		a.Scope,
	)
	if err != nil {
		return fmt.Errorf("repository: update account tokens: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgAuthRepository) ListAccountsForUser(ctx context.Context, userID string) ([]models.Account, error) {
	const q = `
		select id, account_id, provider_id, user_id, access_token, refresh_token,
		       id_token, access_token_expires_at, refresh_token_expires_at,
		       scope, password_hash, failed_login_attempts, locked_until, created_at, updated_at
		from account
		where user_id = $1`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list accounts for user: %w", err)
	}
	defer rows.Close()

	accounts := make([]models.Account, 0)
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(
			&a.ID, &a.AccountID, &a.ProviderID, &a.UserID, &a.AccessToken, &a.RefreshToken,
			&a.IDToken, &a.AccessTokenExpiresAt, &a.RefreshTokenExpiresAt,
			&a.Scope, &a.PasswordHash, &a.FailedLoginAttempts, &a.LockedUntil, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list accounts for user: %w", err)
	}
	return accounts, nil
}

func (r *pgAuthRepository) CreateSession(ctx context.Context, s *models.Session) error {
	const q = `
		insert into "session" (id, expires_at, token, ip_address, user_agent, user_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, now(), now())
		returning created_at, updated_at`

	err := r.pool.QueryRow(ctx, q,
		s.ID, s.ExpiresAt, s.Token, s.IPAddress, s.UserAgent, s.UserID,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository: create session: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	const q = `
		select id, expires_at, token, created_at, updated_at, ip_address, user_agent, user_id
		from "session"
		where token = $1`

	var s models.Session
	err := r.pool.QueryRow(ctx, q, tokenHash).Scan(
		&s.ID, &s.ExpiresAt, &s.Token, &s.CreatedAt, &s.UpdatedAt,
		&s.IPAddress, &s.UserAgent, &s.UserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get session by token: %w", err)
	}
	return &s, nil
}

func (r *pgAuthRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	const q = `delete from "session" where token = $1`
	_, err := r.pool.Exec(ctx, q, tokenHash)
	if err != nil {
		return fmt.Errorf("repository: delete session: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) DeleteAllSessionsForUser(ctx context.Context, userID string) error {
	const q = `delete from "session" where user_id = $1`
	_, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("repository: delete all sessions for user: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	const q = `delete from "session" where expires_at < now()`
	ct, err := r.pool.Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("repository: delete expired sessions: %w", err)
	}
	return ct.RowsAffected(), nil
}

func (r *pgAuthRepository) CreateVerification(ctx context.Context, v *models.Verification) error {
	const q = `
		insert into verification (id, identifier, value_hash, expires_at, created_at, updated_at)
		values ($1, $2, $3, $4, now(), now())
		returning created_at, updated_at`

	err := r.pool.QueryRow(
		ctx,
		q,
		v.ID,
		v.Identifier,
		v.ValueHash,
		v.ExpiresAt,
	).Scan(
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create verification: %w", err)
	}

	return nil
}

func (r *pgAuthRepository) ConsumeVerification(ctx context.Context, identifier, valueHash string) (*models.Verification, error) {
	const q = `
		update verification
		set used_at = now(), updated_at = now()
		where identifier = $1 and value_hash = $2 and used_at is null and expires_at > now()
		returning id, identifier, value_hash, used_at, expires_at, created_at, updated_at`

	var v models.Verification
	err := r.pool.QueryRow(ctx, q, identifier, valueHash).Scan(
		&v.ID, &v.Identifier, &v.ValueHash, &v.UsedAt, &v.ExpiresAt, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: consume verification: %w", err)
	}
	return &v, nil
}

func (r *pgAuthRepository) DeleteVerification(ctx context.Context, id string) error {
	const q = `delete from verification where id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repository: delete verification: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) DeleteVerificationsByIdentifier(ctx context.Context, identifier string) error {
	const q = `delete from verification where identifier = $1`
	_, err := r.pool.Exec(ctx, q, identifier)
	if err != nil {
		return fmt.Errorf("repository: delete verifications by identifier: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) DeleteVerificationsByIdentifierPrefix(ctx context.Context, prefix string) error {
	const q = `delete from verification where identifier like $1`
	_, err := r.pool.Exec(ctx, q, prefix+"%")
	if err != nil {
		return fmt.Errorf("repository: delete verifications by identifier prefix: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) UpdateUserProfile(
	ctx context.Context,
	userID, name string,
	username, image *string,
	defaultPrepTimeSeconds int,
	timezone *string,
) error {
	const q = `
		update "user"
		set name = $2,
		    username = $3,
		    image = $4,
		    default_prep_time_seconds = $5,
		    timezone = $6,
		    updated_at = now()
		where id = $1`

	ct, err := r.pool.Exec(ctx, q, userID, name, username, image, defaultPrepTimeSeconds, timezone)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("repository: update user profile: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgAuthRepository) UpdateUserEmail(ctx context.Context, userID, email string) error {
	const q = `
		update "user"
		set email = $2, email_verified = true, email_verified_at = now(), updated_at = now()
		where id = $1`

	ct, err := r.pool.Exec(ctx, q, userID, email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("repository: update user email: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgAuthRepository) UpdateUserCountryCode(ctx context.Context, userID, countryCode string) error {
	const q = `update "user" set country_code = $2, updated_at = now() where id = $1`
	ct, err := r.pool.Exec(ctx, q, userID, countryCode)
	if err != nil {
		return fmt.Errorf("repository: update user country code: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgAuthRepository) DeleteAccountsForUser(ctx context.Context, userID string) error {
	const q = `delete from account where user_id = $1`
	_, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("repository: delete accounts for user: %w", err)
	}
	return nil
}

func (r *pgAuthRepository) DeleteUser(ctx context.Context, userID string) error {
	const q = `delete from "user" where id = $1`
	ct, err := r.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("repository: delete user: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgAuthRepository) UpdateUserImage(ctx context.Context, userID, imageURL string) error {
	const q = `update "user" set image = $2, updated_at = now() where id = $1`
	ct, err := r.pool.Exec(ctx, q, userID, imageURL)
	if err != nil {
		return fmt.Errorf("repository: update user image: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgAuthRepository) ListSessionsForUser(ctx context.Context, userID string) ([]models.Session, error) {
	const q = `
		select id, expires_at, token, created_at, updated_at, ip_address, user_agent, user_id
		from "session"
		where user_id = $1
		order by created_at desc`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list sessions for user: %w", err)
	}
	defer rows.Close()

	sessions := make([]models.Session, 0)
	for rows.Next() {
		var s models.Session
		if err := rows.Scan(
			&s.ID, &s.ExpiresAt, &s.Token, &s.CreatedAt, &s.UpdatedAt,
			&s.IPAddress, &s.UserAgent, &s.UserID,
		); err != nil {
			return nil, fmt.Errorf("repository: scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list sessions for user: %w", err)
	}
	return sessions, nil
}

func (r *pgAuthRepository) DeleteSessionByID(ctx context.Context, id, userID string) error {
	const q = `delete from "session" where id = $1 and user_id = $2`
	ct, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("repository: delete session by id: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
