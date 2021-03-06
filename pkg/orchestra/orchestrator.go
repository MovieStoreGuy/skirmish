package orchestra

import (
	"context"
	"fmt"
	"time"

	"github.com/MovieStoreGuy/skirmish/pkg/minions"
	"github.com/MovieStoreGuy/skirmish/pkg/signal"
	"github.com/MovieStoreGuy/skirmish/pkg/types"

	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"
)

type orchestrator struct {
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *zap.Logger
	handler  *signal.Handler
	metadata types.Metadata
	services *types.Services
	factory  map[string]func(*zap.Logger, *types.Services, *types.Metadata) minions.Minion
}

// NewRunner returns an orchestrator configured to party
func NewRunner(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger) (Runner, error) {
	o := &orchestrator{
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		handler:  signal.NewHandler(),
		services: &types.Services{},
		factory: map[string]func(*zap.Logger, *types.Services, *types.Metadata) minions.Minion{
			"instance": minions.NewInstance,
			"ingress":  minions.NewNetworkDriver("INGRESS"),
			"egress":   minions.NewNetworkDriver("EGRESS"),
		},
	}
	return o, nil
}

func (o *orchestrator) Execute(plan *types.Plan) error {
	if err := o.loadServices(); err != nil {
		return err
	}
	handler := signal.NewHandler()
	// In the event something horrid happens, we need to ensure service is restored
	// so if any events have been stored then we need to clean up and report back
	defer handler.Finalise()
	defer handler.Done()
	for _, project := range plan.Projects {
		if err := o.collectMetadata(project); err != nil {
			return err
		}
	}
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

func (o *orchestrator) loadServices() error {
	c, err := compute.NewService(o.ctx)
	if err != nil {
		return err
	}
	o.services.Compute = c
	return nil
}

func (o *orchestrator) String() string {
	loaded := make([]string, 0, len(o.factory))
	for name := range o.factory {
		loaded = append(loaded, name)
	}
	return fmt.Sprintf("loaded minions:%v", loaded)
}

// collectMetadata will return all the zones, region only once to save
func (o *orchestrator) collectMetadata(project string) error {
	var err error
	o.metadata.Once.Do(func() {
		err = o.services.Compute.Zones.List(project).Pages(o.ctx, func(list *compute.ZoneList) error {
			for _, item := range list.Items {
				o.metadata.Zones = append(o.metadata.Zones, item.Name)
			}
			return nil
		})
		if err != nil {
			return
		}
		err = o.services.Compute.Regions.List(project).Pages(o.ctx, func(list *compute.RegionList) error {
			for _, item := range list.Items {
				o.metadata.Regions = append(o.metadata.Regions, item.Name)
			}
			return nil
		})
	})
	return err
}
