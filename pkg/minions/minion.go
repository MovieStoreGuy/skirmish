package minions

import (
	"context"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

// Minion defines a task that targets one part of the cloud
// Each minion needs to keep an internal lock to ensure the restore doesn't happen
// at the same time as the Do
type Minion interface {
	// Do will execute the minions job against the given step at the correct mode
	Do(ctx context.Context, step types.Step, mode string)

	// Restore ensures all the resources are put back in place
	// it should only be able to execute if the do function has finished
	Restore()
}
