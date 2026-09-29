package event

import (
	"context"

	eventmodel "github.com/kazuy/lifnex/app/internal/model/event"
)

// Provider retrieves events from an external source.
type Provider interface {
	Search(context.Context, eventmodel.SearchCondition) ([]eventmodel.Event, int, error)
}
