package compose

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

// ProjectFromDockerInspect 将 docker inspect 结果反推为单服务 compose Project。
// imageConfig 为镜像的默认配置（来自 Dockerfile），用于过滤掉镜像内置的默认值，
// 避免将 Dockerfile 中的 CMD/ENV/WORKDIR/USER 等冗余写入 compose yml。
// imageConfig 可为 nil，此时不做过滤。
func ProjectFromDockerInspect(info container.InspectResponse, imageConfig *dockerspec.DockerOCIImageConfig) (*types.Project, error) {
	if info.Config == nil || info.HostConfig == nil {
		return nil, fmt.Errorf("容器 inspect 数据不完整")
	}
	name := defaultString(strings.TrimPrefix(info.Name, "/"), info.ID)

	// entrypoint：与镜像默认相同则省略
	var entrypoint types.ShellCommand
	if len(info.Config.Entrypoint) > 0 && (imageConfig == nil || !sliceEqual(info.Config.Entrypoint, imageConfig.Entrypoint)) {
		entrypoint = types.ShellCommand(info.Config.Entrypoint)
	}

	// hostname：与容器名相同则省略（docker 默认行为）
	hostname := info.Config.Hostname
	if hostname == name {
		hostname = ""
	}

	svc := types.ServiceConfig{
		Name:          name,
		Image:         info.Config.Image,
		ContainerName: name,
		Command:       types.ShellCommand(diffCmd(info.Config.Cmd, imageConfig)),
		Entrypoint:    entrypoint,
		Environment:   sliceToEnv(diffEnv(info.Config.Env, imageConfig)),
		WorkingDir:    diffString(info.Config.WorkingDir, imageConfig, func(c *dockerspec.DockerOCIImageConfig) string { return c.WorkingDir }),
		User:          diffString(info.Config.User, imageConfig, func(c *dockerspec.DockerOCIImageConfig) string { return c.User }),
		Hostname:      hostname,
		Privileged:    info.HostConfig.Privileged,
		CapAdd:        []string(info.HostConfig.CapAdd),
		CapDrop:       []string(info.HostConfig.CapDrop),
		Restart:       restartPolicy(string(info.HostConfig.RestartPolicy.Name)),
		ExtraHosts:    extraHostsToMap(info.HostConfig.ExtraHosts),
		Labels:        info.Config.Labels,
	}

	applyInspectDNS(&svc, info)
	projectNetworks := applyInspectNetworks(&svc, info, name)
	applyInspectPorts(&svc, info)
	applyInspectVolumes(&svc, info)
	applyInspectResources(&svc, info)

	return &types.Project{
		Name:     name,
		Services: types.Services{name: svc},
		Networks: projectNetworks,
	}, nil
}

func applyInspectDNS(svc *types.ServiceConfig, info container.InspectResponse) {
	if len(info.HostConfig.DNS) > 0 {
		svc.DNS = info.HostConfig.DNS
	}
	if len(info.HostConfig.DNSSearch) > 0 {
		svc.DNSSearch = info.HostConfig.DNSSearch
	}
}

func applyInspectNetworks(svc *types.ServiceConfig, info container.InspectResponse, name string) types.Networks {
	networkMode := string(info.HostConfig.NetworkMode)
	// 内置模式（bridge/host/none/container:*/service:*）直接写入 network_mode，不解析 NetworkSettings
	if isBuiltinNetworkMode(networkMode) {
		svc.NetworkMode = networkMode
		return nil
	}
	// networkMode 为空或自定义网络名时，优先从 NetworkSettings.Networks 读取实际连接的网络
	if info.NetworkSettings == nil || len(info.NetworkSettings.Networks) == 0 {
		if networkMode != "" {
			svc.NetworkMode = networkMode
		}
		return nil
	}

	svc.Networks = make(map[string]*types.ServiceNetworkConfig, len(info.NetworkSettings.Networks))
	projectNetworks := make(types.Networks, len(info.NetworkSettings.Networks))
	for netName, ep := range info.NetworkSettings.Networks {
		netCfg := &types.ServiceNetworkConfig{}
		if ep.IPAMConfig != nil {
			netCfg.Ipv4Address = ep.IPAMConfig.IPv4Address
			netCfg.Ipv6Address = ep.IPAMConfig.IPv6Address
		}
		for _, alias := range ep.Aliases {
			if alias != name && !isContainerID(alias) {
				netCfg.Aliases = append(netCfg.Aliases, alias)
			}
		}
		svc.Networks[netName] = netCfg
		projectNetworks[netName] = types.NetworkConfig{External: true}
	}
	return projectNetworks
}

