package event

import "time"

// SearchCondition describes filters for finding events.
type SearchCondition struct {
	From      time.Time
	To        time.Time
	Keyword   string
	Locations []string
}
