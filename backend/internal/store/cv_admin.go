package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrSkillTaken = errors.New("skill name already taken")
	ErrDateOrder  = errors.New("end date before start date")
)

type ProfileInput struct {
	FullName    string
	Headline    string
	Summary     string
	Location    string
	Email       *string
	GithubURL   *string
	LinkedinURL *string
}

func (s *Store) PutProfile(ctx context.Context, p ProfileInput) (Profile, error) {
	return queryOne[Profile](ctx, s, `
		INSERT INTO profile (full_name, headline, summary, location, email, github_url, linkedin_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			full_name    = excluded.full_name,
			headline     = excluded.headline,
			summary      = excluded.summary,
			location     = excluded.location,
			email        = excluded.email,
			github_url   = excluded.github_url,
			linkedin_url = excluded.linkedin_url,
			updated_at   = now()
		RETURNING `+profileColumns,
		p.FullName, p.Headline, p.Summary, p.Location, p.Email, p.GithubURL, p.LinkedinURL)
}

type NewExperience struct {
	Company     string
	Role        string
	Location    string
	StartDate   time.Time
	EndDate     *time.Time
	Description string
	Highlights  []string
}

// ExperiencePatch holds the fields to change. Nil fields are left unchanged.
// ClearEndDate sets end_date to NULL and wins over EndDate.
type ExperiencePatch struct {
	Company      *string
	Role         *string
	Location     *string
	StartDate    *time.Time
	EndDate      *time.Time
	ClearEndDate bool
	Description  *string
	Highlights   *[]string
}

func (s *Store) CreateExperience(ctx context.Context, e NewExperience) (Experience, error) {
	return queryOne[Experience](ctx, s, `
		INSERT INTO experience (company, role, location, start_date, end_date, description, highlights)
		VALUES ($1, $2, $3, $4, $5, $6, coalesce($7, '{}'::text[]))
		RETURNING `+experienceColumns,
		e.Company, e.Role, e.Location, e.StartDate, e.EndDate, e.Description, e.Highlights)
}

func (s *Store) UpdateExperience(ctx context.Context, id int64, e ExperiencePatch) (Experience, error) {
	return queryOne[Experience](ctx, s, `
		UPDATE experience SET
			company     = coalesce($2, company),
			role        = coalesce($3, role),
			location    = coalesce($4, location),
			start_date  = coalesce($5, start_date),
			end_date    = CASE WHEN $7 THEN NULL ELSE coalesce($6, end_date) END,
			description = coalesce($8, description),
			highlights  = coalesce($9, highlights),
			updated_at  = now()
		WHERE id = $1
		RETURNING `+experienceColumns,
		id, e.Company, e.Role, e.Location, e.StartDate, e.EndDate, e.ClearEndDate, e.Description, e.Highlights)
}

func (s *Store) DeleteExperience(ctx context.Context, id int64) error {
	return s.deleteByID(ctx, "experience", id)
}

type NewEducation struct {
	Institution string
	Degree      string
	Field       string
	StartDate   time.Time
	EndDate     *time.Time
	Description string
}

// EducationPatch works like ExperiencePatch.
type EducationPatch struct {
	Institution  *string
	Degree       *string
	Field        *string
	StartDate    *time.Time
	EndDate      *time.Time
	ClearEndDate bool
	Description  *string
}

func (s *Store) CreateEducation(ctx context.Context, e NewEducation) (Education, error) {
	return queryOne[Education](ctx, s, `
		INSERT INTO education (institution, degree, field, start_date, end_date, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+educationColumns,
		e.Institution, e.Degree, e.Field, e.StartDate, e.EndDate, e.Description)
}

func (s *Store) UpdateEducation(ctx context.Context, id int64, e EducationPatch) (Education, error) {
	return queryOne[Education](ctx, s, `
		UPDATE education SET
			institution = coalesce($2, institution),
			degree      = coalesce($3, degree),
			field       = coalesce($4, field),
			start_date  = coalesce($5, start_date),
			end_date    = CASE WHEN $7 THEN NULL ELSE coalesce($6, end_date) END,
			description = coalesce($8, description),
			updated_at  = now()
		WHERE id = $1
		RETURNING `+educationColumns,
		id, e.Institution, e.Degree, e.Field, e.StartDate, e.EndDate, e.ClearEndDate, e.Description)
}

func (s *Store) DeleteEducation(ctx context.Context, id int64) error {
	return s.deleteByID(ctx, "education", id)
}

type NewSkill struct {
	Name      string
	Category  string
	SortOrder int
}

// SkillPatch holds the fields to change. Nil fields are left unchanged.
type SkillPatch struct {
	Name      *string
	Category  *string
	SortOrder *int
}

func (s *Store) CreateSkill(ctx context.Context, sk NewSkill) (Skill, error) {
	return queryOne[Skill](ctx, s, `
		INSERT INTO skills (name, category, sort_order)
		VALUES ($1, $2, $3)
		RETURNING `+skillColumns,
		sk.Name, sk.Category, sk.SortOrder)
}

func (s *Store) UpdateSkill(ctx context.Context, id int64, sk SkillPatch) (Skill, error) {
	return queryOne[Skill](ctx, s, `
		UPDATE skills SET
			name       = coalesce($2, name),
			category   = coalesce($3, category),
			sort_order = coalesce($4, sort_order),
			updated_at = now()
		WHERE id = $1
		RETURNING `+skillColumns,
		id, sk.Name, sk.Category, sk.SortOrder)
}

func (s *Store) DeleteSkill(ctx context.Context, id int64) error {
	return s.deleteByID(ctx, "skills", id)
}

// queryOne runs a query that returns one row and maps it to T.
// It is a function, not a method, because Go methods cannot have type parameters.
func queryOne[T any](ctx context.Context, s *Store, sql string, args ...any) (T, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	return row, cvError(err)
}

// deleteByID deletes one row. table must be a constant, never user input.
func (s *Store) deleteByID(ctx context.Context, table string, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM `+table+` WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func cvError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23505" && pgErr.ConstraintName == "skills_name_key":
			return ErrSkillTaken
		case pgErr.Code == "23514" && (pgErr.TableName == "experience" || pgErr.TableName == "education"):
			return ErrDateOrder
		}
	}
	return err
}
