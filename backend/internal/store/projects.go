package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrSlugTaken = errors.New("slug already taken")

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

const projectColumns = `id, title, slug, summary, description, tech_stack, repo_url, live_url,
	sort_order, published, created_at, updated_at`

type NewProject struct {
	Title       string
	Slug        string
	Summary     string
	Description string
	TechStack   []string
	RepoURL     *string
	LiveURL     *string
	SortOrder   int
	Published   bool
}
type ProjectPatch struct {
	Title       *string
	Slug        *string
	Summary     *string
	Description *string
	TechStack   *[]string
	RepoURL     *string
	LiveURL     *string
	SortOrder   *int
	Published   *bool
}

func (s *Store) ListPublishedProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+projectColumns+`
		FROM projects
		WHERE published
		ORDER BY sort_order, created_at DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Project])
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+projectColumns+`
		FROM projects
		ORDER BY sort_order, created_at DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Project])
}

func (s *Store) CreateProject(ctx context.Context, p NewProject) (Project, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO projects (title, slug, summary, description, tech_stack, repo_url, live_url, sort_order, published)
		VALUES ($1, $2, $3, $4, coalesce($5, '{}'::text[]), $6, $7, $8, $9)
		RETURNING `+projectColumns,
		p.Title, p.Slug, p.Summary, p.Description, p.TechStack, p.RepoURL, p.LiveURL, p.SortOrder, p.Published)
	if err != nil {
		return Project{}, err
	}
	project, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Project])
	return project, projectError(err)
}

func (s *Store) UpdateProject(ctx context.Context, id int64, p ProjectPatch) (Project, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE projects SET
			title       = coalesce($2, title),
			slug        = coalesce($3, slug),
			summary     = coalesce($4, summary),
			description = coalesce($5, description),
			tech_stack  = coalesce($6, tech_stack),
			repo_url    = nullif(coalesce($7, repo_url), ''),
			live_url    = nullif(coalesce($8, live_url), ''),
			sort_order  = coalesce($9, sort_order),
			published   = coalesce($10, published),
			updated_at  = now()
		WHERE id = $1
		RETURNING `+projectColumns,
		id, p.Title, p.Slug, p.Summary, p.Description, p.TechStack, p.RepoURL, p.LiveURL, p.SortOrder, p.Published)
	if err != nil {
		return Project{}, err
	}
	project, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Project])
	return project, projectError(err)
}

func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// projectError maps Postgres errors to the store's sentinel errors.
func projectError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "projects_slug_key" {
		return ErrSlugTaken
	}
	return err
}
