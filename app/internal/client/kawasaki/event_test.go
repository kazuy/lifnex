package kawasaki

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
)

func TestEventClientSearch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}

		query := r.URL.Query()
		for key, want := range map[string]string{
			"format": "JSON",
			"page":   "1",
			"from":   "2026-10-01",
			"to":     "2026-10-31",
			"title":  "音楽",
			"place":  "1,2",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("query %q = %q, want %q", key, got, want)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"total_numbers": 1,
			"event_data": [{
				"title": "かわさきジャズ",
				"content": "川崎で開催される音楽イベントです。",
				"date_list": [{
					"date": "2026-10-02",
					"time_from": "19:00:00",
					"time_to": "21:00:00",
					"time_ext": "開場は18時30分"
				}],
				"place": "幸区",
				"place_adr": "川崎市幸区堀川町",
				"event_location": [{
					"place": "川崎区",
					"venue_address": "川崎市川崎区宮本町"
				}],
				"open_url": "https://example.com/event",
				"rel_list": [{"rel_url": "https://example.com/related"}]
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := NewEventClient(server.Client())
	client.endpoint = server.URL
	condition := eventmodel.SearchCondition{
		From:      time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
		Keyword:   "音楽",
		Locations: []string{"川崎区", "幸区"},
	}

	events, totalCount, err := client.Search(t.Context(), condition)
	if err != nil {
		t.Fatalf("search events: %v", err)
	}

	wantEvents := []eventmodel.Event{
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
				{Area: "川崎区", Address: "川崎市川崎区宮本町"},
			},
			URL: "https://example.com/event",
		},
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Errorf("events = %#v, want %#v", events, wantEvents)
	}
	if totalCount != 1 {
		t.Errorf("total count = %d, want 1", totalCount)
	}
}

func TestEventClientSearchUsesRelatedURLWhenOpenURLIsEmpty(t *testing.T) {
	t.Parallel()

	server := newEventServer(t, http.StatusOK, `{
		"total_numbers": 1,
		"event_data": [{
			"title": "Event",
			"date_list": [],
			"open_url": null,
			"rel_list": [
				{"rel_url": ""},
				{"rel_url": "https://example.com/related"}
			]
		}]
	}`)

	events, _, err := eventClientForServer(server).Search(t.Context(), validSearchCondition())
	if err != nil {
		t.Fatalf("search events: %v", err)
	}

	if events[0].URL != "https://example.com/related" {
		t.Errorf("URL = %q, want %q", events[0].URL, "https://example.com/related")
	}
}

func TestEventClientSearchRejectsUnsupportedLocation(t *testing.T) {
	t.Parallel()

	client := NewEventClient(http.DefaultClient)
	condition := validSearchCondition()
	condition.Locations = []string{"未対応地域"}
	_, _, err := client.Search(t.Context(), condition)
	if err == nil {
		t.Fatal("search events error = nil, want unsupported location error")
	}
}

func TestEventClientSearchReturnsHTTPError(t *testing.T) {
	t.Parallel()

	server := newEventServer(t, http.StatusInternalServerError, `{"message":"Internal server error"}`)

	_, _, err := eventClientForServer(server).Search(t.Context(), validSearchCondition())
	if err == nil {
		t.Fatal("search events error = nil, want HTTP error")
	}
}

func TestEventClientSearchReturnsDecodeError(t *testing.T) {
	t.Parallel()

	server := newEventServer(t, http.StatusOK, `{`)

	_, _, err := eventClientForServer(server).Search(t.Context(), validSearchCondition())
	if err == nil {
		t.Fatal("search events error = nil, want decode error")
	}
}

func TestEventClientSearchReturnsDateParseError(t *testing.T) {
	t.Parallel()

	server := newEventServer(t, http.StatusOK, `{
		"total_numbers": 1,
		"event_data": [{
			"title": "Event",
			"date_list": [{"date": "invalid"}]
		}]
	}`)

	_, _, err := eventClientForServer(server).Search(t.Context(), validSearchCondition())
	if err == nil {
		t.Fatal("search events error = nil, want date parse error")
	}
}

func TestEventClientSearchReturnsRequestError(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("request failed")
	})}

	_, _, err := NewEventClient(httpClient).Search(t.Context(), validSearchCondition())
	if err == nil {
		t.Fatal("search events error = nil, want request error")
	}
}

func newEventServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server
}

func eventClientForServer(server *httptest.Server) *EventClient {
	client := NewEventClient(server.Client())
	client.endpoint = server.URL

	return client
}

func validSearchCondition() eventmodel.SearchCondition {
	date := time.Date(2026, time.October, 10, 0, 0, 0, 0, time.UTC)

	return eventmodel.SearchCondition{From: date, To: date}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

var _ eventusecase.Provider = (*EventClient)(nil)
