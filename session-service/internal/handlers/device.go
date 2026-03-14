package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/shared/types"
	"github.com/jellyfinhanced/session-service/internal/models"
)

// DeviceHandler handles device-related requests.
type DeviceHandler struct {
	db *sql.DB
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(dbPool *sql.DB) *DeviceHandler {
	return &DeviceHandler{db: dbPool}
}

// GetDevices returns all devices for the authenticated user.
func (h *DeviceHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	authInfo, ok := auth.GetAuth(r.Context())
	if !ok || authInfo == nil {
		response.WriteUnauthorized(w, "Authentication required")
		return
	}

	userID := authInfo.UserID.String()

	query := `SELECT id, name, user_id, last_user_id, date_last_activityed, app_name, app_version,
		device_type, device_manufacturer, device_model, device_id, last_activity_date
		FROM devices WHERE user_id = ? OR last_user_id = ? ORDER BY last_activity_date DESC`

	rows, err := h.db.QueryContext(r.Context(), query, userID, userID)
	if err != nil {
		log.Printf("device-handler: GetDevices: query failed: %v", err)
		response.WriteInternalServerError(w, "Failed to retrieve devices")
		return
	}
	defer rows.Close()

	devices := []models.Device{}
	for rows.Next() {
		var d models.Device
		var uid, lastUID sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &uid, &lastUID, &d.DateLastActivityed, &d.AppName, &d.AppVersion,
			&d.DeviceType, &d.DeviceManufacturer, &d.DeviceModel, &d.DeviceID, &d.LastActivityDate); err != nil {
			response.WriteInternalServerError(w, "Failed to scan device")
			return
		}
		if uid.Valid {
			d.UserID = uid.String
		}
		if lastUID.Valid {
			d.LastUserID = lastUID.String
		}
		devices = append(devices, d)
	}

	response.WriteJSON(w, http.StatusOK, devices)
}

// GetDevicesByAccessToken queries devices by AccessToken support (header → user → ApiKeys fallback)
func (h *DeviceHandler) GetDevicesByAccessToken(w http.ResponseWriter, r *http.Request) {
	accessToken := r.URL.Query().Get("AccessToken")
	if accessToken == "" {
		accessToken = r.URL.Query().Get("access_token")
	}

	if accessToken == "" {
		response.WriteBadRequest(w, "AccessToken query parameter required")
		return
	}

	var devices []models.Device

	// Step 1: Try to find device in api_keys where AccessToken is the token
	apiKeyQuery := `
		SELECT d.id, d.name, d.user_id, d.last_user_id, d.date_last_activityed,
			d.app_name, d.app_version, d.device_type, d.device_manufacturer,
			d.device_model, d.device_id, d.last_activity_date
		FROM devices d
		INNER JOIN api_keys k ON d.user_id = k.user_id
		WHERE k.token = ?
		ORDER BY d.last_activity_date DESC
	`

	rows, err := h.db.QueryContext(r.Context(), apiKeyQuery, types.HashToken(accessToken))
	if err != nil {
		log.Printf("device-handler: GetDevicesByAccessToken: api_keys query failed: %v", err)
		response.WriteInternalServerError(w, "Failed to query devices")
		return
	}
	
	for rows.Next() {
		var d models.Device
		var uid, lastUID sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &uid, &lastUID, &d.DateLastActivityed,
			&d.AppName, &d.AppVersion, &d.DeviceType, &d.DeviceManufacturer,
			&d.DeviceModel, &d.DeviceID, &d.LastActivityDate); err != nil {
			log.Printf("device-handler: GetDevicesByAccessToken: scan failed: %v", err)
			response.WriteInternalServerError(w, "Failed to scan device")
			rows.Close()
			return
		}
		if uid.Valid {
			d.UserID = uid.String
		}
		if lastUID.Valid {
			d.LastUserID = lastUID.String
		}
		devices = append(devices, d)
	}
	rows.Close()

	// No devices found - check if AccessToken exists at all (for proper 404 vs 401)
	if len(devices) == 0 {
		notifyExistsUserQuery := `
			SELECT u.Id FROM users u WHERE u.AccessToken = ?
		`
		var userId string
		err = h.db.QueryRowContext(r.Context(), notifyExistsUserQuery, accessToken).Scan(&userId)
		if err == nil {
			// User exists but no devices - return empty array (not 404)
			response.WriteJSON(w, http.StatusOK, []models.Device{})
			return
		}
	}

	log.Printf("device-handler: GetDevicesByAccessToken: found %d devices", len(devices))
	response.WriteJSON(w, http.StatusOK, devices)
}

