// ─── 系统探测 ───

export interface SystemProbeResponse {
    apisix: { available: boolean }
    docker: { available: boolean }
    swarm: { available: boolean }
}

// ─── 系统设置 ───

export interface SystemServerSettings {
    debug: boolean
    listenAddr: string
    // 写入时为空表示保留原值
    jwtSecret: string
    // 读取时后端返回，标识是否已设置
    jwtSecretSet?: boolean
    proxyHeaderName: string
    rootDirectory: string
}

export interface SystemApisixSettings {
    adminUrl: string
    // 写入时为空表示保留原值
    adminKey: string
    // 读取时后端返回，标识是否已设置
    adminKeySet?: boolean
}

export interface SystemAgentSettings {
    model: string
    baseUrl: string
    // 写入时为空表示保留原值
    apiKey: string
    // 读取时后端返回，标识是否已设置
    apiKeySet?: boolean
}

export interface SystemDockerSettings {
    host: string
    containerRoot: string
}

export interface SystemMarketplaceSettings {
    url: string
}

export interface SystemLinkSetting {
    label: string
    url: string
    icon: string
}

export interface SystemAllSettings {
    server: SystemServerSettings
    agent: SystemAgentSettings
    apisix: SystemApisixSettings
    docker: SystemDockerSettings
    marketplace: SystemMarketplaceSettings
    links: SystemLinkSetting[]
}

// ─── 系统统计（/api/system/stats 响应） ───

export interface SystemVersionCheck {
    latest: string
    update: boolean
    release: string
}

export interface SystemNetInterface {
    Name: string
    BytesRecv: number
    BytesSent: number
}

export interface SystemDiskPartition {
    Device: string
    Mountpoint: string
    Fstype: string
    Used: number
    Total: number
}

export interface SystemDiskIO {
    Name: string
    ReadBytes: number
    WriteBytes: number
}

export interface SystemGoRuntimeStat {
    version: string
    numCPU: number
    numGoroutine: number
    HeapAlloc: number
    HeapInuse: number
    Sys: number
    StackInuse: number
    TotalAlloc: number
    NumGC: number
    LastGC: number
}

export interface SystemInfo {
    HostName: string
    Platform: string
    KernelArch: string
    Uptime: number
    CpuCore: number
    CpuCoreLogic: number
    CpuModel: string[]
    CpuPercent: number[]
    MemoryUsed: number
    MemoryTotal: number
    DiskTotal: number
    DiskUsed: number
    NetInterface: SystemNetInterface[]
    DiskPartition: SystemDiskPartition[]
}

export interface SystemGPU {
    index: number
    deviceKey?: string
    name: string
    vendor: string         // "nvidia" | "amd" | "intel"
    memoryUsed: number
    memoryTotal: number
    utilization: number    // 0-100
    temperature: number    // -1 = N/A
    powerUsage: number     // watts, -1 = N/A
    fanSpeed: number       // percent, -1 = N/A
}

export interface SystemStat {
    system: SystemInfo
    diskIO: SystemDiskIO[]
    gpu: SystemGPU[]
    go: SystemGoRuntimeStat
    version: string
    versionCheck: SystemVersionCheck
}

// ─── 审计日志 ───

export interface AuditLog {
    timestamp: string
    username: string
    method: string
    uri: string
    body: string
    ip: string
    statusCode: number
    success: boolean
    duration: number
}
