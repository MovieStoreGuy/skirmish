package minions

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
	"go.uber.org/zap"
)

type ingressDriver struct {
	lock     sync.Mutex
	log      *zap.Logger
	svc      *types.Services
	metadata *types.Metadata
	recover  []types.Instance
}

func (id *ingressDriver) Do(ctx context.Context, step types.Step, mode string) {
	id.lock.Lock()
	defer id.lock.Unlock()
	instances, err := filterInstances(ctx, id.svc, id.metadata, &step)
	if err != nil {
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, instance := range instances {
		if r.Float32() > step.Sample {
			id.log.Info("Ignoring instance due to sampling", zap.String("instance", instance.Name))
			continue
		}
		switch mode {
		case types.Repairable, types.Destruction:

		case types.DryRun:

		}
	}
}
