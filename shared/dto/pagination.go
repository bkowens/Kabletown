// Package dto contains shared data transfer objects used across Kabletown services.
package dto

// PagedResult[T] is a generic wrapper for paginated query results
type PagedResult[T any] struct {
	// Items is the slice of results for this page
	Items []T `json:"Items"`

	// TotalRecordCount is the total number of matching records across all pages
	TotalRecordCount int `json:"TotalRecordCount"`

	// StartIndex is the 0-based index of the first item in this page
	StartIndex int `json:"StartIndex"`

	// Limit is the maximum number of items per page (may be less for the last page)
	Limit int `json:"Limit"`

	// IsLastPage is true if there are no more pages after this one
	IsLastPage bool `json:"IsLastPage"`
}

// PagedResultOpts configures pagination options
type PagedResultOpts struct {
	// StartIndex is the 0-based index of the first item to return
	StartIndex int `json:"StartIndex"`

	// Limit is the maximum number of items to return
	// If <= 0, a default of 1000 is used for safety
	Limit int `json:"Limit"`

	// IncludeTotal determines whether to calculate TotalRecordCount
	// Setting false improves performance when you don't need the total
	IncludeTotal bool `json:"IncludeTotal"`

	// SortBy is the field name to sort by
	SortBy string `json:"SortBy,omitempty"`

	// SortOrder is "Ascending" or "Descending"
	SortOrder string `json:"SortOrder,omitempty"`
}

// DefaultPagedResultOpts returns sensible defaults
func DefaultPagedResultOpts() PagedResultOpts {
	return PagedResultOpts{
		StartIndex:   0,
		Limit:        20,
		IncludeTotal: true,
		SortOrder:    "Ascending",
	}
}

// NewPagedResult creates a new PagedResult with computed values
func NewPagedResult[T any](items []T, totalCount, startIndex, limit int) PagedResult[T] {
	if items == nil {
		items = []T{}
	}
	if totalCount < 0 {
		totalCount = 0
	}
	
	// Calculate if this is the last page
	isLastPage := startIndex+limit >= totalCount || len(items) < limit
	
	// Clamp startIndex if it exceeds the total
	if startIndex >= totalCount && totalCount > 0 {
		startIndex = totalCount - 1
		if startIndex < 0 {
			startIndex = 0
		}
	}
	
	return PagedResult[T]{
		Items:            items,
		TotalRecordCount: totalCount,
		StartIndex:       startIndex,
		Limit:            limit,
		IsLastPage:       isLastPage,
	}
}

// HasNextPage returns true if there are more pages after this one
func (p PagedResult[T]) HasNextPage() bool {
	return !p.IsLastPage
}

// NextStartIndex returns the StartIndex for the next page
func (p PagedResult[T]) NextStartIndex() int {
	return p.StartIndex + p.Limit
}

// NextPageOpts returns options configured for the next page
func (p PagedResult[T]) NextPageOpts(opts PagedResultOpts) PagedResultOpts {
	opts.StartIndex = p.NextStartIndex()
	return opts
}

// SliceRange returns a slice representing the item range for display
func (p PagedResult[T]) SliceRange() string {
	if len(p.Items) == 0 {
		return "0-0 of 0"
	}
	
	end := p.StartIndex + len(p.Items)
	display := end
	if display > p.TotalRecordCount {
		display = p.TotalRecordCount
	}
	
	return string(p.StartIndex) + "-" + string(display) + " of " + string(p.TotalRecordCount)
}

// Map transforms a PagedResult[T] to a PagedResult[U]
func MapPagedResult[T any, U any](p PagedResult[T], fn func(T) U) PagedResult[U] {
	items := make([]U, len(p.Items))
	for i, item := range p.Items {
		items[i] = fn(item)
	}
	
	return PagedResult[U]{
		Items:            items,
		TotalRecordCount: p.TotalRecordCount,
		StartIndex:       p.StartIndex,
		Limit:            p.Limit,
		IsLastPage:       p.IsLastPage,
	}
}

// Filter filters items based on a predicate
func (p PagedResult[T]) Filter(fn func(T) bool) PagedResult[T] {
	filtered := make([]T, 0, len(p.Items))
	for _, item := range p.Items {
		if fn(item) {
			filtered = append(filtered, item)
		}
	}
	
	return PagedResult[T]{
		Items:            filtered,
		TotalRecordCount: len(filtered), // Adjusted to actual filtered count
		StartIndex:       p.StartIndex,
		Limit:            p.Limit,
		IsLastPage:       len(filtered) < p.Limit,
	}
}