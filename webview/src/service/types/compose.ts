// ─── Compose 统一部署 ───

export type ComposeDeployTarget = 'docker' | 'swarm'

// 统一的 compose 部署请求
// 项目名（projectName）从 compose 文件中的 name: 字段自动获取，无需前端填写
export interface ComposeDeploy {
    content: string         // 完整 compose yaml 文本（前端已完成 ${VAR} 插值）
    initURL?: string        // 可选：附加运行文件 zip 下载地址
    initFile?: File         // 可选：附加运行文件（与 initURL 互斥，文件优先）
}

// Compose Redeploy 请求
// - content 非空，serviceName 为空：全量重建
// - serviceName + image 非空：按服务更新镜像重建
export interface ComposeRedeploy {
    content?: string
    serviceName?: string
    image?: string
}

export interface ComposeDeployResult {
    target: ComposeDeployTarget
    projectName: string     // 实际使用的项目名
    items: string[]
    installDir?: string     // 仅 docker 落盘时返回
}

// ─── 应用市场协议 ───

// 回传给父组件的精简 payload
export interface ComposeMarketplacePick {
    name: string
    compose: string
    initURL?: string
}
