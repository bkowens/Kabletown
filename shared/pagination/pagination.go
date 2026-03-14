package pagination

// PaginationParams holds common pagination query parameters
type PaginationParams struct {
	StartIndex    int    `json:"StartIndex" uri:"startIndex"`
	Limit         int    `json:"Limit" uri:"limit"`
	Recursive     bool   `json:"Recursive" uri:"recursive"`
	SearchTerm    string `json:"SearchTerm" uri:"searchTerm"`
	IncludeItemTypes string `json:"IncludeItemTypes" uri:"includeItemTypes"`
	ExcludeItemTypes   string `json:"ExcludeItemTypes" uri:"excludeItemTypes"`
}

// DefaultPagination returns default pagination values
func DefaultPagination() PaginationParams {
	return PaginationParams{
		StartIndex: 0,
		Limit:      20,
		Recursive:  true,
	}
}

// HasFilters checks if any filter is set beyond pagination
func (p PaginationParams) HasFilters() bool {
	return p.SearchTerm != "" || p.IncludeItemTypes != "" || p.ExcludeItemTypes != ""
}

// SQLLimit returns a properly validated LIMIT clause value
func (p PaginationParams) SQLLimit() int {
	const maxLimit = 1000
	if p.Limit <= 0 || p.Limit > maxLimit {
		return maxLimit
	}
	return p.Limit
}

// OFFSET returns the OFFSET clause value
func (p PaginationParams) SQLOffset() int {
	if p.StartIndex < 0 {
		return 0
	}
	return p.StartIndex
}

// TotalCountWrapper wraps a result set with total count info
type TotalCountWrapper struct {
	Items           interface{} `json:"Items"`
	TotalRecordCount int         `json:"TotalRecordCount"`
	StartIndex      int         `json:"StartIndex"`
	Limit           int         `json:"Limit"`
}

// WrapResult creates a paginated response wrapper
func WrapResult(items interface{}, totalCount, startIndex, limit int) TotalCountWrapper {
	if totalCount < 0 {
		totalCount = 0
	}
	return TotalCountWrapper{
		Items:           items,
		TotalRecordCount: totalCount,
		StartIndex:      startIndex,
		Limit:           limit,
	}
}
