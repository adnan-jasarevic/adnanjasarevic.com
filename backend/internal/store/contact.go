package store

import "context"

type NewContactMessage struct {
	Name    string
	Email   string
	Message string
}

func (s *Store) CreateContactMessage(ctx context.Context, m NewContactMessage) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO contact_messages (name, email, message)
		VALUES ($1, $2, $3)`,
		m.Name, m.Email, m.Message)
	return err
}
