// ─── 认证相关 ───

export interface AuthInfoResponse {
    mode: 'jwt' | 'header'
    username?: string
    member?: MemberInfo
}

export interface AuthLoginRequest {
    username: string
    password: string
}

export interface AuthLoginResponse {
    token: string
    username: string
}

// ─── 成员管理 ───

// 模块权限定义：'' 表示无权限，'r' 表示只读，'rw' 表示读写
export type ModulePermission = '' | 'r' | 'rw'

export interface MemberPermissions {
    overview: ModulePermission
    system: ModulePermission
    account: ModulePermission
    shell: ModulePermission
    filer: ModulePermission
    agent: ModulePermission
    apisix: ModulePermission
    docker: ModulePermission
    swarm: ModulePermission
    compose: ModulePermission
}

export interface MemberInfo {
    username: string
    homeDirectory: string
    passwordSet: boolean
    founder: boolean
    permissions: Record<string, string>
}

export interface MemberUpsertRequest {
    username: string
    // 写入时为空表示保留原值（仅更新场景）
    password: string
    homeDirectory: string
    permissions: Record<string, string>
}
