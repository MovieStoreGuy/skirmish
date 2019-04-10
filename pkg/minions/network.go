package minions

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/MovieStoreGuy/skirmish/pkg/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"
)

type networkDriver struct {
	flow string

	lock     sync.Mutex
	log      *zap.Logger
	svc      *types.Services
	metadata *types.Metadata

	firewalls map[string]*types.Firewall
}

// NewNetworkDriver returns a function that will ensure that the correct INGRESS or EGRESS type is used.
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
	id, err := uuid.NewRandom()
	if err != nil {
		nd.log.Error("Unable to generate UUID", zap.Error(err))
		return
	}
	// Applying labels to affected instances to not block the entire network
	for _, instance := range instances {
		if r.Float32() > step.Sample {
			nd.log.Info("Ignoring instance due to sampling", zap.String("instance", instance.Name))
			continue
		}
		switch mode {
		case types.Repairable, types.Destruction:
			req := &compute.InstancesSetLabelsRequest{
				Labels: instance.Labels,
			}
			req.Labels[id.String()] = "wargames"
			resp, err := nd.svc.Compute.Instances.SetLabels(instance.Project, instance.Zone, instance.Name, req).Do()
			if err != nil {
				nd.log.Error("Unable to apply label changes", zap.Error(err), zap.String("instance", instance.Name))
				continue
			}
			if resp.Error != nil {
				nd.log.Error("Response returned a failure", zap.Any("error", resp.Error), zap.String("instance", instance.Name))
				continue
			}
			firewall, exist := nd.firewalls[instance.Project]
			if !exist {
				f := &types.Firewall{}
				nd.firewalls[instance.Project], firewall = f, f
			}
			firewall.Labels[id.String()] = "wargames"
			firewall.Instances = append(firewall.Instances, instance)
			fallthrough
		case types.DryRun:
			nd.log.Info("Applying network rules against", zap.String("instance", instance.Name), zap.String("flow", nd.flow))
		}
	}
	gen := nameAppendor()
	for _, conf := range step.Settings.Network {
		f := nd.firewalls[conf.Project]
		switch mode {
		case types.Repairable, types.Destruction:
			fw := buildFirewall(conf.Deny, gen("wargames", nd.flow), conf.Network, nd.flow, id.String())
			resp, err := nd.svc.Compute.Firewalls.Insert(conf.Project, fw).Do()
			if err != nil {
				nd.log.Error("Unable to create firewall", zap.Error(err), zap.String("project", conf.Project))
				continue
			}
			if resp.Error != nil {
				nd.log.Error("Response returned a failure", zap.Any("error", resp.Error))
				continue
			}
			f.Name = fw.Name
			fallthrough
		case types.DryRun:
			nd.log.Info("Applied firewall changes",
				zap.String("name", f.Name),
				zap.String("label", id.String()),
				zap.String("network", conf.Network),
				zap.String("project", conf.Project))
		}
	}
}

func (nd *networkDriver) Restore() {
	nd.lock.Lock()
	defer nd.lock.Unlock()
	for project, firewall := range nd.firewalls {
		for _, instance := range firewall.Instances {
			resp, err := nd.svc.Compute.Instances.SetLabels(instance.Project, instance.Zone, instance.Name, &compute.InstancesSetLabelsRequest{
				Labels: instance.Labels,
			}).Do()
			if err != nil {
				nd.log.Error("Failed to reset labels", zap.Error(err), zap.String("instance", instance.Name), zap.String("project", instance.Project))
				continue
			}
			if resp.Error != nil {
				nd.log.Error("Response failed to reset labels", zap.Any("resp.error", resp.Error), zap.String("instance", instance.Name), zap.String("project", instance.Project))
			}
		}
		resp, err := nd.svc.Compute.Firewalls.Delete(project, firewall.Name).Do()
		if err != nil {
			nd.log.Error("Failed to remove firewall", zap.Error(err), zap.String("project", project), zap.String("firewall", firewall.Name))
			continue
		}
		if resp.Error != nil {
			nd.log.Error("Response failed to delete firewall", zap.Any("resp.error", resp.Error), zap.String("project", project), zap.String("firewall", firewall.Name))
			continue
		}
		nd.log.Info("Removed firewall", zap.String("project", project), zap.String("firewall", firewall.Name))
	}
}
