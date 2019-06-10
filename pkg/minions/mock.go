package minions

import (
	"context"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

// Mock should only be used for testing purposes
type Mock struct {
}

func (m *Mock) Do(context.Context, types.Step, string) {
	// Do nothing
}

func (m *Mock) Restore() {
	// Do nothing
}
