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
    AllConfigResponse,
    ComposeDeployResult,
    SystemStat,
    AuditLog
} from './types'

// API 服务类，统一管理所有 API 请求
class ApiService {
    // ==================== Overview 系统概览 ====================

    serviceProbe() {
        return http.get<SystemProbeResponse>('/api/overview/probe')
    }

    systemStatus() {
        return http.get<SystemStat>('/api/overview/status')
    }

    // ==================== System 系统相关 ====================

    getConfig(params?: Record<string, string>) {
        return http.get<AllConfigResponse>('/api/system/config', { params })
    }

    updateAllConfig(data: Partial<AllConfigResponse>) {
        return http.put<void>('/api/system/config', data)
    }

    getAuditLogs(params?: { username?: string; limit?: number }) {
        return http.get<AuditLog[]>('/api/system/audit/logs', { params })
    }

    // ==================== Account 账户相关 ====================

    login(data: { username: string; password: string }) {
        return http.post<AuthLoginResponse>('/api/account/login', data)
    }

    getMe() {
        return http.get<AuthInfoResponse>('/api/account/info')
    }

    listRoutes() {
        return http.get<RouteInfo[]>('/api/account/routes')
    }

    listMembers() {
        return http.get<MemberInfo[]>('/api/account/members')
    }

    createMember(data: MemberUpsertRequest) {
        return http.post<void>('/api/account/members', data)
    }

    updateMember(username: string, data: MemberUpsertRequest) {
        return http.put<void>(`/api/account/members/${encodeURIComponent(username)}`, data)
    }

    deleteMember(username: string) {
        return http.delete<void>(`/api/account/members/${encodeURIComponent(username)}`)
    }

    createApiToken(data: ApiTokenCreateRequest) {
        return http.post<ApiTokenCreateResponse>('/api/account/token', data)
    }

    // ==================== Filer 文件管理相关 ====================

    list(path: string) {
        return http.post<FilerListResponse>('/api/filer/list', { path })
    }

    delete(path: string) {
        return http.post<void>('/api/filer/delete', { path })
    }

    rename(path: string, target: string) {
        return http.post<void>('/api/filer/rename', { path, target })
    }

    mkdir(path: string) {
        return http.post<void>('/api/filer/mkdir', { path })
    }

    create(path: string, content = '') {
        return http.post<void>('/api/filer/create', { path, content })
    }

    read(path: string) {
        return http.post<FilerReadResponse>('/api/filer/read', { path })
    }

    modify(path: string, content: string) {
        return http.post<void>('/api/filer/modify', { path, content })
    }

    chmod(path: string, mode: string) {
        return http.post<void>('/api/filer/chmod', { path, mode })
    }

    zip(path: string) {
        return http.post<void>('/api/filer/zip', { path })
    }

    unzip(path: string) {
        return http.post<void>('/api/filer/unzip', { path })
    }

