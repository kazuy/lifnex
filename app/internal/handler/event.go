package handler

import (
	"context"
	"fmt"
	"time"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const eventDateLayout = "2006-01-02"

// SearchEventsInput contains the event search conditions accepted by the MCP tool.
type SearchEventsInput struct {
	From      string   `json:"from" jsonschema:"Search start date in YYYY-MM-DD format. Required."`
	To        string   `json:"to" jsonschema:"Search end date in YYYY-MM-DD format. Required. Use the same date as from to search a single day."`
	Keyword   string   `json:"keyword,omitempty" jsonschema:"Optional keyword contained in the event title."`
	Locations []string `json:"locations,omitempty" jsonschema:"Optional locations. Supported values are 川崎区, 幸区, 中原区, 高津区, 宮前区, 多摩区, 麻生区, 横浜市, 東京都, その他, and オンライン."`
}

// SearchEventsOutput contains matching events or guidance to refine the search.
type SearchEventsOutput struct {
	Events     []EventOutput `json:"events"`
	TotalCount int           `json:"totalCount"`
	Message    string        `json:"message,omitempty"`
}

// EventOutput is the MCP representation of an event.
type EventOutput struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Schedules   []ScheduleOutput `json:"schedules"`
	Locations   []LocationOutput `json:"locations"`
	URL         string           `json:"url,omitempty"`
}

// ScheduleOutput is the MCP representation of one event occurrence.
type ScheduleOutput struct {
	Date      string `json:"date"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
	Details   string `json:"details,omitempty"`
}

// LocationOutput is the MCP representation of an event location.
type LocationOutput struct {
	Area    string `json:"area,omitempty"`
	Address string `json:"address,omitempty"`
}

// NewSearchEvents creates an MCP handler for the event search use case.
func NewSearchEvents(search *eventusecase.Search) mcp.ToolHandlerFor[SearchEventsInput, SearchEventsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchEventsInput) (*mcp.CallToolResult, SearchEventsOutput, error) {
		condition, err := searchCondition(input)
		if err != nil {
			return nil, SearchEventsOutput{}, err
		}

		result, err := search.Execute(ctx, condition)
		if err != nil {
			return nil, SearchEventsOutput{}, err
		}

		return nil, searchEventsOutput(result), nil
	}
}

func searchCondition(input SearchEventsInput) (eventmodel.SearchCondition, error) {
	from, err := time.Parse(eventDateLayout, input.From)
	if err != nil {
		return eventmodel.SearchCondition{}, fmt.Errorf("parse event search from %q: %w", input.From, err)
	}

	to, err := time.Parse(eventDateLayout, input.To)
	if err != nil {
		return eventmodel.SearchCondition{}, fmt.Errorf("parse event search to %q: %w", input.To, err)
	}

	return eventmodel.SearchCondition{
		From:      from,
		To:        to,
		Keyword:   input.Keyword,
		Locations: input.Locations,
	}, nil
}

func searchEventsOutput(result eventusecase.SearchResult) SearchEventsOutput {
	events := make([]EventOutput, 0, len(result.Events))
	for _, event := range result.Events {
		schedules := make([]ScheduleOutput, 0, len(event.Schedules))
		for _, schedule := range event.Schedules {
			schedules = append(schedules, ScheduleOutput{
				Date:      schedule.Date.Format(eventDateLayout),
				StartTime: schedule.StartTime,
				EndTime:   schedule.EndTime,
				Details:   schedule.Details,
			})
		}

		locations := make([]LocationOutput, 0, len(event.Locations))
		for _, location := range event.Locations {
			locations = append(locations, LocationOutput{
				Area:    location.Area,
				Address: location.Address,
			})
		}

		events = append(events, EventOutput{
			Title:       event.Title,
			Description: event.Description,
			Schedules:   schedules,
			Locations:   locations,
			URL:         event.URL,
		})
	}

	return SearchEventsOutput{
		Events:     events,
		TotalCount: result.TotalCount,
		Message:    result.Message,
	}
}
