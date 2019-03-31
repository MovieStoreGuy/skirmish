package minions

import (
	"context"
	"sync"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
	"go.uber.org/zap"
)

type instanceDriver struct {
	lock     sync.Mutex
	log      *zap.Logger
	svc      *types.Services
	metadata *types.Metadata
	recover  []types.Instance
}

func NewInstance(log *zap.Logger, svc *types.Services, meta *types.Metadata) Minion {
	return &instanceDriver{
		log:      log,
		svc:      svc,
		metadata: meta,
	}
}

func (gik *instanceDriver) Do(ctx context.Context, step types.Step, mode string) {
	gik.lock.Lock()
	defer gik.lock.Unlock()
	gik.log.Info("Gathering instances data", zap.String("mode", mode))
	instances, err := filterInstances(ctx, gik.svc, gik.metadata, &step)
	if err != nil {
		return
	}
	for _, instance := range instances {
		switch mode {
		case types.DryRun:
			gik.log.Info("Deleting instances", zap.String("instance", instance.Name), zap.String("mode", mode))
		case types.Repairable:
			resp, err := gik.svc.Compute.Instances.Stop(instance.Project, instance.Zone, instance.Name).Do()
			if err != nil {
				gik.log.Error("Failed to stop instance", zap.String("instance", instance.Name), zap.Error(err))
				continue
			}
			if resp.Error != nil {
				gik.log.Error("Response error to delete instance", zap.String("instance", instance.Name), zap.Any("response.error", resp.Error))
				continue
			}
			gik.log.Info("Successfully stopped instance", zap.String("instance", instance.Name), zap.String("zone", instance.Zone))
			gik.recover = append(gik.recover, instance)
		case types.Destruction:
			resp, err := gik.svc.Compute.Instances.Delete(instance.Project, instance.Zone, instance.Name).Do()
			if err != nil {
				gik.log.Error("Failed to delete instance", zap.String("instance", instance.Name), zap.Error(err))
				continue
			}
			if resp.Error != nil {
				gik.log.Error("Response error to delete instance", zap.String("instance", instance.Name), zap.Any("response.error", resp.Error))
				continue
			}
			gik.log.Info("Successfully deleted instance", zap.String("instance", instance.Name), zap.String("zone", instance.Zone))
		}
	}
}

func (gik *instanceDriver) Restore() {
	for _, instance := range gik.recover {
		resp, err := gik.svc.Compute.Instances.Start(instance.Project, instance.Zone, instance.Name).Do()
		if err != nil {
			gik.log.Error("Failed to start instance", zap.String("instance", instance.Name), zap.Error(err))
			continue
		}
		if resp.Error != nil {
			gik.log.Error("Response error to start instance", zap.String("instance", instance.Name), zap.Any("response.error", resp.Error))
			continue
		}
		gik.log.Info("Successfully started instance", zap.String("instance", instance.Name), zap.String("zone", instance.Zone))
	}
}
