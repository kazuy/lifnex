package handler_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/kazuy/lifnex/app/internal/handler"
	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
)

func TestSearchEvents(t *testing.T) {
	t.Parallel()

	provider := &eventProviderStub{
		events: []eventmodel.Event{
			{
				Title:       "かわさきジャズ",
				Description: "川崎で開催される音楽イベントです。",
				Schedules: []eventmodel.Schedule{
					{
						Date:      time.Date(2026, time.October, 2, 0, 0, 0, 0, time.UTC),
						StartTime: "19:00:00",
						EndTime:   "21:00:00",
						Details:   "開場は18時30分",
					},
				},
				Locations: []eventmodel.Location{
					{Area: "幸区", Address: "川崎市幸区堀川町"},
				},
				URL: "https://example.com/event",
			},
		},
		totalCount: 1,
	}
	searchEvents := handler.NewSearchEvents(eventusecase.NewSearch(provider))
	input := handler.SearchEventsInput{
		From:      "2026-10-01",
		To:        "2026-10-31",
		Keyword:   "音楽",
		Locations: []string{"幸区"},
	}

	_, output, err := searchEvents(t.Context(), nil, input)
	if err != nil {
		t.Fatalf("search events: %v", err)
	}

	wantCondition := eventmodel.SearchCondition{
		From:      time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
		Keyword:   "音楽",
		Locations: []string{"幸区"},
	}
	if !reflect.DeepEqual(provider.condition, wantCondition) {
		t.Errorf("condition = %#v, want %#v", provider.condition, wantCondition)
	}

	wantOutput := handler.SearchEventsOutput{
		Events: []handler.EventOutput{
			{
				Title:       "かわさきジャズ",
				Description: "川崎で開催される音楽イベントです。",
				Schedules: []handler.ScheduleOutput{
					{
						Date:      "2026-10-02",
						StartTime: "19:00:00",
						EndTime:   "21:00:00",
						Details:   "開場は18時30分",
					},
				},
				Locations: []handler.LocationOutput{
					{Area: "幸区", Address: "川崎市幸区堀川町"},
				},
				URL: "https://example.com/event",
			},
		},
		TotalCount: 1,
	}
	if !reflect.DeepEqual(output, wantOutput) {
		t.Errorf("output = %#v, want %#v", output, wantOutput)
	}
}

func TestSearchEventsRejectsInvalidDate(t *testing.T) {
	t.Parallel()

	searchEvents := handler.NewSearchEvents(eventusecase.NewSearch(&eventProviderStub{}))

	_, _, err := searchEvents(t.Context(), nil, handler.SearchEventsInput{
		From: "2026/10/01",
		To:   "2026-10-31",
	})
	if err == nil {
		t.Fatal("search events error = nil, want date parse error")
	}
}

type eventProviderStub struct {
	events     []eventmodel.Event
	totalCount int
	condition  eventmodel.SearchCondition
}

func (p *eventProviderStub) Search(_ context.Context, condition eventmodel.SearchCondition) ([]eventmodel.Event, int, error) {
	p.condition = condition

	return p.events, p.totalCount, nil
}