func applyInspectPorts(svc *types.ServiceConfig, info container.InspectResponse) {
	for containerPort, bindings := range info.HostConfig.PortBindings {
		target := parsePort(containerPort.Port())
		proto := defaultString(containerPort.Proto(), "tcp")
		if len(bindings) == 0 {
			svc.Ports = append(svc.Ports, types.ServicePortConfig{
				Target:   target,
				Protocol: proto,
				Mode:     "ingress",
			})
			continue
		}
		for _, b := range bindings {
			svc.Ports = append(svc.Ports, types.ServicePortConfig{
				Target:    target,
				Published: b.HostPort,
				HostIP:    b.HostIP,
				Protocol:  proto,
				Mode:      "ingress",
			})
		}
	}
}

func applyInspectVolumes(svc *types.ServiceConfig, info container.InspectResponse) {
	for _, m := range info.Mounts {
		if m.Destination == "" {
			continue
		}
		svc.Volumes = append(svc.Volumes, types.ServiceVolumeConfig{
			Type:     string(m.Type),
			Source:   m.Source,
			Target:   m.Destination,
			ReadOnly: !m.RW,
		})
	}
	if len(svc.Volumes) > 0 {
		return
	}
	for _, bind := range info.HostConfig.Binds {
		parts := strings.SplitN(bind, ":", 3)
		if len(parts) < 2 {
			continue
		}
		vol := types.ServiceVolumeConfig{
			Type:   types.VolumeTypeBind,
			Source: parts[0],
			Target: parts[1],
		}
		if len(parts) == 3 && strings.Contains(parts[2], "ro") {
			vol.ReadOnly = true
		}
		svc.Volumes = append(svc.Volumes, vol)
	}
}

func applyInspectResources(svc *types.ServiceConfig, info container.InspectResponse) {
	if info.HostConfig.Memory == 0 && info.HostConfig.NanoCPUs == 0 {
		return
	}
	svc.Deploy = &types.DeployConfig{
		Resources: types.Resources{Limits: &types.Resource{
			MemoryBytes: types.UnitBytes(info.HostConfig.Memory),
			NanoCPUs:    types.NanoCPUs(float64(info.HostConfig.NanoCPUs) / 1e9),
		}},
	}
}

// diffCmd 若容器 CMD 与镜像默认 CMD 相同则返回 nil（不写入 compose）
func diffCmd(containerCmd []string, imageConfig *dockerspec.DockerOCIImageConfig) []string {
	if imageConfig == nil {
		return containerCmd
	}
	if sliceEqual(containerCmd, imageConfig.Cmd) {
		return nil
	}
	return containerCmd
}

// diffEnv 过滤掉镜像默认 ENV，只保留容器中新增或覆盖的环境变量
func diffEnv(containerEnv []string, imageConfig *dockerspec.DockerOCIImageConfig) []string {
	if imageConfig == nil {
		return containerEnv
	}
	imageEnvSet := make(map[string]struct{}, len(imageConfig.Env))
	for _, e := range imageConfig.Env {
		imageEnvSet[e] = struct{}{}
	}
	var result []string
	for _, e := range containerEnv {
		if _, ok := imageEnvSet[e]; !ok {
			result = append(result, e)
		}
	}
	return result
}

// diffString 若容器字段值与镜像默认值相同则返回空字符串（不写入 compose）
func diffString(containerVal string, imageConfig *dockerspec.DockerOCIImageConfig, getter func(*dockerspec.DockerOCIImageConfig) string) string {
	if imageConfig == nil || containerVal == "" {
		return containerVal
	}
	if containerVal == getter(imageConfig) {
		return ""
	}
	return containerVal
}

// extraHostsToMap 将 []string{"host:ip"} 转换为 compose HostsList（map[string][]string）
func extraHostsToMap(hosts []string) types.HostsList {
	if len(hosts) == 0 {
		return nil
	}
	result := make(types.HostsList, len(hosts))
	for _, h := range hosts {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			result[parts[0]] = append(result[parts[0]], parts[1])
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// isContainerID 判断字符串是否为 Docker 容器 ID（12 或 64 位十六进制）
func isContainerID(s string) bool {
	if len(s) != 12 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// restartPolicy 将 Docker RestartPolicy.Name 映射到 compose restart 值。
// Docker 的 "" 和 "no" 均表示不重启，compose 对应 "no"；
// 其他值（always/on-failure/unless-stopped）直接透传。
func restartPolicy(name string) string {
	if name == "" || name == "no" {
		return "no"
	}
	return name
}

// sliceEqual 判断两个字符串切片是否完全相同
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// parsePort 解析 "8080" 或 "8080/tcp" 为端口号
func parsePort(s string) uint32 {
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	n, _ := strconv.Atoi(s)
	if n < 0 {
		return 0
	}
	return uint32(n)
}
