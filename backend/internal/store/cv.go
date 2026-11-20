package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Profile struct {
	FullName    string    `json:"fullName" db:"full_name"`
	Headline    string    `json:"headline" db:"headline"`
	Summary     string    `json:"summary" db:"summary"`
	Location    string    `json:"location" db:"location"`
	Email       *string   `json:"email" db:"email"`
	GithubURL   *string   `json:"githubUrl" db:"github_url"`
	LinkedinURL *string   `json:"linkedinUrl" db:"linkedin_url"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Experience struct {
	ID          int64      `json:"id" db:"id"`
	Company     string     `json:"company" db:"company"`
	Role        string     `json:"role" db:"role"`
	Location    string     `json:"location" db:"location"`
	StartDate   time.Time  `json:"startDate" db:"start_date"`
	EndDate     *time.Time `json:"endDate" db:"end_date"`
	Description string     `json:"description" db:"description"`
	Highlights  []string   `json:"highlights" db:"highlights"`
}

type Education struct {
	ID          int64      `json:"id" db:"id"`
	Institution string     `json:"institution" db:"institution"`
	Degree      string     `json:"degree" db:"degree"`
	Field       string     `json:"field" db:"field"`
	StartDate   time.Time  `json:"startDate" db:"start_date"`
	EndDate     *time.Time `json:"endDate" db:"end_date"`
	Description string     `json:"description" db:"description"`
}

type Skill struct {
	ID       int64  `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Category string `json:"category" db:"category"`
}

type CV struct {
	Profile    *Profile     `json:"profile"`
	Experience []Experience `json:"experience"`
	Education  []Education  `json:"education"`
	Skills     []Skill      `json:"skills"`
}

func (s *Store) GetCV(ctx context.Context) (CV, error) {
	batch := &pgx.Batch{}
	batch.Queue(`
		SELECT full_name, headline, summary, location, email, github_url, linkedin_url, updated_at
		FROM profile`)
	batch.Queue(`
		SELECT id, company, role, location, start_date, end_date, description, highlights
		FROM experience
		ORDER BY start_date DESC`)
	batch.Queue(`
		SELECT id, institution, degree, field, start_date, end_date, description
		FROM education
		ORDER BY start_date DESC`)
	batch.Queue(`
		SELECT id, name, category
		FROM skills
		ORDER BY category, sort_order, name`)

	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()

	var cv CV

	rows, err := results.Query()
	if err != nil {
		return CV{}, fmt.Errorf("query profile: %w", err)
	}
	profile, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Profile])
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return CV{}, fmt.Errorf("collect profile: %w", err)
	default:
		cv.Profile = &profile
	}

	rows, err = results.Query()
	if err != nil {
		return CV{}, fmt.Errorf("query experience: %w", err)
	}
	if cv.Experience, err = pgx.CollectRows(rows, pgx.RowToStructByName[Experience]); err != nil {
		return CV{}, fmt.Errorf("collect experience: %w", err)
	}

	rows, err = results.Query()
	if err != nil {
		return CV{}, fmt.Errorf("query education: %w", err)
	}
	if cv.Education, err = pgx.CollectRows(rows, pgx.RowToStructByName[Education]); err != nil {
		return CV{}, fmt.Errorf("collect education: %w", err)
	}

	rows, err = results.Query()
	if err != nil {
		return CV{}, fmt.Errorf("query skills: %w", err)
	}
	if cv.Skills, err = pgx.CollectRows(rows, pgx.RowToStructByName[Skill]); err != nil {
		return CV{}, fmt.Errorf("collect skills: %w", err)
	}

	return cv, nil
}
