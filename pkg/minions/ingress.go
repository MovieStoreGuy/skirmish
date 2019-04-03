package minions

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
	"go.uber.org/zap"
)

type networkDriver struct {
	flow string

	lock     sync.Mutex
	log      *zap.Logger
	svc      *types.Services
	metadata *types.Metadata
	recover  []types.Instance
}

func NewNetworkDriver(flow string) func(*zap.Logger, *types.Services, *types.Metadata) Minion {
	return func(log *zap.Logger, svc *types.Services, meta *types.Metadata) Minion {
		return &networkDriver{
			flow:     flow,
			log:      log,
			svc:      svc,
			metadata: meta,
		}
	}
}

func (nd *networkDriver) Do(ctx context.Context, step types.Step, mode string) {
	nd.lock.Lock()
	defer nd.lock.Unlock()
	instances, err := filterInstances(ctx, nd.svc, nd.metadata, &step)
	if err != nil {
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, instance := range instances {
		if r.Float32() > step.Sample {
			nd.log.Info("Ignoring instance due to sampling", zap.String("instance", instance.Name))
			continue
		}
		switch mode {
		case types.Repairable, types.Destruction:

		case types.DryRun:

		}
	}
}

func (nd *networkDriver) Restore() {
	nd.lock.Lock()
	defer nd.lock.Unlock()
}
