package orchestra

import (
	"context"
	"fmt"
	"sync"

	"github.com/MovieStoreGuy/skirmish/pkg/minions"
	"github.com/MovieStoreGuy/skirmish/pkg/signal"
	"github.com/MovieStoreGuy/skirmish/pkg/types"

	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"
	htransport "google.golang.org/api/transport/http"
)

type orchestrator struct {
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *zap.Logger
	handler *signal.Handler
	// metadata ensure that only load the data once and use it throughout
	metadata struct {
		once    sync.Once
		regions []string
		zones   []string
	}
	services struct {
		compute *compute.Service
	}
	factory map[string]func(*zap.Logger) minions.Minion
}

// NewRunner returns an orchestrator configured to party
func NewRunner(logger *zap.Logger, ctx context.Context, cancel context.CancelFunc) (Runner, error) {
	orc := &orchestrator{
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
	if err := orc.loadServices(); err != nil {
		return nil, err
	}
	return orc, nil
}

func (o *orchestrator) Execute(plan *types.Plan) error {
	// Gather all resources accessible to orchestrator
	// Execute each step within the plan at the given mode
	handler := signal.NewHandler()
	o.handler = handler
	// In the event something horrid happens, we need to ensure service is restored
	// so if any events have been stored then we need to clean up and report back
	defer handler.Finalise()
	defer handler.Done()
	for _, step := range plan.Steps {
		handler.Finalise()
		handler = signal.NewHandler()
		o.handler = handler
		o.logger.Info("Starting execution", zap.String("name", step.Name), zap.String("description", step.Description))
		for _, op := range step.Operations {
			for _, project := range step.Projects {
				// Should only be slow on first execution
				if err := o.collectMetadata(project); err != nil {
					return err
				}
			}
			// Create executor based off steps.
			gen, exist := o.factory[op]
			if !exist {
				return fmt.Errorf("no operation listed as %s", op)
			}
			min := gen(o.logger)
			go min.Do(o.ctx, step, plan.Mode)
			handler.Register(min.Restore)
		}
		handler.Done()
		o.logger.Info("finished execution", zap.String("name", step.Name), zap.String("description", step.Description))
	}
	return nil
}

func (o *orchestrator) Shutdown() error {
	o.cancel()
	o.handler.Finalise()
	return nil
}

func (o *orchestrator) loadServices() error {
	hc, _, err := htransport.NewClient(o.ctx)
	if err != nil {
		return err
	}
	compute, err := compute.New(hc)
	if err != nil {
		return err
	}
	o.services.compute = compute
	return nil
}

func (o *orchestrator) collectMetadata(project string) error {
	var err error
	o.metadata.once.Do(func() {
		err = o.services.compute.Zones.List(project).Pages(o.ctx, func(list *compute.ZoneList) error {
			for _, item := range list.Items {
				o.metadata.zones = append(o.metadata.zones, item.Name)
			}
			return nil
		})
		if err != nil {
			return
		}
		err = o.services.compute.Regions.List(project).Pages(o.ctx, func(list *compute.RegionList) error {
			for _, item := range list.Items {
				o.metadata.regions = append(o.metadata.regions, item.Name)
			}
			return nil
		})
	})
	return err
}
