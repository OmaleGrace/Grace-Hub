package users

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"api/internal/wallet"
)

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrTaken          = errors.New("username or email already in use")
	ErrBadCredentials = errors.New("invalid username or password")
)

var usernameRe = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func Register(ctx context.Context, pool *pgxpool.Pool, username, email, password string) (*User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))
	if !usernameRe.MatchString(username) ||
		!strings.Contains(email, "@") || len(email) > 254 ||
		len(password) < 8 || len(password) > 72 {
		return nil, ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var u User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, created_at`,
		username, email, string(hash),
	).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrTaken
	}
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	if err := wallet.GrantSignupBonus(ctx, tx, u.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &u, nil
}

func Authenticate(ctx context.Context, pool *pgxpool.Pool, login, password string) (*User, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	var u User
	var hash string
	err := pool.QueryRow(ctx, `
		SELECT id, username, email, created_at, password_hash
		FROM users WHERE username = $1 OR email = $1`, login,
	).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &hash)

	if errors.Is(err, pgx.ErrNoRows) {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, ErrBadCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, ErrBadCredentials
	}
	return &u, nil
}