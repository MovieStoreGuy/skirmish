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
	htransport "google.golang.org/api/transport/http"
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
	orc := &orchestrator{
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		handler:  signal.NewHandler(),
		services: &types.Services{},
	}
	if err := orc.loadServices(); err != nil {
		return nil, err
	}
	return orc, nil
}

func (o *orchestrator) Execute(plan *types.Plan) error {
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
		handler.Finalise()
		handler = signal.NewHandler()
		o.handler = handler
		o.logger.Info("Starting execution", zap.String("name", step.Name), zap.String("description", step.Description))
		for _, op := range step.Operations {
			// Create executor based off steps.
			gen, exist := o.factory[op]
			if !exist {
				return fmt.Errorf("no operation listed as %s", op)
			}
			min := gen(o.logger, o.services, &o.metadata)
			go min.Do(o.ctx, step, plan.Mode)
			handler.Register(min.Restore)
		}
		handler.Done()
		o.logger.Info("finished execution", zap.String("name", step.Name), zap.String("description", step.Description))
		time.Sleep(step.Wait)
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
	o.services.Compute = compute
	return nil
}

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
