package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	unreadOnly := r.URL.Query().Get("unread") == "true"
	messages, err := a.db.ListContactMessages(r.Context(), unreadOnly)
	if err != nil {
		log.Printf("list contact messages: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load messages")
		return
	}
	if messages == nil {
		messages = []store.ContactMessage{}
	}
	writeJSON(w, http.StatusOK, messages)
}

func (a *API) updateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req struct {
		Read *bool `json:"read"`
	}
	if !decodeJSON(w, r, &req, 1<<10) {
		return
	}
	if req.Read == nil {
		writeFieldErrors(w, http.StatusBadRequest, map[string]string{"read": "is required"})
		return
	}

	err := a.db.SetContactMessageRead(r.Context(), id, *req.Read)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		log.Printf("update contact message %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "could not update message")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) deleteMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	err := a.db.DeleteContactMessage(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		log.Printf("delete contact message %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "could not delete message")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
