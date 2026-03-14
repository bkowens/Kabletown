package response

import "net/http"

// serverID is set once at startup and included in all responses.
var serverID string

// SetServerID stores the server UUID that will be sent in X-MediaBrowser-Server-Id headers.
func SetServerID(id string) {
	serverID = id
}

// headerResponseWriter wraps http.ResponseWriter to inject required headers before
// the status code is written.
type headerResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (hw *headerResponseWriter) WriteHeader(code int) {
	if !hw.wroteHeader {
		hw.wroteHeader = true
		hw.ResponseWriter.Header().Set("X-Application-Version", "10.10.0")
		hw.ResponseWriter.Header().Set("X-MediaBrowser-Server-Id", serverID)
	}
	hw.ResponseWriter.WriteHeader(code)
}

func (hw *headerResponseWriter) Write(b []byte) (int, error) {
	if !hw.wroteHeader {
		hw.WriteHeader(http.StatusOK)
	}
	return hw.ResponseWriter.Write(b)
}

// RequiredHeaders is HTTP middleware that adds X-Application-Version and
// X-MediaBrowser-Server-Id to every response.
func RequiredHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&headerResponseWriter{ResponseWriter: w}, r)
	})
}
