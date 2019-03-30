package orchestra

import (
	"context"

	"go.uber.org/zap"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

type orchestrator struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger
}

func NewRunner(logger *zap.Logger) Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &orchestrator{
		ctx,
		cancel,
		logger,
	}
}

func (o *orchestrator) Execute(plan *types.Plan) error {
	// Gather all resources accessible to orchestrator
	// Execute each step within the plan at the given mode
	return nil
}

func (o *orchestrator) Shutdown() error {
	o.cancel()
	return nil
}
