package cloud

import (
	"go/types"

	"github.com/MovieStoreGuy/skirmish/pkg/minions"
)

// Provider defines a backend for a specific cloud provider
type Provider interface {

	// Initialise configures the provider ready to interact with the provider
	Initialise(config *types.Config) error

	// LoadMinionsFactory returns a minion generator of the given provider
	LoadMinionsFactory() (map[string]func() minions.Minion, error)
}
