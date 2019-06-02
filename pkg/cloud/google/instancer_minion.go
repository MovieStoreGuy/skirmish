package google

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

type instancer struct {
	*base
	recover []*types.Instance
}

func (in *instancer) Do(ctx context.Context, step types.Step, mode string) {
	in.lock.Lock()
	defer in.lock.Unlock()
	in.logger.Info("Gathering instances data", zap.String("mode", mode))
	instances, err := in.loadInstances(ctx, &step)
	if err != nil {
		in.logger.Error("failed to load instances", zap.Error(err))
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, instance := range instances {
		if r.Float32() > step.Sample {
			in.logger.Info("Ignoring instance due to sampling", zap.String("instance", instance.Name))
			continue
		}
		switch mode {
		case types.DryRun:
			in.logger.Info("Deleting instances", zap.String("instance", instance.Name), zap.String("mode", mode), zap.String("zone", instance.Zone), zap.String("region", instance.Region))
		case types.Repairable:
			resp, err := in.svc.Instances.Stop(instance.Project, instance.CompleteZone(), instance.Name).Do()
			if err != nil {
				in.logger.Error("Failed to stop instance", zap.String("instance", instance.Name), zap.Error(err))
				continue
			}
			if resp.Error != nil {
				in.logger.Error("Response error to delete instance", zap.String("instance", instance.Name), zap.Any("response.error", resp.Error))
				continue
			}
			in.logger.Info("Successfully stopped instance", zap.String("instance", instance.Name), zap.String("zone", instance.Zone), zap.String("region", instance.Region))
			in.recover = append(in.recover, instance)
		case types.Destruction:
			resp, err := in.svc.Instances.Delete(instance.Project, instance.CompleteZone(), instance.Name).Do()
			if err != nil {
				in.logger.Error("Failed to delete instance", zap.String("instance", instance.Name), zap.Error(err))
				continue
			}
			if resp.Error != nil {
				in.logger.Error("Response error to delete instance", zap.String("instance", instance.Name), zap.Any("response.error", resp.Error))
				continue
			}
			in.logger.Info("Successfully deleted instance", zap.String("instance", instance.Name), zap.String("zone", instance.Zone), zap.String("region", instance.Region))
		}
	}
}

func (in *instancer) Restore() {
	in.lock.Lock()
	defer in.lock.Unlock()
	for _, instance := range in.recover {
		resp, err := in.svc.Instances.Start(instance.Project, instance.Zone, instance.Name).Do()
		if err != nil {
			in.logger.Error("Failed to start instance", zap.String("instance", instance.Name), zap.Error(err))
			continue
		}
		if resp.Error != nil {
			in.logger.Error("Response error to start instance", zap.String("instance", instance.Name), zap.Any("response.error", resp.Error))
			continue
		}
		in.logger.Info("Successfully started instance", zap.String("instance", instance.Name), zap.String("zone", instance.Zone), zap.String("region", instance.Region))
	}
}
