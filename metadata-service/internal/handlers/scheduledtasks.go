package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var scheduledTaskLog = logger.NewLogger("scheduled-task-handler")

type ScheduledTaskHandler struct {
	db  *sql.DB
}

func NewScheduledTaskHandler(dbPool *sql.DB) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{db: dbPool}
}

// ListTasks returns all scheduled tasks
func (h *ScheduledTaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, state, last_run_start_time, last_run_end_time, run_times
		FROM scheduled_tasks`
	
	rows, err := h.db.QueryContext(r.Context(), query)
	if err != nil {
		scheduledTaskLog.Error("Failed to list tasks", "error", err)
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []map[string]interface{}{}
	for rows.Next() {
		var t map[string]interface{}
		var lastStart, lastEnd sql.NullTime
		var runTimes sql.NullString
		err := rows.Scan(&t["id"], &t["name"], &t["state"], &lastStart, &lastEnd, &runTimes)
		if err != nil {
			continue
		}
		if lastStart.Valid {
			t["last_run_start_time"] = lastStart.Time.String()
		}
		if lastEnd.Valid {
			t["last_run_end_time"] = lastEnd.Time.String()
		}
		if runTimes.Valid {
			t["run_times"] = runTimes.String
		}
		tasks = append(tasks, t)
	}

	render.JSON(w, r, tasks)
}

// GetTask returns a specific task
func (h *ScheduledTaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	
	query := `SELECT id, name, state, last_run_start_time, last_run_end_time, run_times
		FROM scheduled_tasks WHERE id = ?`
	
	var t map[string]interface{}
	var lastStart, lastEnd sql.NullTime
	var runTimes sql.NullString
	err := h.db.QueryRowContext(r.Context(), query, taskID).Scan(
		&t["id"], &t["name"], &t["state"], &lastStart, &lastEnd, &runTimes)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		scheduledTaskLog.Error("Failed to query task", "error", err)
		http.Error(w, "Failed to retrieve task", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, t)
}

// StartTask starts a scheduled task
func (h *ScheduledTaskHandler) StartTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	
	// Start task and record start time
	query := `UPDATE scheduled_tasks 
		SET state = 'Running', last_run_start_time = ? 
		WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, time.Now(), taskID)
	if err != nil {
		scheduledTaskLog.Error("Failed to start task", "error", err)
		http.Error(w, "Failed to start task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// StopTask stops a scheduled task
func (h *ScheduledTaskHandler) StopTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	
	query := `UPDATE scheduled_tasks 
		SET state = 'Idle', last_run_end_time = ? 
		WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, time.Now(), taskID)
	if err != nil {
		scheduledTaskLog.Error("Failed to stop task", "error", err)
		http.Error(w, "Failed to stop task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
