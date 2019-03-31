package orchestra

import "github.com/MovieStoreGuy/skirmish/pkg/types"

// Runner defines the operation for the orchestration of chaos
type Runner interface {
	// Execute will run the game plan
	Execute(plan *types.Plan) error

	// Shutdown is an idempotent operation that will
	// ensure the stared skirmish will cancel straight away
	Shutdown() error
}
