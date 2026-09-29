package kawasaki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
)

const (
	eventsEndpoint = "https://eventapp.city.kawasaki.jp/data/api/v1/events"
	dateLayout     = "2006-01-02"
)

var locationIDByName = map[string]int{
	"川崎区":   1,
	"幸区":    2,
	"中原区":   3,
	"高津区":   4,
	"宮前区":   5,
	"多摩区":   6,
	"麻生区":   7,
	"横浜市":   8,
	"東京都":   9,
	"その他":   10,
	"オンライン": 11,
}

// EventClient retrieves event information from the Kawasaki City event API.
type EventClient struct {
	httpClient *http.Client
	endpoint   string
}

// NewEventClient creates a Kawasaki City event API client.
func NewEventClient(httpClient *http.Client) *EventClient {
	return &EventClient{
		httpClient: httpClient,
		endpoint:   eventsEndpoint,
	}
}

// Search retrieves events matching condition and converts them to the common event model.
func (c *EventClient) Search(ctx context.Context, condition eventmodel.SearchCondition) ([]eventmodel.Event, int, error) {
	requestURL, err := c.buildRequestURL(condition)
	if err != nil {
		return nil, 0, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create Kawasaki event request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("request Kawasaki events: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, 0, fmt.Errorf("request Kawasaki events: unexpected HTTP status %s", response.Status)
	}

	var payload eventsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, 0, fmt.Errorf("decode Kawasaki events: %w", err)
	}

	events := make([]eventmodel.Event, 0, len(payload.Events))
	for _, source := range payload.Events {
		event, err := convertEvent(source)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	return events, payload.TotalCount, nil
}

func (c *EventClient) buildRequestURL(condition eventmodel.SearchCondition) (string, error) {
	endpoint, err := url.Parse(c.endpoint)
	if err != nil {
		return "", fmt.Errorf("parse Kawasaki event endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("format", "JSON")
	query.Set("page", "1")
	query.Set("from", condition.From.Format(dateLayout))
	query.Set("to", condition.To.Format(dateLayout))
	if condition.Keyword != "" {
		query.Set("title", condition.Keyword)
	}

	locationIDs, err := convertLocationsToIDs(condition.Locations)
	if err != nil {
		return "", err
	}
	if len(locationIDs) > 0 {
		query.Set("place", strings.Join(locationIDs, ","))
	}

	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}

func convertLocationsToIDs(locations []string) ([]string, error) {
	ids := make([]string, 0, len(locations))
	for _, location := range locations {
		id, ok := locationIDByName[location]
		if !ok {
			return nil, fmt.Errorf("convert Kawasaki event location %q: unsupported location", location)
		}
		ids = append(ids, strconv.Itoa(id))
	}

	return ids, nil
}

func convertEvent(source eventResponse) (eventmodel.Event, error) {
	schedules := make([]eventmodel.Schedule, 0, len(source.Dates))
	for _, sourceDate := range source.Dates {
		date, err := time.Parse(dateLayout, sourceDate.Date)
		if err != nil {
			return eventmodel.Event{}, fmt.Errorf("parse Kawasaki event date %q: %w", sourceDate.Date, err)
		}
		schedules = append(schedules, eventmodel.Schedule{
			Date:      date,
			StartTime: sourceDate.StartTime,
			EndTime:   sourceDate.EndTime,
			Details:   sourceDate.Details,
		})
	}

	locations := make([]eventmodel.Location, 0, len(source.Locations)+1)
	if source.Place != "" || source.Address != "" {
		locations = append(locations, eventmodel.Location{
			Area:    source.Place,
			Address: source.Address,
		})
	}
	for _, sourceLocation := range source.Locations {
		locations = append(locations, eventmodel.Location{
			Area:    sourceLocation.Area,
			Address: sourceLocation.Address,
		})
	}

	return eventmodel.Event{
		Title:       source.Title,
		Description: source.Content,
		Schedules:   schedules,
		Locations:   locations,
		URL:         eventURL(source),
	}, nil
}

func eventURL(source eventResponse) string {
	if source.OpenURL != "" {
		return source.OpenURL
	}
	for _, related := range source.RelatedURLs {
		if related.URL != "" {
			return related.URL
		}
	}

	return ""
}
