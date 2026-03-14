package response

import (
	"strings"
	"encoding/json"
	"net/http"
)

// APIResponse is a generic wrapper for API responses
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// ErrorInfo contains error details for problem responses
type ErrorInfo struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details ErrorDetails `json:"details,omitempty"`
}

type ErrorDetails struct {
	Field   string `json:"field,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ProblemDetails follows RFC 7807 format for problem responses
type ProblemDetails struct {
	Type     string            `json:"type,omitempty"`
	Title    string            `json:"title,omitempty"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail,omitempty"`
	Instance string            `json:"instance,omitempty"`
	Errors   map[string][]string `json:"errors,omitempty"`
}

// WriteJSON writes a JSON response with the proper content type
// Returns true if successful, false if the response was already committed
func WriteJSON(w http.ResponseWriter, status int, data any) bool {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	
	if data == nil {
		return true
	}
	
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data) == nil
}

// WriteJSONWithData writes a wrapped API response with data
func WriteJSONWithData(w http.ResponseWriter, status int, message string, data any) bool {
	result := APIResponse{
		Code:    status,
		Message: message,
		Data:    data,
	}
	return WriteJSON(w, status, result)
}

// WriteSuccess writes a 200 OK response with data
func WriteSuccess(w http.ResponseWriter, data any) bool {
	return WriteJSONWithData(w, http.StatusOK, "Success", data)
}

// WriteCreated writes a 201 Created response with the new resource
func WriteCreated(w http.ResponseWriter, data any) bool {
	w.Header().Set("Location", getLocationHeader(w))
	return WriteJSONWithData(w, http.StatusCreated, "Created", data)
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) bool {
	w.WriteHeader(http.StatusNoContent)
	return true
}

// WriteNotFound writes a 404 Not Found response
func WriteNotFound(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Resource not found"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusNotFound,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// WriteUnauthorized writes a 401 Unauthorized response
func WriteUnauthorized(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Authorization required"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// WriteForbidden writes a 403 Forbidden response
func WriteForbidden(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Access denied"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusForbidden,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// WriteBadRequest writes a 400 Bad Request response
func WriteBadRequest(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Bad request"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusBadRequest,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// WriteValidationError writes a 400 Bad Request response with field-specific validation errors
func WriteValidationError(w http.ResponseWriter, errors map[string][]string) bool {
	problem := ProblemDetails{
		Type:    "https://httpstatuses.com/400",
		Title:   "Validation Error",
		Status:  http.StatusBadRequest,
		Detail:  "One or more validation errors occurred",
		Errors:  errors,
	}
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(problem) == nil
}

// WriteConflict writes a 409 Conflict response
func WriteConflict(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Resource conflict"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusConflict,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusConflict)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// WriteInternalServerError writes a 500 Internal Server Error response
func WriteInternalServerError(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Internal server error"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// WriteError writes an error response with the appropriate status code
func WriteError(w http.ResponseWriter, statusCode int, message string) bool {
	switch statusCode {
	case http.StatusNotFound:
		return WriteNotFound(w, message)
	case http.StatusUnauthorized:
		return WriteUnauthorized(w, message)
	case http.StatusForbidden:
		return WriteForbidden(w, message)
	case http.StatusBadRequest:
		return WriteBadRequest(w, message)
	case http.StatusConflict:
		return WriteConflict(w, message)
	case http.StatusInternalServerError:
		return WriteInternalServerError(w, message)
	default:
		errInfo := ErrorInfo{
			Code:    statusCode,
			Message: message,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		return json.NewEncoder(w).Encode(errInfo) == nil
	}
}

// WriteProblem writes an RFC 7807 Problem Details response
func WriteProblem(w http.ResponseWriter, problem ProblemDetails) bool {
	contentType := "application/problem+json; charset=utf-8"
	if problem.Type == "" {
		contentType = "application/json; charset=utf-8"
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(problem.Status)
	return json.NewEncoder(w).Encode(problem) == nil
}

// getLocationHeader extracts the Location header value
func getLocationHeader(w http.ResponseWriter) string {
	return w.Header().Get("Location")
}

// WriteOk writes a 200 OK response with an optional message
func WriteOk(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "OK"
	}
	return WriteJSONWithData(w, http.StatusOK, message, nil)
}

// WriteNotImplemented writes a 501 Not Implemented response
func WriteNotImplemented(w http.ResponseWriter, message string) bool {
	if message == "" {
		message = "Not implemented"
	}
	errInfo := ErrorInfo{
		Code:    http.StatusNotImplemented,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotImplemented)
	return json.NewEncoder(w).Encode(errInfo) == nil
}

// RequiredHeaders sets headers that are required for the request
func RequiredHeaders(w http.ResponseWriter, headers []string) bool {
	w.Header().Set("X-Required-Headers", strings.Join(headers, ","))
	w.WriteHeader(http.StatusBadRequest)
	return true
}


// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Items       interface{} `json:"Items"`
	TotalCount  int         `json:"TotalRecordCount"`
	StartIndex  int         `json:"StartIndex"`
	Limit       int         `json:"Limit"`
}

func WritePaginated(w http.ResponseWriter, items interface{}, totalCount, startIndex, limit int) bool {
	resp := PaginatedResponse{
		Items:      items,
		TotalCount: totalCount,
		StartIndex: startIndex,
		Limit:      limit,
	}
	return WriteSuccess(w, resp)
}
