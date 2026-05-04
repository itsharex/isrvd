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

// ─── 路由权限 ───

// RouteAccess 路由访问级别（与后端枚举对应）
export const RouteAccessPerm = 0 // 需要权限控制（默认）
export const RouteAccessAuth = 1 // 登录即可访问
export const RouteAccessAnon = 2 // 匿名，无需认证

export interface RouteInfo {
    key: string
    module: string
    label: string
    access: number
}

// ─── 成员管理 ───

export interface MemberInfo {
    username: string
    homeDirectory: string
    passwordSet: boolean
    founder: boolean
    permissions: string[]
}

export interface MemberUpsertRequest {
    username: string
    // 写入时为空表示保留原值（仅更新场景）
    password: string
    homeDirectory: string
    permissions: string[]
}

// ─── API Token ───

export interface ApiTokenCreateRequest {
    name: string
    expiresIn: number // 过期时间（秒），0 表示永不过期
}

export interface ApiTokenCreateResponse {
    token: string
    name: string
}

// ─── 修改密码 ───

export interface ChangePasswordRequest {
    oldPassword: string
    newPassword: string
}
