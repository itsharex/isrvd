package docker

import (
	"context"
	"fmt"

	pkgdocker "isrvd/pkgs/docker"
)

// NetworkList 列出网络
func (s *Service) NetworkList(ctx context.Context) (any, error) {
	return s.docker.NetworkList(ctx)
}

// NetworkAction 网络操作
func (s *Service) NetworkAction(ctx context.Context, req pkgdocker.NetworkActionRequest) error {
	return s.docker.NetworkAction(ctx, req.ID, req.Action)
}

// NetworkCreate 创建网络
func (s *Service) NetworkCreate(ctx context.Context, req pkgdocker.NetworkCreateRequest) (map[string]string, error) {
	id, err := s.docker.NetworkCreate(ctx, req.Name, req.Driver)
	if err != nil {
		return nil, err
	}
	return map[string]string{"id": id, "name": req.Name}, nil
}

// NetworkInspect 获取网络详情
func (s *Service) NetworkInspect(ctx context.Context, id string) (any, error) {
	if id == "" {
		return nil, fmt.Errorf("网络ID不能为空")
	}
	return s.docker.NetworkInspect(ctx, id)
}

// VolumeList 列出卷
func (s *Service) VolumeList(ctx context.Context) (any, error) {
	return s.docker.VolumeList(ctx)
}

// VolumeAction 卷操作
func (s *Service) VolumeAction(ctx context.Context, req pkgdocker.VolumeActionRequest) error {
	return s.docker.VolumeAction(ctx, req.Name, req.Action)
}

// VolumeCreate 创建卷
func (s *Service) VolumeCreate(ctx context.Context, req pkgdocker.VolumeCreateRequest) (map[string]string, error) {
	name, mountpoint, err := s.docker.VolumeCreate(ctx, req.Name, req.Driver)
	if err != nil {
		return nil, err
	}
	return map[string]string{"name": name, "mountpoint": mountpoint}, nil
}

// VolumeInspect 获取卷详情
func (s *Service) VolumeInspect(ctx context.Context, name string) (any, error) {
	if name == "" {
		return nil, fmt.Errorf("卷名称不能为空")
	}
	return s.docker.VolumeInspect(ctx, name)
}
