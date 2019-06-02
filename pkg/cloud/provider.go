package cloud

import (
	"context"

	"github.com/MovieStoreGuy/skirmish/pkg/minions"
	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

// Provider defines a backend for a specific cloud provider
type Provider interface {

	// Initialise configures the provider ready to interact with the provider
	Initialise(ctx context.Context, conf *types.Config) error

	// LoadMinionsFactory returns a minion generator of the given provider
	LoadMinionsFactory() (map[string]func() minions.Minion, error)
}