    upload(formData: FormData, config: AxiosRequestConfig = {}) {
        return http.post<void>('/api/filer/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            },
            ...config
        })
    }

    download(path: string) {
        return httpBlob.post('/api/filer/download', { path }, { responseType: 'blob' })
    }

    // ==================== APISIX 管理相关 ====================

    // 路由管理
    apisixListRoutes() {
        return http.get<ApisixRoute[]>('/api/apisix/routes')
    }

    apisixGetRoute(id: string) {
        return http.get<ApisixRoute>(`/api/apisix/routes/${id}`)
    }

    apisixCreateRoute(data: ApisixRoute) {
        return http.post('/api/apisix/routes', data)
    }

    apisixUpdateRoute(id: string, data: ApisixRoute) {
        return http.put(`/api/apisix/routes/${id}`, data)
    }

    apisixPatchRouteStatus(id: string, status: number) {
        return http.patch<void>(`/api/apisix/routes/${id}/status`, { status })
    }

    apisixDeleteRoute(id: string) {
        return http.delete<void>(`/api/apisix/routes/${id}`)
    }

    // Consumer 管理
    apisixListConsumers() {
        return http.get<ApisixConsumer[]>('/api/apisix/consumers')
    }

    apisixCreateConsumer(data: ApisixCreateConsumerRequest) {
        return http.post('/api/apisix/consumers', data)
    }

    apisixUpdateConsumer(username: string, data: ApisixUpdateConsumerRequest) {
        return http.put(`/api/apisix/consumers/${username}`, data)
    }

    apisixDeleteConsumer(username: string) {
        return http.delete<void>(`/api/apisix/consumers/${username}`)
    }

    // 白名单管理
    apisixGetWhitelist() {
        return http.get<ApisixRoute[]>('/api/apisix/whitelist')
    }

    apisixRevokeWhitelist(routeId: string, consumerName: string) {
        return http.post<void>('/api/apisix/whitelist/revoke', { route_id: routeId, consumer_name: consumerName })
    }

    // PluginConfig 管理（修复：使用横杠命名，与后端对齐）
    apisixListPluginConfigs() {
        return http.get<ApisixPluginConfig[]>('/api/apisix/plugin-configs')
    }

    apisixGetPluginConfig(id: string) {
        return http.get<ApisixPluginConfig>(`/api/apisix/plugin-configs/${id}`)
    }

    apisixCreatePluginConfig(data: ApisixCreatePluginConfigRequest) {
        return http.post('/api/apisix/plugin-configs', data)
    }

    apisixUpdatePluginConfig(id: string, data: ApisixUpdatePluginConfigRequest) {
        return http.put(`/api/apisix/plugin-configs/${id}`, data)
    }

    apisixDeletePluginConfig(id: string) {
        return http.delete<void>(`/api/apisix/plugin-configs/${id}`)
    }

    // Upstream 管理（修复：使用复数 upstreams）
    apisixListUpstreams() {
        return http.get<ApisixUpstream[]>('/api/apisix/upstreams')
    }

    apisixGetUpstream(id: string) {
        return http.get<ApisixUpstream>(`/api/apisix/upstreams/${id}`)
    }

    apisixCreateUpstream(data: ApisixCreateUpstreamRequest) {
        return http.post('/api/apisix/upstreams', data)
    }

    apisixUpdateUpstream(id: string, data: ApisixUpdateUpstreamRequest) {
        return http.put(`/api/apisix/upstreams/${id}`, data)
    }

    apisixDeleteUpstream(id: string) {
        return http.delete<void>(`/api/apisix/upstreams/${id}`)
    }

    // SSL 管理（修复：使用复数 ssls）
    apisixListSSLs() {
        return http.get<ApisixSSL[]>('/api/apisix/ssls')
    }

    apisixGetSSL(id: string) {
        return http.get<ApisixSSL>(`/api/apisix/ssls/${id}`)
    }

    apisixCreateSSL(data: ApisixCreateSSLRequest) {
        return http.post('/api/apisix/ssls', data)
    }

    apisixUpdateSSL(id: string, data: ApisixUpdateSSLRequest) {
        return http.put(`/api/apisix/ssls/${id}`, data)
    }

    apisixDeleteSSL(id: string) {
        return http.delete<void>(`/api/apisix/ssls/${id}`)
    }

    apisixListPlugins() {
        return http.get<Record<string, { schema: Record<string, unknown> }>>('/api/apisix/plugins')
    }

    // ==================== Docker 服务相关 ====================

    // Docker 概览信息
    dockerInfo() {
        return http.get<DockerInfo>('/api/docker/info')
    }

    // 容器管理（修复：使用复数 containers）
    listContainers(all = false) {
        return http.get<DockerContainerInfo[]>('/api/docker/containers', { params: { all } })
    }

    containerAction(id: string, action: string) {
        return http.post<void>(`/api/docker/containers/${id}/action`, { action })
    }

    createContainer(data: DockerContainerCreateRequest) {
        return http.post('/api/docker/containers', data)
    }

    containerLogs(id: string, tail = '100') {
        return http.get<{ logs: string[] }>(`/api/docker/containers/${id}/logs`, { params: { tail } })
    }

    containerStats(id: string) {
        return http.get<DockerContainerStatsResponse>(`/api/docker/containers/${id}/stats`)
    }

    getContainerCompose(name: string) {
        return http.get<{ content: string }>(`/api/compose/docker/${name}`)
    }

    // 镜像管理（修复：使用复数 images）
    listImages(all = false) {
        return http.get<DockerImageInfo[]>('/api/docker/images', { params: { all } })
    }

    imageInspect(id: string) {
        return http.get<DockerImageInspectResponse>(`/api/docker/images/${id}`)
    }

    imageAction(id: string, action: string) {
        return http.post<void>(`/api/docker/images/${id}/action`, { action })
    }

    imageTag(id: string, repoTag: string) {
        return http.post<void>('/api/docker/images/tag', { id, repoTag })
    }

    imageSearch(term: string) {
        return http.get<DockerImageSearchResult[]>(`/api/docker/images/search/${encodeURIComponent(term)}`)
    }

    imageBuild(dockerfile: string, tag = '') {
        return http.post<void>('/api/docker/images/build', { dockerfile, tag })
    }

    // 网络管理（修复：使用复数 networks）
    listNetworks() {
        return http.get<DockerNetworkInfo[]>('/api/docker/networks')
    }

    networkInspect(id: string) {
        return http.get<DockerNetworkInspectResponse>(`/api/docker/networks/${id}`)
    }

    networkAction(id: string, action: string) {
        return http.post<void>(`/api/docker/networks/${id}/action`, { action })
    }

    createNetwork(data: DockerNetworkCreateRequest) {
        return http.post('/api/docker/networks', data)
    }

    // 卷管理（修复：使用复数 volumes）
    listVolumes() {
        return http.get<DockerVolumeInfo[]>('/api/docker/volumes')
    }

    volumeInspect(name: string) {
        return http.get<DockerVolumeInspectResponse>(`/api/docker/volumes/${encodeURIComponent(name)}`)
    }

    volumeAction(name: string, action: string) {
        return http.post<void>(`/api/docker/volumes/${encodeURIComponent(name)}/action`, { action })
    }

    createVolume(data: DockerVolumeCreateRequest) {
        return http.post('/api/docker/volumes', data)
    }

    // 镜像仓库管理
    listRegistries() {
        return http.get<DockerRegistryInfo[]>('/api/docker/registries')
    }

    createRegistry(data: DockerRegistryUpsertRequest) {
        return http.post<void>('/api/docker/registries', data)
    }

    updateRegistry(url: string, data: DockerRegistryUpsertRequest) {
        return http.put<void>('/api/docker/registries', data, { params: { url } })
    }

    deleteRegistry(url: string) {
        return http.delete<void>('/api/docker/registries', { params: { url } })
    }

    pushImage(image: string, registryUrl: string, namespace: string) {
        return http.post<void>('/api/docker/images/push', { image, registryUrl, namespace })
    }

    pullFromRegistry(image: string, registryUrl: string, namespace: string) {
        return http.post<void>('/api/docker/images/pull', { image, registryUrl, namespace })
    }

    // ==================== Docker Swarm 管理相关 ====================

    swarmInfo() {
        return http.get<SwarmInfo>('/api/swarm/info')
    }

    swarmListNodes() {
        return http.get<SwarmNodeDTO[]>('/api/swarm/nodes')
    }

    swarmGetJoinTokens() {
        return http.get<{ worker: string; manager: string }>('/api/swarm/tokens')
    }

    swarmInspectNode(id: string) {
        return http.get<SwarmNodeInspect>(`/api/swarm/nodes/${id}`)
    }

    NodeDTOAction(id: string, action: string) {
        return http.post<void>(`/api/swarm/nodes/${id}/action`, { action })
    }

    // 服务管理（修复：使用复数 services，路径对齐后端）
    swarmListServices() {
        return http.get<SwarmServiceInfo[]>('/api/swarm/services')
    }

    swarmInspectService(id: string) {
        return http.get<SwarmServiceDetail>(`/api/swarm/services/${id}`)
    }

    swarmServiceAction(id: string, action: string, replicas?: number) {
        return http.post<void>(`/api/swarm/services/${id}/action`, { action, replicas })
    }

    swarmCreateService(data: SwarmCreateServiceRequest) {
        return http.post('/api/swarm/services', data)
    }

    swarmRedeployService(id: string) {
        return http.post<void>(`/api/swarm/services/${id}/force-update`)
    }

    swarmGetServiceCompose(name: string) {
        return http.get<{ content: string }>(`/api/compose/swarm/${name}`)
    }

    swarmServiceLogs(id: string, tail = '100') {
        return http.get<{ logs: string[] }>(`/api/swarm/services/${id}/logs`, { params: { tail } })
    }

    swarmListTasks(serviceID = '') {
        return http.get<SwarmTask[]>('/api/swarm/tasks', { params: serviceID ? { serviceID } : {} })
    }

    // ==================== Compose 部署 ====================

    composeDeployDocker(data: { content: string; projectName: string; initURL?: string; initFile?: File }) {
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

    composeDeploySwarm(data: { content: string; projectName: string }) {
        return http.post<ComposeDeployResult>('/api/compose/swarm/deploy', data)
    }

    composeRedeployDocker(name: string, data: { content: string }) {
        return http.post<ComposeDeployResult>(`/api/compose/docker/${name}/redeploy`, data)
    }

    composeRedeploySwarm(name: string, data: { content: string }) {
        return http.post<ComposeDeployResult>(`/api/compose/swarm/${name}/redeploy`, data)
    }
}

// 导出单例实例
export default new ApiService()
