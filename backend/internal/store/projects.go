package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Project struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	Summary     string    `json:"summary" db:"summary"`
	Description string    `json:"description" db:"description"`
	TechStack   []string  `json:"techStack" db:"tech_stack"`
	RepoURL     *string   `json:"repoUrl" db:"repo_url"`
	LiveURL     *string   `json:"liveUrl" db:"live_url"`
	SortOrder   int       `json:"sortOrder" db:"sort_order"`
	Published   bool      `json:"published" db:"published"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

func (s *Store) ListPublishedProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, slug, summary, description, tech_stack, repo_url, live_url,
		       sort_order, published, created_at, updated_at
		FROM projects
		WHERE published
		ORDER BY sort_order, created_at DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Project])
}
