package event_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
)

func TestSearchExecute(t *testing.T) {
	t.Parallel()

	condition := eventmodel.SearchCondition{
		From:      time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
		Keyword:   "音楽",
		Locations: []string{"川崎区", "幸区"},
	}
	events := []eventmodel.Event{
		{Title: "Event 1"},
		{Title: "Event 2"},
	}
	provider := &providerStub{
		events:     events,
		totalCount: len(events),
	}

	result, err := eventusecase.NewSearch(provider).Execute(t.Context(), condition)
	if err != nil {
		t.Fatalf("execute search: %v", err)
	}

	if !reflect.DeepEqual(provider.condition, condition) {
		t.Errorf("provider condition = %#v, want %#v", provider.condition, condition)
	}
	if !reflect.DeepEqual(result.Events, events) {
		t.Errorf("events = %#v, want %#v", result.Events, events)
	}
	if result.TotalCount != len(events) {
		t.Errorf("total count = %d, want %d", result.TotalCount, len(events))
	}
	if result.Message != "" {
		t.Errorf("message = %q, want empty", result.Message)
	}
}

func TestSearchExecuteAcceptsSameDayRange(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.October, 10, 0, 0, 0, 0, time.UTC)
	condition := eventmodel.SearchCondition{From: date, To: date}
	provider := &providerStub{}

	_, err := eventusecase.NewSearch(provider).Execute(t.Context(), condition)
	if err != nil {
		t.Fatalf("execute search: %v", err)
	}

	if !provider.called {
		t.Error("provider was not called")
	}
}

func TestSearchExecuteRejectsInvalidDateRange(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.October, 10, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		condition eventmodel.SearchCondition
		wantErr   error
	}{
		{
			name:    "missing both boundaries",
			wantErr: eventusecase.ErrDateRangeRequired,
		},
		{
			name:      "missing from",
			condition: eventmodel.SearchCondition{To: date},
			wantErr:   eventusecase.ErrDateRangeRequired,
		},
		{
			name:      "missing to",
			condition: eventmodel.SearchCondition{From: date},
			wantErr:   eventusecase.ErrDateRangeRequired,
		},
		{
			name: "from after to",
			condition: eventmodel.SearchCondition{
				From: date.AddDate(0, 0, 1),
				To:   date,
			},
			wantErr: eventusecase.ErrInvalidDateRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := &providerStub{}
			_, err := eventusecase.NewSearch(provider).Execute(t.Context(), tt.condition)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("execute search error = %v, want %v", err, tt.wantErr)
			}
			if provider.called {
				t.Error("provider was called for an invalid condition")
			}
		})
	}
}

func TestSearchExecuteReturnsLimitedResultsWithRefinementMessage(t *testing.T) {
	t.Parallel()

	events := make([]eventmodel.Event, 42)
	for i := range events {
		events[i].Title = fmt.Sprintf("Event %d", i+1)
	}
	provider := &providerStub{
		events:     events,
		totalCount: 42,
	}

	date := time.Date(2026, time.October, 10, 0, 0, 0, 0, time.UTC)
	result, err := eventusecase.NewSearch(provider).Execute(t.Context(), eventmodel.SearchCondition{From: date, To: date})
	if err != nil {
		t.Fatalf("execute search: %v", err)
	}

	if len(result.Events) != 30 {
		t.Errorf("event count = %d, want 30", len(result.Events))
	}
	if result.Events[29].Title != "Event 30" {
		t.Errorf("last event title = %q, want %q", result.Events[29].Title, "Event 30")
	}
	if result.TotalCount != 42 {
		t.Errorf("total count = %d, want 42", result.TotalCount)
	}
	wantMessage := "Showing the first 30 of 42 matching events. Narrow the search further to find more relevant events."
	if result.Message != wantMessage {
		t.Errorf("message = %q, want %q", result.Message, wantMessage)
	}
}

func TestSearchExecuteReturnsProviderError(t *testing.T) {
	t.Parallel()

	providerErr := errors.New("provider unavailable")
	provider := &providerStub{err: providerErr}

	date := time.Date(2026, time.October, 10, 0, 0, 0, 0, time.UTC)
	_, err := eventusecase.NewSearch(provider).Execute(t.Context(), eventmodel.SearchCondition{From: date, To: date})
	if !errors.Is(err, providerErr) {
		t.Fatalf("execute search error = %v, want wrapped provider error", err)
	}
}

type providerStub struct {
	events     []eventmodel.Event
	totalCount int
	err        error
	condition  eventmodel.SearchCondition
	called     bool
}

func (p *providerStub) Search(_ context.Context, condition eventmodel.SearchCondition) ([]eventmodel.Event, int, error) {
	p.called = true
	p.condition = condition

	return p.events, p.totalCount, p.err
}
