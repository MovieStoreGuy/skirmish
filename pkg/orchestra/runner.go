package orchestra

import "github.com/MovieStoreGuy/skirmish/pkg/types"

type Runner interface {
	// Execute will run the game plan
	Execute(plan *types.Plan) error

	Shutdown() error
}
