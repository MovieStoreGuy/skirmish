package google

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/compute/v1"

	"github.com/MovieStoreGuy/skirmish/pkg/types"
)

type networker struct {
	*base

	flow      string
	firewalls map[string]*types.Firewall
}

func nameAppendor() func(...string) string {
	count := 0
	return func(prefix ...string) string {
		s := fmt.Sprintf("%s-%d", strings.Join(prefix, "-"), count)
		count++
		return s
	}
}

func buildFirewall(values []types.Deny, name, network, direction, label string) *compute.Firewall {
	firewall := &compute.Firewall{
		Direction:  direction,
		Name:       name,
		Network:    network,
		Priority:   1,
		TargetTags: []string{fmt.Sprintf("%s:wargames", label)},
		SourceTags: []string{fmt.Sprintf("%s:wargames", label)},
	}
	for _, deny := range values {
		firewall.Denied = append(firewall.Denied, &compute.FirewallDenied{
			IPProtocol: deny.Protocol,
			Ports:      deny.Ports,
		})
	}
	return firewall
}

func (net *networker) Do(ctx context.Context, step types.Step, mode string) {
	net.lock.Lock()
	defer net.lock.Unlock()
	net.logger.Info("Gathering instances data", zap.String("mode", mode))
	instances, err := net.loadInstances(ctx, &step)
	if err != nil {
		net.logger.Error("failed to load instances", zap.Error(err))
		return
	}
	id, err := uuid.NewRandom()
	if err != nil {
		net.logger.Error("Unable to generate UUID", zap.Error(err))
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, instance := range instances {
		if r.Float32() > step.Sample {
			net.logger.Info("Ignoring instance due to sampling", zap.String("instance", instance.Name))
			continue
		}
		switch mode {
		case types.Repairable, types.Destruction:
			req := &compute.InstancesSetLabelsRequest{
				Labels: instance.Labels,
			}
			req.Labels[id.String()] = "wargames"
			resp, err := net.svc.Instances.SetLabels(instance.Project, instance.CompleteZone(), instance.Name, req).Do()
			if err != nil {
				net.logger.Error("Unable to apply label changes", zap.Error(err), zap.String("instance", instance.Name))
				continue
			}
			if resp.Error != nil {
				net.logger.Error("Response returned a failure", zap.Any("error", resp.Error), zap.String("instance", instance.Name))
				continue
			}
			firewall, exist := net.firewalls[instance.Project]
			if !exist {
				f := &types.Firewall{}
				net.firewalls[instance.Project], firewall = f, f
			}
			firewall.Labels[id.String()] = "wargames"
			firewall.Instances = append(firewall.Instances, instance)
			fallthrough
		case types.DryRun:
			net.logger.Info("Applying network rules against", zap.String("instance", instance.Name), zap.String("flow", net.flow))
		}
	}
	gen := nameAppendor()
	for _, conf := range step.Settings.Network {
		f := net.firewalls[conf.Project]
		switch mode {
		case types.Repairable, types.Destruction:
			fw := buildFirewall(conf.Deny, gen("wargames", net.flow), conf.Network, net.flow, id.String())
			resp, err := net.svc.Firewalls.Insert(conf.Project, fw).Do()
			if err != nil {
				net.logger.Error("Unable to create firewall", zap.Error(err), zap.String("project", conf.Project))
				continue
			}
			if resp.Error != nil {
				net.logger.Error("Response returned a failure", zap.Any("error", resp.Error))
				continue
			}
			f.Name = fw.Name
			fallthrough
		case types.DryRun:
			net.logger.Info("Applied firewall changes",
				zap.String("name", f.Name),
				zap.String("label", id.String()),
				zap.String("network", conf.Network),
				zap.String("project", conf.Project))
		}
	}
}

func (net *networker) Restore() {
	net.lock.Lock()
	defer net.lock.Unlock()
	for project, firewall := range net.firewalls {
		for _, instance := range firewall.Instances {
			resp, err := net.svc.Instances.SetLabels(instance.Project, instance.Zone, instance.Name, &compute.InstancesSetLabelsRequest{
				Labels: instance.Labels,
			}).Do()
			if err != nil {
				net.logger.Error("Failed to reset labels", zap.Error(err), zap.String("instance", instance.Name), zap.String("project", instance.Project))
				continue
			}
			if resp.Error != nil {
				net.logger.Error("Response failed to reset labels", zap.Any("resp.error", resp.Error), zap.String("instance", instance.Name), zap.String("project", instance.Project))
			}
		}
		resp, err := net.svc.Firewalls.Delete(project, firewall.Name).Do()
		if err != nil {
			net.logger.Error("Failed to remove firewall", zap.Error(err), zap.String("project", project), zap.String("firewall", firewall.Name))
			continue
		}
		if resp.Error != nil {
			net.logger.Error("Response failed to delete firewall", zap.Any("resp.error", resp.Error), zap.String("project", project), zap.String("firewall", firewall.Name))
			continue
		}
		net.logger.Info("Removed firewall", zap.String("project", project), zap.String("firewall", firewall.Name))
	}
}
