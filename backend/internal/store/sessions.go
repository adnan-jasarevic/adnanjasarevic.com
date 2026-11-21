package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("not found")

type AdminSession struct {
	GithubID    int64     `db:"github_id"`
	GithubLogin string    `db:"github_login"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (s *Store) CreateAdminSession(ctx context.Context, tokenHash []byte, githubID int64, githubLogin string, expiresAt time.Time) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM admin_sessions WHERE expires_at <= now()`); err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO admin_sessions (token_hash, github_id, github_login, expires_at)
		VALUES ($1, $2, $3, $4)`,
		tokenHash, githubID, githubLogin, expiresAt)
	return err
}

func (s *Store) GetAdminSession(ctx context.Context, tokenHash []byte) (AdminSession, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT github_id, github_login, expires_at
		FROM admin_sessions
		WHERE token_hash = $1 AND expires_at > now()`,
		tokenHash)
	if err != nil {
		return AdminSession{}, err
	}
	session, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[AdminSession])
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminSession{}, ErrNotFound
	}
	return session, err
}

func (s *Store) DeleteAdminSession(ctx context.Context, tokenHash []byte) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
	return err
}
