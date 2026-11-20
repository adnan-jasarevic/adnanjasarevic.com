package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

const maxContactBodyBytes = 16 << 10

type contactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (c *contactRequest) normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Email = strings.TrimSpace(c.Email)
	c.Message = strings.TrimSpace(c.Message)
}

func (c contactRequest) validate() map[string]string {
	errs := map[string]string{}

	switch {
	case c.Name == "":
		errs["name"] = "is required"
	case utf8.RuneCountInString(c.Name) > 100:
		errs["name"] = "must be at most 100 characters"
	}

	switch {
	case c.Email == "":
		errs["email"] = "is required"
	case len(c.Email) > 254 || !isEmail(c.Email):
		errs["email"] = "must be a valid email address"
	}

	switch {
	case c.Message == "":
		errs["message"] = "is required"
	case utf8.RuneCountInString(c.Message) > 5000:
		errs["message"] = "must be at most 5000 characters"
	}

	return errs
}

func isEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s
}

func (a *API) createContact(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxContactBodyBytes)

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.normalize()
	if errs := req.validate(); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": errs})
		return
	}

	err := a.db.CreateContactMessage(r.Context(), store.NewContactMessage{
		Name:    req.Name,
		Email:   req.Email,
		Message: req.Message,
	})
	if err != nil {
		log.Printf("create contact message: %v", err)
		writeError(w, http.StatusInternalServerError, "could not send message")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "received"})
}
