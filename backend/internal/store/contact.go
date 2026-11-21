package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type NewContactMessage struct {
	Name    string
	Email   string
	Message string
}

func (s *Store) CreateContactMessage(ctx context.Context, m NewContactMessage) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO contact_messages (name, email, message)
		VALUES ($1, $2, $3)
		RETURNING id`,
		m.Name, m.Email, m.Message).Scan(&id)
	return id, err
}

type ContactMessage struct {
	ID        int64      `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Message   string     `json:"message" db:"message"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	ReadAt    *time.Time `json:"readAt" db:"read_at"`
}

func (s *Store) ListContactMessages(ctx context.Context, unreadOnly bool) ([]ContactMessage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, email, message, created_at, read_at
		FROM contact_messages
		WHERE NOT $1 OR read_at IS NULL
		ORDER BY created_at DESC`,
		unreadOnly)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[ContactMessage])
}

func (s *Store) SetContactMessageRead(ctx context.Context, id int64, read bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE contact_messages
		SET read_at = CASE WHEN $2 THEN coalesce(read_at, now()) ELSE NULL END
		WHERE id = $1`,
		id, read)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteContactMessage(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM contact_messages WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
