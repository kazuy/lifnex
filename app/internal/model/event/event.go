package event

import "time"

// Event represents event information independent of its source.
type Event struct {
	Title       string
	Description string
	Schedules   []Schedule
	Locations   []Location
	URL         string
}

// Schedule represents one occurrence of an event.
type Schedule struct {
	Date      time.Time
	StartTime string
	EndTime   string
	Details   string
}

// Location represents a place where an event is held.
type Location struct {
	Area    string
	Address string
}
