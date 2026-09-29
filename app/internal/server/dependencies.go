package server

import (
	"net/http"
	"time"

	"github.com/kazuy/lifnex/app/internal/client/kawasaki"
	eventusecase "github.com/kazuy/lifnex/app/internal/usecase/event"
)

const eventRequestTimeout = 10 * time.Second

// Dependencies provides the application services used by the MCP server.
type Dependencies struct {
	EventSearch *eventusecase.Search
}

// NewDependencies creates the production dependencies used by the MCP server.
func NewDependencies() Dependencies {
	eventClient := kawasaki.NewEventClient(&http.Client{Timeout: eventRequestTimeout})

	return Dependencies{
		EventSearch: eventusecase.NewSearch(eventClient),
	}
}
