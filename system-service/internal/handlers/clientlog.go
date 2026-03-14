package handlers

import (
	"net/http"
)

// SubmitClientLog handles POST /ClientLog/Document.
func (h *Handler) SubmitClientLog(w http.ResponseWriter, r *http.Request) {
	// Accept and discard client log submissions
	w.WriteHeader(http.StatusOK)
}
