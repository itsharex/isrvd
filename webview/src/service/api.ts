import type { AxiosRequestConfig } from 'axios'
import { http, httpBlob } from './axios'
import type {
    DockerContainerCreateRequest,
    DockerContainerInfo, DockerContainerStatsResponse,
    DockerImageInfo, DockerImageInspectResponse, DockerImageSearchResult,
    DockerNetworkInfo, DockerNetworkInspectResponse, DockerNetworkCreateRequest,
    DockerVolumeInfo, DockerVolumeInspectResponse, DockerVolumeCreateRequest,
    DockerRegistryInfo, DockerRegistryUpsertRequest,
    SwarmInfo, SwarmNodeDTO, SwarmNodeInspect,
    SwarmServiceInfo, SwarmServiceDetail, SwarmTask,
    SwarmCreateServiceRequest,
    ApisixRoute, ApisixConsumer, ApisixCreateConsumerRequest, ApisixUpdateConsumerRequest,
    ApisixCreatePluginConfigRequest,
    ApisixCreateSSLRequest, ApisixCreateUpstreamRequest, ApisixPluginConfig, ApisixSSL,
    ApisixUpstream, ApisixUpdatePluginConfigRequest, ApisixUpdateSSLRequest, ApisixUpdateUpstreamRequest,
    SystemProbeResponse, DockerInfo,
    FilerListResponse, FilerReadResponse,
    AuthLoginResponse, AuthInfoResponse,
    MemberInfo, MemberUpsertRequest, RouteInfo,
    ApiTokenCreateRequest, ApiTokenCreateResponse,
    ChangePasswordRequest,
    AllConfigResponse,
    ComposeDeployResult,
    SystemStat,
    AuditLog
} from './types'

// API 服务类，统一管理所有 API 请求
class ApiService {
    // ==================== Overview 系统概览 ====================

    overviewProbe() {
        return http.get<SystemProbeResponse>('/api/overview/probe')
    }

    overviewStatus() {
        return http.get<SystemStat>('/api/overview/status')
    }

    // ==================== System 系统相关 ====================

    systemConfig(params?: Record<string, string>) {
        return http.get<AllConfigResponse>('/api/system/config', { params })
    }

    systemConfigUpdate(data: Partial<AllConfigResponse>) {
        return http.put<void>('/api/system/config', data)
    }

    systemAuditLogs(params?: { username?: string; limit?: number }) {
        return http.get<AuditLog[]>('/api/system/audit/logs', { params })
    }

    // ==================== Account 账户相关 ====================

    accountLogin(data: { username: string; password: string }) {
        return http.post<AuthLoginResponse>('/api/account/login', data)
    }

    accountInfo() {
        return http.get<AuthInfoResponse>('/api/account/info')
    }

    accountRouteList() {
        return http.get<RouteInfo[]>('/api/account/routes')
    }

    accountMemberList() {
        return http.get<MemberInfo[]>('/api/account/members')
    }

    accountMemberCreate(data: MemberUpsertRequest) {
        return http.post<void>('/api/account/member', data)
    }

    accountMemberUpdate(username: string, data: MemberUpsertRequest) {
        return http.put<void>(`/api/account/member/${encodeURIComponent(username)}`, data)
    }

    accountMemberDelete(username: string) {
        return http.delete<void>(`/api/account/member/${encodeURIComponent(username)}`)
    }

    accountTokenCreate(data: ApiTokenCreateRequest) {
        return http.post<ApiTokenCreateResponse>('/api/account/token', data)
    }

    accountPasswordChange(data: ChangePasswordRequest) {
        return http.put<void>('/api/account/password', data)
    }

    // ==================== Filer 文件管理相关 ====================

    filerList(path: string) {
        return http.post<FilerListResponse>('/api/filer/list', { path })
    }

    filerDelete(path: string) {
        return http.post<void>('/api/filer/delete', { path })
    }

    filerRename(path: string, target: string) {
        return http.post<void>('/api/filer/rename', { path, target })
    }

