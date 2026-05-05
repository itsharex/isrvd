// ─── 系统探测 ───

export interface SystemProbe {
    agent: { available: boolean }
    apisix: { available: boolean }
    docker: { available: boolean }
    swarm: { available: boolean }
    compose: { available: boolean }
}

// ─── 系统统计 ───

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
