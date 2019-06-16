package orchestra

import (
	"context"
	"fmt"
	"time"

	"github.com/MovieStoreGuy/skirmish/pkg/cloud"
	"github.com/MovieStoreGuy/skirmish/pkg/cloud/google"
	"github.com/MovieStoreGuy/skirmish/pkg/signal"
	"github.com/MovieStoreGuy/skirmish/pkg/types"

	"go.uber.org/zap"
)

type orchestrator struct {
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	handler   *signal.Handler
	providers map[string]func(*zap.Logger) cloud.Provider
}

// NewRunner returns an orchestrator that has all the providers registered
func NewRunner(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger) (Runner, error) {
	o := &orchestrator{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		handler:   signal.NewHandler(),
		providers: make(map[string]func(*zap.Logger) cloud.Provider, 0),
	}
	return o, nil
}

func (o *orchestrator) Initialise(plan *types.Plan) error {
	choices := map[string]func(*zap.Logger) cloud.Provider{
		"google": google.NewProvider,
	}
	for _, c := range plan.Providers {
		if op, exist := choices[c.Name]; exist {
			o.providers[c.Name] = op
		} else {
			names := make([]string, 0)
			for name := range choices {
				names = append(names, name)
			}
			return fmt.Errorf("unknown provider %s, list is %v", c.Name, names)
		}
	}
	return nil
}

func (o *orchestrator) Execute(plan *types.Plan) error {
	handler := signal.NewHandler()
	// In the event something horrid happens, we need to ensure service is restored
	// so if any events have been stored then we need to clean up and report back
	defer handler.Finalise()
	defer handler.Done()
	for _, step := range plan.Steps {
		handler.Done()
		handler.Finalise()
		handler = signal.NewHandler()
		o.handler = handler
		o.logger.Info("Starting execution", zap.String("name", step.Name), zap.String("description", step.Description))
		for _, op := range step.Operations {
			gen, exist := o.factory[op]
			if !exist {
				return fmt.Errorf("no operation listed as %s", op)
			}
			min := gen(o.logger, o.services, &o.metadata)
			go min.Do(o.ctx, step, plan.Mode)
			handler.Register(min.Restore)
		}
		o.logger.Info("finished starting all operations", zap.String("name", step.Name), zap.String("description", step.Description))
		if plan.Mode != types.DryRun {
			time.Sleep(step.Wait)
		}
	}
	return nil
}

func (o *orchestrator) Shutdown() error {
	if o.cancel != nil {
		o.cancel()
	}
	o.handler.Finalise()
	return nil
}

func (o *orchestrator) String() string {
	loaded := make([]string, 0, len(o.providers))
	for name := range o.providers {
		loaded = append(loaded, name)
	}
	return fmt.Sprintf("loaded providers:%v", loaded)
}
