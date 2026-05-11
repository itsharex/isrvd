package compose

import (
	"fmt"
	"strconv"

	"github.com/compose-spec/compose-go/v2/types"

	"isrvd/pkgs/swarm"
)

// ProjectFromSwarmInspect 将 ServiceInfo 反推为单服务 compose Project
func ProjectFromSwarmInspect(info *swarm.ServiceInfo) (*types.Project, error) {
	if info == nil || info.Image == "" {
		return nil, fmt.Errorf("swarm 服务数据不完整")
	}

	name := defaultString(info.Name, info.ID)
	svc := types.ServiceConfig{
		Name:        name,
		Image:       info.Image,
		Environment: sliceToEnv(info.Env),
		Command:     types.ShellCommand(info.Args),
		Labels:      info.Labels,
	}

	// deploy
	if info.Mode == "global" {
		svc.Deploy = &types.DeployConfig{Mode: "global"}
	} else if info.Replicas != nil {
		r := int(*info.Replicas)
		svc.Deploy = &types.DeployConfig{Mode: "replicated", Replicas: &r}
	}
	if len(info.Constraints) > 0 {
		if svc.Deploy == nil {
			svc.Deploy = &types.DeployConfig{}
		}
		svc.Deploy.Placement.Constraints = info.Constraints
	}

	// ports
	for _, p := range info.Ports {
		svc.Ports = append(svc.Ports, types.ServicePortConfig{
			Target:    p.TargetPort,
			Published: strconv.Itoa(int(p.PublishedPort)),
			Protocol:  defaultString(p.Protocol, "tcp"),
			Mode:      defaultString(p.PublishMode, "ingress"),
		})
	}

	// volumes
	for _, m := range info.Mounts {
		svc.Volumes = append(svc.Volumes, types.ServiceVolumeConfig{
			Type:     m.Type,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// networks
	var projectNetworks types.Networks
	if len(info.Networks) > 0 {
		svc.Networks = make(map[string]*types.ServiceNetworkConfig, len(info.Networks))
		projectNetworks = make(types.Networks, len(info.Networks))
		for _, n := range info.Networks {
			svc.Networks[n] = &types.ServiceNetworkConfig{}
			projectNetworks[n] = types.NetworkConfig{Name: n, Driver: "overlay"}
		}
	}

	return &types.Project{
		Name:     name,
		Services: types.Services{name: svc},
		Networks: projectNetworks,
	}, nil
}
