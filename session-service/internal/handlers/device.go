package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/logger"
	"kabletown/session-service/internal/models"
)

var deviceLog = logger.NewLogger("device-handler")

type DeviceHandler struct {
	db  *sql.DB
}

func NewDeviceHandler(dbPool *sql.DB) *DeviceHandler {
	return &DeviceHandler{db: dbPool}
}

// GetDevices returns all devices for the authenticated user
func (h *DeviceHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	
	query := `SELECT id, name, user_id, last_user_id, date_last_activityed, app_name, app_version, 
		device_type, device_manufacturer, device_model, device_id, last_activity_date
		FROM devices WHERE user_id = ? OR last_user_id = ? ORDER BY last_activity_date DESC`
	
	rows, err := h.db.QueryContext(r.Context(), query, userID, userID)
	if err != nil {
		deviceLog.Error("Failed to query devices", "error", err, "user_id", userID)
		http.Error(w, "Failed to retrieve devices", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	devices := []models.Device{}
	for rows.Next() {
		var d models.Device
		var userID, lastUserID sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &userID, &lastUserID, &d.LastActivityDate, &d.AppName, &d.AppVersion,
			&d.DeviceType, &d.DeviceManufacturer, &d.DeviceModel, &d.DeviceID, &d.DateLastActivityed); err != nil {
			http.Error(w, "Failed to scan device", http.StatusInternalServerError)
			return
		}
		if userID.Valid {
			d.UserID = userID.String
		}
		if lastUserID.Valid {
			d.LastUserID = lastUserID.String
		}
		devices = append(devices, d)
	}

	render.JSON(w, r, devices)
}

// RegisterDevice registers a new device
func (h *DeviceHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterDeviceRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserFromContext(r.Context())
	if userID == "" {
		// For initial startup flow, device may not have user yet
		userID = ""
	}

	deviceID := req.DeviceID
	if deviceID == "" {
		http.Error(w, "Device ID required", http.StatusBadRequest)
		return
	}

	insertQuery := `INSERT INTO devices (id, device_id, name, user_id, last_user_id, app_name, app_version,
		device_type, device_manufacturer, device_model, last_activity_date, date_last_activityed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := h.db.ExecContext(r.Context(), insertQuery,
		req.DeviceID, req.DeviceID, req.Name, userID, userID, req.AppName, req.AppVersion,
		req.DeviceType, req.DeviceManufacturer, req.DeviceModel, time.Now(), time.Now())
	if err != nil {
		deviceLog.Error("Failed to register device", "error", err, "device_id", deviceID)
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]string{"deviceId": deviceID})
}

// UpdateDeviceName updates a device's name
func (h *DeviceHandler) UpdateDeviceName(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	
	var req models.UpdateDeviceNameRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updateQuery := `UPDATE devices SET name = ?, last_activity_date = ? WHERE device_id = ?`
	result, err := h.db.ExecContext(r.Context(), updateQuery, req.Name, time.Now(), deviceID)
	if err != nil {
		deviceLog.Error("Failed to update device name", "error", err, "device_id", deviceID)
		http.Error(w, "Failed to update device name", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteDevice removes a device
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	userID := auth.GetUserFromContext(r.Context())

	deleteQuery := `DELETE FROM devices WHERE device_id = ? AND user_id = ?`
	result, err := h.db.ExecContext(r.Context(), deleteQuery, deviceID, userID)
	if err != nil {
		deviceLog.Error("Failed to delete device", "error", err, "device_id", deviceID)
		http.Error(w, "Failed to delete device", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Device not found or unauthorized", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetDevice returns a specific device
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	
	query := `SELECT id, name, user_id, last_user_id, date_last_activityed, app_name, app_version,
		device_type, device_manufacturer, device_model, device_id, last_activity_date
		FROM devices WHERE device_id = ?`
	
	var d models.Device
	var userID, lastUserID sql.NullString
	err := h.db.QueryRowContext(r.Context(), query, deviceID).Scan(
		&d.ID, &d.Name, &userID, &lastUserID, &d.DateLastActivityed, &d.AppName, &d.AppVersion,
		&d.DeviceType, &d.DeviceManufacturer, &d.DeviceModel, &d.DeviceID, &d.LastActivityDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	if err != nil {
		deviceLog.Error("Failed to query device", "error", err)
		http.Error(w, "Failed to retrieve device", http.StatusInternalServerError)
		return
	}
	if userID.Valid {
		d.UserID = userID.String
	}
	if lastUserID.Valid {
		d.LastUserID = lastUserID.String
	}

	render.JSON(w, r, d)
}