// RegisterDevice registers a new device.
func (h *DeviceHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "Invalid request body")
		return
	}

	authInfo, ok := auth.GetAuth(r.Context())
	if !ok || authInfo == nil {
		response.WriteUnauthorized(w, "Authentication required")
		return
	}

	userID := authInfo.UserID.String()

	deviceID := req.DeviceID
	if deviceID == "" {
		response.WriteBadRequest(w, "Device ID required")
		return
	}

	insertQuery := `INSERT INTO devices (id, device_id, name, user_id, last_user_id, app_name, app_version,
		device_type, device_manufacturer, device_model, last_activity_date, date_last_activityed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := h.db.ExecContext(r.Context(), insertQuery,
		req.DeviceID, req.DeviceID, req.Name, userID, userID, req.AppName, req.AppVersion,
		req.DeviceType, req.DeviceManufacturer, req.DeviceModel, time.Now(), time.Now())
	if err != nil {
		log.Printf("device-handler: RegisterDevice: insert failed: %v", err)
		response.WriteInternalServerError(w, "Failed to register device")
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]string{"DeviceId": deviceID})
}

// UpdateDeviceName updates a device's name.
func (h *DeviceHandler) UpdateDeviceName(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")

	var req models.UpdateDeviceNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "Invalid request body")
		return
	}

	updateQuery := `UPDATE devices SET name = ?, last_activity_date = ? WHERE device_id = ?`
	result, err := h.db.ExecContext(r.Context(), updateQuery, req.Name, time.Now(), deviceID)
	if err != nil {
		log.Printf("device-handler: UpdateDeviceName: update failed: %v", err)
		response.WriteInternalServerError(w, "Failed to update device name")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response.WriteNotFound(w, "Device not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteDevice removes a device.
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("Id")
	if deviceID == "" {
		deviceID = chi.URLParam(r, "id")
	}
	
	authInfo, ok := auth.GetAuth(r.Context())
	if !ok || authInfo == nil {
		response.WriteUnauthorized(w, "Authentication required")
		return
	}

	userID := authInfo.UserID.String()

	deleteQuery := `DELETE FROM devices WHERE device_id = ? AND user_id = ?`
	result, err := h.db.ExecContext(r.Context(), deleteQuery, deviceID, userID)
	if err != nil {
		log.Printf("device-handler: DeleteDevice: delete failed: %v", err)
		response.WriteInternalServerError(w, "Failed to delete device")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response.WriteNotFound(w, "Device not found or unauthorized")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDevice returns a specific device.
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")

	query := `SELECT id, name, user_id, last_user_id, date_last_activityed, app_name, app_version,
		device_type, device_manufacturer, device_model, device_id, last_activity_date
		FROM devices WHERE device_id = ?`

	var d models.Device
	var uid, lastUID sql.NullString
	err := h.db.QueryRowContext(r.Context(), query, deviceID).Scan(
		&d.ID, &d.Name, &uid, &lastUID, &d.DateLastActivityed, &d.AppName, &d.AppVersion,
		&d.DeviceType, &d.DeviceManufacturer, &d.DeviceModel, &d.DeviceID, &d.LastActivityDate)
	if err == sql.ErrNoRows {
		response.WriteNotFound(w, "Device not found")
		return
	}
	if err != nil {
		log.Printf("device-handler: GetDevice: query failed: %v", err)
		response.WriteInternalServerError(w, "Failed to retrieve device")
		return
	}
	if uid.Valid {
		d.UserID = uid.String
	}
	if lastUID.Valid {
		d.LastUserID = lastUID.String
	}

	response.WriteJSON(w, http.StatusOK, d)
}

// RegisterRoutes adds all device routes to the router
func RegisterRoutes(r chi.Router, handler *DeviceHandler) {
	r.Route("/Devices", func(r chi.Router) {
		r.Get("/", handler.GetDevices)
		r.Post("/", handler.RegisterDevice)
		r.Get("/ByAccessToken", handler.GetDevicesByAccessToken)
		r.Get("/{id}", handler.GetDevice)
		r.Post("/{id}/Name", handler.UpdateDeviceName)
		r.Delete("/{id}", handler.DeleteDevice)
	})
}
