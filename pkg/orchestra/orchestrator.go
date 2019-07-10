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
	providers map[string]cloud.Provider
}

// NewRunner returns an orchestrator that has all the providers registered
func NewRunner(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger) (Runner, error) {
	o := &orchestrator{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		handler:   signal.NewHandler(),
		providers: make(map[string]cloud.Provider, 0),
	}
	return o, nil
}

func (o *orchestrator) Initialise(ctx context.Context, plan *types.Plan) error {
	choices := map[string]func(*zap.Logger) cloud.Provider{
		"google": google.NewProvider,
	}
	for _, c := range plan.Providers {
		if op, exist := choices[c.Name]; exist {
			if _, init := o.providers[c.Name]; init {
				// Provider has already been successfully created
				// No need to create a new provider again
				continue
			}
			provider := op(o.logger)
			// TODO(Sean Marciniak): Update plan definition to include config object
			if err := provider.Initialise(ctx, &types.Config{}); err != nil {
				return err
			}
			o.providers[c.Name] = provider
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
			provider, exist := o.providers[step.Provider]
			if !exist {
				return fmt.Errorf("no provider initialised called %s", step.Provider)
			}
			factory, err := provider.LoadMinionsFactory()
			if err != nil {
				return err
			}
			min, err := factory.CreateMinion(op)
			if err != nil{
				return err
			}
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
