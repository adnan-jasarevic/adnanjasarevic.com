package handlers

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

const maxCVBodyBytes = 64 << 10

type profileRequest struct {
	FullName    string `json:"fullName"`
	Headline    string `json:"headline"`
	Summary     string `json:"summary"`
	Location    string `json:"location"`
	Email       string `json:"email"`
	GithubURL   string `json:"githubUrl"`
	LinkedinURL string `json:"linkedinUrl"`
}

func (a *API) putProfile(w http.ResponseWriter, r *http.Request) {
	var req profileRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	trimAll(&req.FullName, &req.Headline, &req.Summary, &req.Location, &req.Email, &req.GithubURL, &req.LinkedinURL)

	errs := map[string]string{}
	requireText(errs, "fullName", &req.FullName, true, 100)
	limitText(errs, "headline", &req.Headline, 200)
	limitText(errs, "summary", &req.Summary, 5000)
	limitText(errs, "location", &req.Location, 100)
	if req.Email != "" && (len(req.Email) > 254 || !isEmail(req.Email)) {
		errs["email"] = "must be a valid email address"
	}
	if req.GithubURL != "" && !isWebURL(req.GithubURL) {
		errs["githubUrl"] = "must be an http or https URL"
	}
	if req.LinkedinURL != "" && !isWebURL(req.LinkedinURL) {
		errs["linkedinUrl"] = "must be an http or https URL"
	}
	if len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}

	profile, err := a.db.PutProfile(r.Context(), store.ProfileInput{
		FullName:    req.FullName,
		Headline:    req.Headline,
		Summary:     req.Summary,
		Location:    req.Location,
		Email:       nilIfEmpty(&req.Email),
		GithubURL:   nilIfEmpty(&req.GithubURL),
		LinkedinURL: nilIfEmpty(&req.LinkedinURL),
	})
	if err != nil {
		log.Printf("put profile: %v", err)
		writeError(w, http.StatusInternalServerError, "could not save profile")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// experienceRequest is used for both create and update. Nil fields were not sent.
// An empty endDate clears it.
type experienceRequest struct {
	Company     *string   `json:"company"`
	Role        *string   `json:"role"`
	Location    *string   `json:"location"`
	StartDate   *string   `json:"startDate"`
	EndDate     *string   `json:"endDate"`
	Description *string   `json:"description"`
	Highlights  *[]string `json:"highlights"`
}

func (e *experienceRequest) validate(creating bool) (dateRange, map[string]string) {
	trimAll(e.Company, e.Role, e.Location, e.StartDate, e.EndDate, e.Description)
	trimList(e.Highlights)

	errs := map[string]string{}
	requireText(errs, "company", e.Company, creating, 200)
	requireText(errs, "role", e.Role, creating, 200)
	limitText(errs, "location", e.Location, 100)
	limitText(errs, "description", e.Description, 5000)
	limitList(errs, "highlights", e.Highlights, 20, 500)
	dates := parseDateRange(errs, e.StartDate, e.EndDate, creating)
	return dates, errs
}

func (a *API) createExperience(w http.ResponseWriter, r *http.Request) {
	var req experienceRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	dates, errs := req.validate(true)
	if len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}
	e, err := a.db.CreateExperience(r.Context(), store.NewExperience{
		Company:     valueOrZero(req.Company),
		Role:        valueOrZero(req.Role),
		Location:    valueOrZero(req.Location),
		StartDate:   *dates.start,
		EndDate:     dates.end,
		Description: valueOrZero(req.Description),
		Highlights:  valueOrZero(req.Highlights),
	})
	if err != nil {
		writeCVError(w, err, "create", "experience")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (a *API) updateExperience(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req experienceRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	if req == (experienceRequest{}) {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	dates, errs := req.validate(false)
	if len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}
	e, err := a.db.UpdateExperience(r.Context(), id, store.ExperiencePatch{
		Company:      req.Company,
		Role:         req.Role,
		Location:     req.Location,
		StartDate:    dates.start,
		EndDate:      dates.end,
		ClearEndDate: dates.clearEnd,
		Description:  req.Description,
		Highlights:   req.Highlights,
	})
	if err != nil {
		writeCVError(w, err, "update", "experience")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// educationRequest works like experienceRequest.
type educationRequest struct {
	Institution *string `json:"institution"`
	Degree      *string `json:"degree"`
	Field       *string `json:"field"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
	Description *string `json:"description"`
}

func (e *educationRequest) validate(creating bool) (dateRange, map[string]string) {
	trimAll(e.Institution, e.Degree, e.Field, e.StartDate, e.EndDate, e.Description)

	errs := map[string]string{}
	requireText(errs, "institution", e.Institution, creating, 200)
	requireText(errs, "degree", e.Degree, creating, 200)
	limitText(errs, "field", e.Field, 200)
	limitText(errs, "description", e.Description, 5000)
	dates := parseDateRange(errs, e.StartDate, e.EndDate, creating)
	return dates, errs
}

func (a *API) createEducation(w http.ResponseWriter, r *http.Request) {
	var req educationRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	dates, errs := req.validate(true)
	if len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}
	e, err := a.db.CreateEducation(r.Context(), store.NewEducation{
		Institution: valueOrZero(req.Institution),
		Degree:      valueOrZero(req.Degree),
		Field:       valueOrZero(req.Field),
		StartDate:   *dates.start,
		EndDate:     dates.end,
		Description: valueOrZero(req.Description),
	})
	if err != nil {
		writeCVError(w, err, "create", "education")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (a *API) updateEducation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req educationRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	if req == (educationRequest{}) {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	dates, errs := req.validate(false)
	if len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}
	e, err := a.db.UpdateEducation(r.Context(), id, store.EducationPatch{
		Institution:  req.Institution,
		Degree:       req.Degree,
		Field:        req.Field,
		StartDate:    dates.start,
		EndDate:      dates.end,
		ClearEndDate: dates.clearEnd,
		Description:  req.Description,
	})
	if err != nil {
		writeCVError(w, err, "update", "education")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// skillRequest is used for both create and update. Nil fields were not sent.
type skillRequest struct {
	Name      *string `json:"name"`
	Category  *string `json:"category"`
	SortOrder *int    `json:"sortOrder"`
}

func (s *skillRequest) validate(creating bool) map[string]string {
	trimAll(s.Name, s.Category)

	errs := map[string]string{}
	requireText(errs, "name", s.Name, creating, 50)
	limitText(errs, "category", s.Category, 50)
	if s.SortOrder != nil && (*s.SortOrder < math.MinInt32 || *s.SortOrder > math.MaxInt32) {
		errs["sortOrder"] = "is out of range"
	}
	return errs
}

func (a *API) createSkill(w http.ResponseWriter, r *http.Request) {
	var req skillRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	if errs := req.validate(true); len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}
	s, err := a.db.CreateSkill(r.Context(), store.NewSkill{
		Name:      valueOrZero(req.Name),
		Category:  valueOrZero(req.Category),
		SortOrder: valueOrZero(req.SortOrder),
	})
	if err != nil {
		writeCVError(w, err, "create", "skill")
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (a *API) updateSkill(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req skillRequest
	if !decodeJSON(w, r, &req, maxCVBodyBytes) {
		return
	}
	if req == (skillRequest{}) {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	if errs := req.validate(false); len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}
	s, err := a.db.UpdateSkill(r.Context(), id, store.SkillPatch(req))
	if err != nil {
		writeCVError(w, err, "update", "skill")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// deleteCVItem returns a handler that deletes the item with the {id} path value.
func (a *API) deleteCVItem(what string, del func(ctx context.Context, id int64) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		if err := del(r.Context(), id); err != nil {
			writeCVError(w, err, "delete", what)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// writeCVError writes the response for an error from a CV store call and logs unexpected ones.
func writeCVError(w http.ResponseWriter, err error, action, what string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, what+" not found")
	case errors.Is(err, store.ErrDateOrder):
		writeFieldErrors(w, http.StatusBadRequest, map[string]string{"endDate": "must not be before startDate"})
	case errors.Is(err, store.ErrSkillTaken):
		writeFieldErrors(w, http.StatusConflict, map[string]string{"name": "is already taken"})
	default:
		log.Printf("%s %s: %v", action, what, err)
		writeError(w, http.StatusInternalServerError, "could not "+action+" "+what)
	}
}

type dateRange struct {
	start    *time.Time
	end      *time.Time
	clearEnd bool
}

// parseDateRange parses startDate and endDate (YYYY-MM-DD) and adds any problems to errs.
// With creating set, startDate is required. An empty endDate means it should be cleared.
func parseDateRange(errs map[string]string, start, end *string, creating bool) dateRange {
	var d dateRange
	switch {
	case start == nil && creating, start != nil && *start == "":
		errs["startDate"] = "is required"
	case start != nil:
		if t, err := time.Parse(time.DateOnly, *start); err == nil {
			d.start = &t
		} else {
			errs["startDate"] = "must be a date like 2024-09-01"
		}
	}
	switch {
	case end == nil:
	case *end == "":
		d.clearEnd = true
	default:
		if t, err := time.Parse(time.DateOnly, *end); err == nil {
			d.end = &t
		} else {
			errs["endDate"] = "must be a date like 2024-09-01"
		}
	}
	if d.start != nil && d.end != nil && d.end.Before(*d.start) {
		errs["endDate"] = "must not be before startDate"
	}
	return d
}

func trimAll(fields ...*string) {
	for _, s := range fields {
		if s != nil {
			*s = strings.TrimSpace(*s)
		}
	}
}

func trimList(items *[]string) {
	if items == nil {
		return
	}
	for i, s := range *items {
		(*items)[i] = strings.TrimSpace(s)
	}
}

// requireText checks a required field. It must be sent when creating, and can never be empty.
func requireText(errs map[string]string, field string, v *string, creating bool, max int) {
	switch {
	case v == nil && creating, v != nil && *v == "":
		errs[field] = "is required"
	default:
		limitText(errs, field, v, max)
	}
}

func limitText(errs map[string]string, field string, v *string, max int) {
	if v != nil && utf8.RuneCountInString(*v) > max {
		errs[field] = "must be at most " + strconv.Itoa(max) + " characters"
	}
}

func limitList(errs map[string]string, field string, items *[]string, maxItems, maxLen int) {
	if items == nil {
		return
	}
	if len(*items) > maxItems {
		errs[field] = "must have at most " + strconv.Itoa(maxItems) + " items"
		return
	}
	for _, s := range *items {
		if s == "" || utf8.RuneCountInString(s) > maxLen {
			errs[field] = "items must be 1 to " + strconv.Itoa(maxLen) + " characters"
			return
		}
	}
}
