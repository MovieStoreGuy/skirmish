package minions

import (
	"context"
	"regexp"
	"strings"

	"google.golang.org/api/compute/v1"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

// filterInstances will return a list of instances that aren't part of the exclusion list.
func filterInstances(ctx context.Context, svc *types.Services, metadata *types.Metadata, step *types.Step) ([]types.Instance, error) {
	instances := []types.Instance{}
	for _, project := range step.Projects {
		for _, zone := range metadata.Zones {
			err := svc.Compute.Instances.List(project, zone).Pages(ctx, func(list *compute.InstanceList) error {
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
					for _, exclude := range step.Exclude.Zones {
						if strings.HasPrefix(item.Zone, exclude) {
							excluded = true
						}
					}
					for entry, key := range step.Exclude.Labels {
						if value, ok := item.Labels[entry]; ok && value == key {
							excluded = true
						}
					}
					if !excluded {
						instances = append(instances, types.Instance{
							Id:      item.Id,
							Name:    item.Name,
							Zone:    item.Zone,
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
