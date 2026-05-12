package docker

import (
	"context"
	"fmt"

	pkgdocker "isrvd/pkgs/docker"
)

// ImageList 列出镜像
func (s *Service) ImageList(ctx context.Context, all bool) ([]*pkgdocker.ImageInfo, error) {
	return s.docker.ImageList(ctx, all)
}

// ImageAction 镜像操作
func (s *Service) ImageAction(ctx context.Context, req pkgdocker.ImageActionRequest) error {
	return s.docker.ImageAction(ctx, req.ID, req.Action)
}

// ImageTag 镜像打标签
func (s *Service) ImageTag(ctx context.Context, req pkgdocker.ImageTagRequest) error {
	return s.docker.ImageTag(ctx, req.ID, req.RepoTag)
}

// ImageSearch 搜索镜像
func (s *Service) ImageSearch(ctx context.Context, term string) ([]*pkgdocker.ImageSearchResult, error) {
	if term == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}
	return s.docker.ImageSearch(ctx, term)
}

// ImageBuild 构建镜像
func (s *Service) ImageBuild(ctx context.Context, req pkgdocker.ImageBuildRequest) (map[string]string, error) {
	msg, err := s.docker.ImageBuild(ctx, req.Dockerfile, req.Tag)
	if err != nil {
		return nil, err
	}
	return map[string]string{"tag": req.Tag, "message": msg}, nil
}

// ImagePrune 清理未使用的镜像
func (s *Service) ImagePrune(ctx context.Context, req pkgdocker.ImagePruneRequest) (*pkgdocker.ImagePruneReport, error) {
	return s.docker.ImagePrune(ctx, req)
}

// ImageInspect 获取镜像详情
func (s *Service) ImageInspect(ctx context.Context, id string) (*pkgdocker.ImageDetail, error) {
	if id == "" {
		return nil, fmt.Errorf("镜像ID不能为空")
	}
	return s.docker.ImageInspect(ctx, id)
}
