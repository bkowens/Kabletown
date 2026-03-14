// Package healthcheck provides service health monitoring and diagnostics
package healthcheck

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

// HealthStatus represents the health status of a service or component
type HealthStatus struct {
	Service   string            `json:"service"`
	Status    string            `json:"status"` // healthy, unhealthy, degraded
	Message   string            `json:"message,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Duration  time.Duration     `json:"duration_ms"`
	Details   map[string]string `json:"details,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// ComponentHealth represents individual component health status
type ComponentHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Check database connectivity and return health status
func CheckDB(db *sql.DB) HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := db.PingContext(ctx)
	duration := time.Since(start)

	status := HealthStatus{
		Service:   "database",
		Timestamp: time.Now(),
		Duration:  duration,
		Details: map[string]string{
			"duration_ms": fmt.Sprintf("%d", duration.Milliseconds()),
		},
	}

	if err != nil {
		status.Status = "unhealthy"
		status.Message = "Failed to connect to database"
		status.Error = err.Error()
	} else {
		status.Status = "healthy"
	}

	return status
}

// CheckHTTP checks HTTP connectivity to a target service
func CheckHTTP(target string) ComponentHealth {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return ComponentHealth{
			Name:    target,
			Status:  "unhealthy",
			Message: "Failed to create request: " + err.Error(),
		}
	}

	_, err = client.Do(req)
	if err != nil {
		return ComponentHealth{
			Name:    target,
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return ComponentHealth{
		Name:    target,
		Status:  "healthy",
		Message: "OK",
	}
}

// Check database connectivity and return health status