    filerMkdir(path: string) {
        return http.post<void>('/api/filer/mkdir', { path })
    }

    filerCreate(path: string, content = '') {
        return http.post<void>('/api/filer/create', { path, content })
    }

    filerRead(path: string) {
        return http.post<FilerReadResponse>('/api/filer/read', { path })
    }

    filerModify(path: string, content: string) {
        return http.post<void>('/api/filer/modify', { path, content })
    }

    filerChmod(path: string, mode: string) {
        return http.post<void>('/api/filer/chmod', { path, mode })
    }

    filerZip(path: string) {
        return http.post<void>('/api/filer/zip', { path })
    }

    filerUnzip(path: string) {
        return http.post<void>('/api/filer/unzip', { path })
    }

    filerUpload(formData: FormData, config: AxiosRequestConfig = {}) {
        return http.post<void>('/api/filer/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            },
            ...config
        })
    }

    filerDownload(path: string) {
        return httpBlob.post('/api/filer/download', { path }, { responseType: 'blob' })
    }

    // ==================== APISIX 管理相关 ====================

    // 路由管理
    apisixRouteList() {
        return http.get<ApisixRoute[]>('/api/apisix/routes')
    }

    apisixRoute(id: string) {
        return http.get<ApisixRoute>(`/api/apisix/route/${id}`)
    }

    apisixRouteCreate(data: ApisixRoute) {
        return http.post('/api/apisix/route', data)
    }

    apisixRouteUpdate(id: string, data: ApisixRoute) {
        return http.put(`/api/apisix/route/${id}`, data)
    }

    apisixRouteStatus(id: string, status: number) {
        return http.patch<void>(`/api/apisix/route/${id}/status`, { status })
    }

    apisixRouteDelete(id: string) {
        return http.delete<void>(`/api/apisix/route/${id}`)
    }

    // Consumer 管理
    apisixConsumerList() {
        return http.get<ApisixConsumer[]>('/api/apisix/consumers')
    }

    apisixConsumerCreate(data: ApisixCreateConsumerRequest) {
        return http.post('/api/apisix/consumer', data)
    }

    apisixConsumerUpdate(username: string, data: ApisixUpdateConsumerRequest) {
        return http.put(`/api/apisix/consumer/${username}`, data)
    }

    apisixConsumerDelete(username: string) {
        return http.delete<void>(`/api/apisix/consumer/${username}`)
    }

    // 白名单管理
    apisixWhitelist() {
        return http.get<ApisixRoute[]>('/api/apisix/whitelist')
    }

    apisixWhitelistRevoke(routeId: string, consumerName: string) {
        return http.post<void>('/api/apisix/whitelist/revoke', { route_id: routeId, consumer_name: consumerName })
    }

    // PluginConfig 管理
    apisixPluginConfigList() {
        return http.get<ApisixPluginConfig[]>('/api/apisix/plugin-configs')
    }

    apisixPluginConfig(id: string) {
        return http.get<ApisixPluginConfig>(`/api/apisix/plugin-config/${id}`)
    }

    apisixPluginConfigCreate(data: ApisixCreatePluginConfigRequest) {
        return http.post('/api/apisix/plugin-config', data)
    }

    apisixPluginConfigUpdate(id: string, data: ApisixUpdatePluginConfigRequest) {
        return http.put(`/api/apisix/plugin-config/${id}`, data)
    }

    apisixPluginConfigDelete(id: string) {
        return http.delete<void>(`/api/apisix/plugin-config/${id}`)
    }

    // Upstream 管理
    apisixUpstreamList() {
        return http.get<ApisixUpstream[]>('/api/apisix/upstreams')
    }

    apisixUpstream(id: string) {
        return http.get<ApisixUpstream>(`/api/apisix/upstream/${id}`)
    }

    apisixUpstreamCreate(data: ApisixCreateUpstreamRequest) {
        return http.post('/api/apisix/upstream', data)
    }

    apisixUpstreamUpdate(id: string, data: ApisixUpdateUpstreamRequest) {
        return http.put(`/api/apisix/upstream/${id}`, data)
    }

    apisixUpstreamDelete(id: string) {
        return http.delete<void>(`/api/apisix/upstream/${id}`)
    }

    // SSL 管理
    apisixSSLList() {
        return http.get<ApisixSSL[]>('/api/apisix/ssls')
    }

    apisixSSL(id: string) {
        return http.get<ApisixSSL>(`/api/apisix/ssl/${id}`)
    }

    apisixSSLCreate(data: ApisixCreateSSLRequest) {
        return http.post('/api/apisix/ssl', data)
    }

    apisixSSLUpdate(id: string, data: ApisixUpdateSSLRequest) {
        return http.put(`/api/apisix/ssl/${id}`, data)
    }

    apisixSSLDelete(id: string) {
        return http.delete<void>(`/api/apisix/ssl/${id}`)
    }

    apisixPluginList() {
        return http.get<Record<string, { schema: Record<string, unknown> }>>('/api/apisix/plugins')
    }

    // ==================== Docker 服务相关 ====================

    // Docker 概览信息
    dockerInfo() {
        return http.get<DockerInfo>('/api/docker/info')
    }

    // 容器管理
    dockerContainerList(all = false) {
        return http.get<DockerContainerInfo[]>('/api/docker/containers', { params: { all } })
    }

    dockerContainerAction(id: string, action: string) {
        return http.post<void>(`/api/docker/container/${id}/action`, { action })
    }

    dockerContainerCreate(data: DockerContainerCreateRequest) {
        return http.post('/api/docker/container', data)
    }

    dockerContainerLogs(id: string, tail = '100') {
        return http.get<{ logs: string[] }>(`/api/docker/container/${id}/logs`, { params: { tail } })
    }

    dockerContainerStats(id: string) {
        return http.get<DockerContainerStatsResponse>(`/api/docker/container/${id}/stats`)
    }

    dockerContainerCompose(name: string) {
        return http.get<{ content: string }>(`/api/compose/docker/${name}`)
    }

    // 镜像管理
    dockerImageList(all = false) {
        return http.get<DockerImageInfo[]>('/api/docker/images', { params: { all } })
    }

    dockerImage(id: string) {
        return http.get<DockerImageInspectResponse>(`/api/docker/image/${id}`)
    }

    dockerImageAction(id: string, action: string) {
        return http.post<void>(`/api/docker/image/${id}/action`, { action })
    }

    dockerImageTag(id: string, repoTag: string) {
        return http.post<void>(`/api/docker/image/${id}/tag`, { repoTag })
    }

    dockerImageSearch(name: string) {
        return http.get<DockerImageSearchResult[]>('/api/docker/images/search', { params: { name } })
    }

    dockerImageBuild(dockerfile: string, tag = '') {
        return http.post<void>('/api/docker/image/build', { dockerfile, tag })
    }

    dockerImagePush(image: string, registryUrl: string, namespace: string) {
        return http.post<void>('/api/docker/image/push', { image, registryUrl, namespace })
    }

    dockerImagePull(image: string, registryUrl: string, namespace: string) {
        return http.post<void>('/api/docker/image/pull', { image, registryUrl, namespace })
    }

    // 网络管理
    dockerNetworkList() {
        return http.get<DockerNetworkInfo[]>('/api/docker/networks')
    }

    dockerNetwork(id: string) {
        return http.get<DockerNetworkInspectResponse>(`/api/docker/network/${id}`)
    }

    dockerNetworkAction(id: string, action: string) {
        return http.post<void>(`/api/docker/network/${id}/action`, { action })
    }

    dockerNetworkCreate(data: DockerNetworkCreateRequest) {
        return http.post('/api/docker/network', data)
    }

    // 卷管理
    dockerVolumeList() {
        return http.get<DockerVolumeInfo[]>('/api/docker/volumes')
    }

    dockerVolume(name: string) {
        return http.get<DockerVolumeInspectResponse>(`/api/docker/volume/${encodeURIComponent(name)}`)
    }

    dockerVolumeAction(name: string, action: string) {
        return http.post<void>(`/api/docker/volume/${encodeURIComponent(name)}/action`, { action })
    }

    dockerVolumeCreate(data: DockerVolumeCreateRequest) {
        return http.post('/api/docker/volume', data)
    }

    // 镜像仓库管理
    dockerRegistryList() {
        return http.get<DockerRegistryInfo[]>('/api/docker/registries')
    }

    dockerRegistryCreate(data: DockerRegistryUpsertRequest) {
        return http.post<void>('/api/docker/registry', data)
    }

    dockerRegistryUpdate(url: string, data: DockerRegistryUpsertRequest) {
        return http.put<void>('/api/docker/registry', data, { params: { url } })
    }

    dockerRegistryDelete(url: string) {
        return http.delete<void>('/api/docker/registry', { params: { url } })
    }

    // ==================== Docker Swarm 管理相关 ====================

    swarmInfo() {
        return http.get<SwarmInfo>('/api/swarm/info')
    }

    swarmNodeList() {
        return http.get<SwarmNodeDTO[]>('/api/swarm/nodes')
    }

    swarmTokens() {
        return http.get<{ worker: string; manager: string }>('/api/swarm/tokens')
    }

    swarmNode(id: string) {
        return http.get<SwarmNodeInspect>(`/api/swarm/node/${id}`)
    }

    swarmNodeAction(id: string, action: string) {
        return http.post<void>(`/api/swarm/node/${id}/action`, { action })
    }

    // 服务管理
    swarmServiceList() {
        return http.get<SwarmServiceInfo[]>('/api/swarm/services')
    }

    swarmService(id: string) {
        return http.get<SwarmServiceDetail>(`/api/swarm/service/${id}`)
    }

    swarmServiceAction(id: string, action: string, replicas?: number) {
        return http.post<void>(`/api/swarm/service/${id}/action`, { action, replicas })
    }

    swarmServiceCreate(data: SwarmCreateServiceRequest) {
        return http.post('/api/swarm/service', data)
    }

    swarmServiceRedeploy(id: string) {
        return http.post<void>(`/api/swarm/service/${id}/force-update`)
    }

    swarmServiceCompose(name: string) {
        return http.get<{ content: string }>(`/api/compose/swarm/${name}`)
    }

    swarmServiceLogs(id: string, tail = '100') {
        return http.get<{ logs: string[] }>(`/api/swarm/service/${id}/logs`, { params: { tail } })
    }

    swarmTaskList(serviceID = '') {
        return http.get<SwarmTask[]>('/api/swarm/tasks', { params: serviceID ? { serviceID } : {} })
    }

    // ==================== Compose 部署 ====================

    composeDockerDeploy(data: { content: string; projectName: string; initURL?: string; initFile?: File }) {
        const form = new FormData()
        form.append('content', data.content)
        form.append('projectName', data.projectName)
        // 文件优先，二者互斥
        if (data.initFile) {
            form.append('initFile', data.initFile)
        } else if (data.initURL) {
            form.append('initURL', data.initURL)
        }
        return http.post<ComposeDeployResult>('/api/compose/docker/deploy', form)
    }

    composeSwarmDeploy(data: { content: string; projectName: string }) {
        return http.post<ComposeDeployResult>('/api/compose/swarm/deploy', data)
    }

    composeDockerRedeploy(name: string, data: { content: string }) {
        return http.post<ComposeDeployResult>(`/api/compose/docker/${name}/redeploy`, data)
    }

    composeSwarmRedeploy(name: string, data: { content: string }) {
        return http.post<ComposeDeployResult>(`/api/compose/swarm/${name}/redeploy`, data)
    }
}

// 导出单例实例
export default new ApiService()
