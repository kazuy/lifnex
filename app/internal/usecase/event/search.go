package event

import (
	"context"
	"errors"
	"fmt"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
)

const (
	maxSearchResults        = 30
	refinementMessageFormat = "Showing the first %d of %d matching events. Narrow the search further to find more relevant events."
)

var (
	// ErrDateRangeRequired indicates that an event search is missing a date boundary.
	ErrDateRangeRequired = errors.New("event search requires both from and to")
	// ErrInvalidDateRange indicates that an event search starts after it ends.
	ErrInvalidDateRange = errors.New("event search from must not be after to")
)

// SearchResult contains events and metadata needed by callers to present them.
type SearchResult struct {
	Events     []eventmodel.Event
	TotalCount int
	Message    string
}

// Search finds events through a Provider and applies application-level result limits.
type Search struct {
	provider Provider
}

// NewSearch creates an event search use case.
func NewSearch(provider Provider) *Search {
	return &Search{provider: provider}
}

// Execute searches for events matching condition.
func (s *Search) Execute(ctx context.Context, condition eventmodel.SearchCondition) (SearchResult, error) {
	if condition.From.IsZero() || condition.To.IsZero() {
		return SearchResult{}, ErrDateRangeRequired
	}
	if condition.From.After(condition.To) {
		return SearchResult{}, ErrInvalidDateRange
	}

	events, totalCount, err := s.provider.Search(ctx, condition)
	if err != nil {
		return SearchResult{}, fmt.Errorf("search events: %w", err)
	}

	if totalCount > maxSearchResults {
		if len(events) > maxSearchResults {
			events = events[:maxSearchResults]
		}

		return SearchResult{
			Events:     events,
			TotalCount: totalCount,
			Message:    fmt.Sprintf(refinementMessageFormat, len(events), totalCount),
		}, nil
	}

	return SearchResult{
		Events:     events,
		TotalCount: totalCount,
	}, nil
}
