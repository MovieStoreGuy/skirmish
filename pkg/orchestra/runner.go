package orchestra

import (
	"context"
	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

// Runner defines the operation for the orchestration of chaos
type Runner interface {
	// Initialise configures the orchestrator and loads the required providers
	Initialise(ctx context.Context, plan *types.Plan) error

	// Execute will run the game plan and load all the required services
	Execute(plan *types.Plan) error

	// Shutdown is an idempotent operation that will
	// ensure the stared skirmish will cancel straight away
	Shutdown() error
}
