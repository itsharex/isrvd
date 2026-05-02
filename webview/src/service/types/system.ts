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
