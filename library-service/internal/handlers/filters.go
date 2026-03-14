// Filter handlers for library service.
package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jmoiron/sqlx"

	"github.com/jellyfinhanced/library-service/internal/dto"
)

// FilterHandler handles genre/studio/person filter endpoints.
type FilterHandler struct {
	db *sqlx.DB
}

// NewFilterHandler creates a new FilterHandler.
func NewFilterHandler(db *sqlx.DB) *FilterHandler {
	return &FilterHandler{db: db}
}

// GetGenres handles GET /Genres.
func (h *FilterHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `SELECT DISTINCT Genre FROM item_genres ORDER BY Genre`)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get genres")
		return
	}
	defer rows.Close()

	var genres []dto.GenreDto
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		genres = append(genres, dto.GenreDto{Name: name})
	}
	if genres == nil {
		genres = []dto.GenreDto{}
	}
	response.WriteJSON(w, http.StatusOK, genres)
}

// GetStudios handles GET /Studios.
func (h *FilterHandler) GetStudios(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `SELECT DISTINCT Studio FROM item_studios ORDER BY Studio`)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get studios")
		return
	}
	defer rows.Close()

	var studios []dto.StudioDto
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		studios = append(studios, dto.StudioDto{Name: name})
	}
	if studios == nil {
		studios = []dto.StudioDto{}
	}
	response.WriteJSON(w, http.StatusOK, studios)
}

// GetPersons handles GET /Persons.
func (h *FilterHandler) GetPersons(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `SELECT PersonId, Name FROM item_people ORDER BY Name ASC`)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get persons")
		return
	}
	defer rows.Close()

	var persons []dto.PersonDto
	for rows.Next() {
		var p dto.PersonDto
		if err := rows.Scan(&p.Id, &p.Name); err != nil {
			continue
		}
		persons = append(persons, p)
	}
	if persons == nil {
		persons = []dto.PersonDto{}
	}
	response.WriteJSON(w, http.StatusOK, persons)
}

// GetYears handles GET /Years.
func (h *FilterHandler) GetYears(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT DISTINCT ProductionYear FROM base_items WHERE ProductionYear IS NOT NULL ORDER BY ProductionYear ASC`)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get years")
		return
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var year int
		if err := rows.Scan(&year); err != nil {
			continue
		}
		years = append(years, year)
	}
	if years == nil {
		years = []int{}
	}
	response.WriteJSON(w, http.StatusOK, years)
}

// RegisterFilterRoutes registers filter routes on the given mux.
func RegisterFilterRoutes(mux *http.ServeMux, h *FilterHandler) {
	mux.HandleFunc("GET /Genres", h.GetGenres)
	mux.HandleFunc("GET /Studios", h.GetStudios)
	mux.HandleFunc("GET /Persons", h.GetPersons)
	mux.HandleFunc("GET /Years", h.GetYears)
}
