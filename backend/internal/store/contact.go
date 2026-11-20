package store

import "context"

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
