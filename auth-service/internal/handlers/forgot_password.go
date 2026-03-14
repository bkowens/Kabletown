package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bowens/kabletown/shared/response"
)

// ForgotPassword handles POST /Users/ForgotPassword.
// Body: {"EnteredUsername":"..."}
// Always returns a ContactAdmin action — password resets require administrator intervention.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// Drain the body gracefully; we do not use the username.
	var body interface{}
	json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"Action":            "ContactAdmin",
		"PinFile":           nil,
		"PinExpirationDate": nil,
	})
}

// ForgotPasswordPin handles POST /Users/ForgotPasswordPin.
// Body: {"EnteredPin":"..."}
// Always returns failure — PIN-based reset is not supported.
func (h *Handler) ForgotPasswordPin(w http.ResponseWriter, r *http.Request) {
	// Drain the body gracefully.
	var body interface{}
	json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck

	response.JSON(w, http.StatusOK, map[string]bool{
		"Success":   false,
		"IsInError": true,
	})
}
