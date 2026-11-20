package handlers

import (
	"log"
	"net/http"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

func (a *API) getCV(w http.ResponseWriter, r *http.Request) {
	cv, err := a.db.GetCV(r.Context())
	if err != nil {
		log.Printf("get cv: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load cv")
		return
	}
	if cv.Experience == nil {
		cv.Experience = []store.Experience{}
	}
	if cv.Education == nil {
		cv.Education = []store.Education{}
	}
	if cv.Skills == nil {
		cv.Skills = []store.Skill{}
	}
	writeJSON(w, http.StatusOK, cv)
}
