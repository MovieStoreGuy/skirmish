package google

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"

	"github.com/MovieStoreGuy/skirmish/pkg/cloud"
	"github.com/MovieStoreGuy/skirmish/pkg/minions"
	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

type provider struct {
	svc      *compute.Service
	metadata types.Metadata
	logger   *zap.Logger
}

func NewProvider(logger *zap.Logger) cloud.Provider {
	return &provider{
		logger: logger,
	}
}

func (p *provider) Initialise(ctx context.Context, conf *types.Config) error {
	if conf == nil {
		return errors.New("google config was nil")
	}
	client, err := compute.NewService(ctx)
	if err != nil {
		return err
	}
	p.svc = client
	var once sync.Once
	for _, project := range conf.Google.Projects {
		// Optimisation in order to preload all regions and zones within Google
		// since the api doesn't have an auto-generated list.
		once.Do(func() {
			err = p.svc.Zones.List(project).Pages(ctx, func(list *compute.ZoneList) error {
				for _, item := range list.Items {
					p.metadata.Zones = append(p.metadata.Zones, item.Name)
				}
				return nil
			})
			if err != nil {
				return
			}
			err = p.svc.Regions.List(project).Pages(ctx, func(list *compute.RegionList) error {
				for _, item := range list.Items {
					p.metadata.Regions = append(p.metadata.Regions, item.Name)
				}
				return nil
			})
			if err != nil {
				return
			}
		})
	}
	return err
}

func (p *provider) LoadMinionsFactory() (cloud.Factory, error) {
	if p.svc == nil {
		return nil, errors.New("provider not initialised")
	}
	factory := make(cloud.Factory, 0)
	err := factory.RegisterMinion("instance", func() minions.Minion {
		return &instancer{
			base: &base{
				logger:   p.logger,
				metadata: &p.metadata,
				svc:      p.svc,
			},
			recover: make([]*types.Instance, 0),
		}
	})
	if err != nil {
		return nil, err
	}
	err = factory.RegisterMinion("ingress", func() minions.Minion {
		return &networker{
			base: &base{
				logger:   p.logger,
				metadata: &p.metadata,
				svc:      p.svc,
			},
			firewalls: make(map[string]*types.Firewall, 0),
			flow:      "INGRESS",
		}
	})
	if err != nil {
		return nil, err
	}
	err = factory.RegisterMinion("egress", func() minions.Minion {
		return &networker{
			base: &base{
				logger:   p.logger,
				metadata: &p.metadata,
				svc:      p.svc,
			},
			firewalls: make(map[string]*types.Firewall, 0),
			flow:      "EGRESS",
		}
	})
	return factory, err
}

func (p *provider) String() string {
	return "Google Cloud Provider"
}
