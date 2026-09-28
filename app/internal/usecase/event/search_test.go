package event_test

import (
	"context"
	"errors"
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

func TestSearchExecuteSuggestsRefinementInsteadOfReturningTooManyResults(t *testing.T) {
	t.Parallel()

	events := make([]eventmodel.Event, 31)
	for i := range events {
		events[i].Title = "Event"
	}
	provider := &providerStub{
		events:     events,
		totalCount: 42,
	}

	result, err := eventusecase.NewSearch(provider).Execute(t.Context(), eventmodel.SearchCondition{})
	if err != nil {
		t.Fatalf("execute search: %v", err)
	}

	if len(result.Events) != 0 {
		t.Errorf("event count = %d, want 0", len(result.Events))
	}
	if result.TotalCount != 42 {
		t.Errorf("total count = %d, want 42", result.TotalCount)
	}
	if result.Message == "" {
		t.Error("message is empty, want refinement guidance")
	}
}

func TestSearchExecuteReturnsProviderError(t *testing.T) {
	t.Parallel()

	providerErr := errors.New("provider unavailable")
	provider := &providerStub{err: providerErr}

	_, err := eventusecase.NewSearch(provider).Execute(t.Context(), eventmodel.SearchCondition{})
	if !errors.Is(err, providerErr) {
		t.Fatalf("execute search error = %v, want wrapped provider error", err)
	}
}

type providerStub struct {
	events     []eventmodel.Event
	totalCount int
	err        error
	condition  eventmodel.SearchCondition
}

func (p *providerStub) Search(_ context.Context, condition eventmodel.SearchCondition) ([]eventmodel.Event, int, error) {
	p.condition = condition

	return p.events, p.totalCount, p.err
}
