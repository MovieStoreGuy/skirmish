package google

import (
	"context"
	"errors"
	"path"
	"regexp"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

type base struct {
	lock     sync.Mutex
	logger   *zap.Logger
	metadata *types.Metadata
	svc      *compute.Service
}

func (b *base) loadInstances(ctx context.Context, step *types.Step) ([]*types.Instance, error) {
	if step == nil {
		return nil, errors.New("step was nil")
	}
	instances := make([]*types.Instance, 0)
	for _, project := range step.Projects {
		for _, zone := range b.metadata.Zones {
			err := b.svc.Instances.List(project, zone).Pages(ctx, func(list *compute.InstanceList) error {
				for _, item := range list.Items {
					excluded := false
					for _, wildcard := range step.Exclude.Wildcards {
						r, err := regexp.Compile(wildcard)
						if err != nil {
							continue
						}
						if r.MatchString(item.Name) {
							excluded = true
						}
					}
					combined := strings.Split(path.Base(item.Zone), "-")
					if len(combined) != 3 {
						return errors.New("incorrect amount of values to use")
					}
					region, zone := combined[0]+"-"+combined[1], combined[2]
					for _, exclude := range step.Exclude.Zones {
						if strings.HasPrefix(zone, exclude) {
							excluded = true
						}
					}
					for _, exclude := range step.Exclude.Regions {
						if strings.HasPrefix(region, exclude) {
							excluded = true
						}
					}
					for entry, key := range step.Exclude.Labels {
						if value, ok := item.Labels[entry]; ok && value == key {
							excluded = true
						}
					}
					if !excluded {
						instances = append(instances, &types.Instance{
							Id:      item.Id,
							Name:    item.Name,
							Zone:    zone,
							Region:  region,
							Project: project,
							Labels:  item.Labels,
						})
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return instances, nil
}
