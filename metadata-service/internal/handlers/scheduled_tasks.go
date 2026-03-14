// Package handlers provides HTTP handlers for metadata service
package handlers

import (
	"github.com/jmoiron/sqlx"

	"net/http"

	"github.com/jellyfinhanced/shared/dto"
	"github.com/jellyfinhanced/shared/response"
)

// ItemRefreshTaskHandler handles scheduled task operations
type ItemRefreshTaskHandler struct {
	dbPool *sqlx.DB
}

// NewItemRefreshTaskHandler creates a new ItemRefreshTaskHandler
func NewItemRefreshTaskHandler(pool *sqlx.DB) *ItemRefreshTaskHandler {
	return &ItemRefreshTaskHandler{dbPool: pool}
}

// ListTasks handles GET /ScheduledTasks
// @Summary List scheduled tasks
// @Description List all scheduled metadata refresh tasks
// @Tags Scheduled Tasks
// @Success 200 {array} dto.ScheduledTaskDto
// @Router /ScheduledTasks [get]
func (h *ItemRefreshTaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// Query scheduled_tasks table
	response.WriteJSON(w, http.StatusOK, []dto.ScheduledTaskDto{})
}

// StartTask handles POST /ScheduledTasks/{id}/Start
// @Summary Start task
// @Description Manually start a scheduled task
// @Tags Scheduled Tasks
// @Param id path string true "Task ID"
// @Success 204
// @Router /ScheduledTasks/{id}/Start [post]
func (h *ItemRefreshTaskHandler) StartTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if taskID == "" {
		response.WriteBadRequest(w, "Task ID required")
		return
	}

	// Start the task
	response.WriteNoContent(w)
}