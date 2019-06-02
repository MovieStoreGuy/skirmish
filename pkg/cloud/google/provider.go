package google

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"

	"github.com/MovieStoreGuy/skirmish/pkg/minions"
	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

type provider struct {
	svc      *compute.Service
	metadata types.Metadata
	logger   *zap.Logger
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

func (p *provider) LoadMinionsFactory() (map[string]func() minions.Minion, error) {
	factory := map[string]func() minions.Minion{}

	return factory, nil
}