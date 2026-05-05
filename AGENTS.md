# AGENTS.md — isrvd Agent Operating Guide

> 本文件是 `isrvd` 仓库唯一代码规范入口。目标：可执行、可验证、低歧义。

---

## 1) 指令优先级

冲突时：用户明确需求 → 安全稳定性 → 本规范 → 现有代码风格
无法消解时：不泄露敏感信息、不引入破坏性变更、不修改需求范围外逻辑

---

## 2) 工作流

- **先理解再改动**：定位模块、调用链、类型与边界，不基于猜测修改
- **小步提交**：最小可行改动，每步可解释"为什么改、改了什么、如何验证"
- **变更后验证**：执行相关静态检查；无法完整验证时说明风险

---

## 3) 项目架构

### 分层依赖方向

`config → internal/registry → pkgs → internal/service → internal/server`

- `pkgs/`：原生客户端层，返回 SDK/底层类型
- `internal/service/`：业务组合、类型转换、参数校验
- `internal/server/`：HTTP 入口，解析请求 → 调用 service → 返回响应
- `internal/registry/`：外部服务初始化、生命周期管理与可用性检查

### 禁止

- `pkgs/` 依赖 `internal/`；handler 中堆叠业务逻辑；service/handler 直接从配置创建外部客户端

### 内聚与耦合

- 同领域功能聚合同包（`pkgs/docker/`、`pkgs/swarm/`），类型就近定义，不集中 `types.go`
- 层间通过接口或注入解耦；前端组件通过 `Provide/Inject` 获取全局状态
- 判定：改一个功能只需改一处；若需同时改多个包同名函数，说明内聚不足

---

## 4) 后端编码规范（Go）

### HTTP 与响应

- 状态码用 `net/http` 常量；成功 `helper.RespondSuccess`，失败 `helper.RespondError`
- 绑定优先 `ShouldBindJSON/ShouldBindQuery/ShouldBindURI`，绑定失败返回 `err.Error()`
- WebSocket 统一 `helper.WsUpgrader`，不在 handler 中定义私有 upgrader

### 错误处理

- 禁止 `fmt.Errorf(err.Error())`，使用 `fmt.Errorf("...: %w", err)` 包装
- `pkgs` 调外部失败时透传原始错误；仅本层自有逻辑失败时加上下文

### 命名与日志

- 缩写全大写：`CPU`、`ID`、`URL`、`HTTP`
- 日志统一 `logman`，禁止 `log.Println`/`fmt.Println`；使用键值对不拼接字符串

### 方法命名规范（强制）

**Handler（`internal/server/`）** — 格式：`{module}{Resource}{Action}`

| 操作 | 命名模式 | 示例 |
|---|---|---|
| 列表 | `{module}{Resource}List` | `dockerContainerList`、`apisixRouteList` |
| 单条 | `{module}{Resource}Inspect` | `dockerImageInspect`、`swarmNodeInspect` |
| 创建/更新/删除 | `{module}{Resource}Create/Update/Delete` | `apisixRouteCreate`、`dockerImageDelete` |
| 操作/动作 | `{module}{Resource}Action` | `dockerContainerAction`、`swarmNodeAction` |
| 状态切换 | `{module}{Resource}StatusPatch` | `apisixRouteStatusPatch` |
| 日志/统计 | `{module}{Resource}Logs/Stats` | `dockerContainerLogs`、`dockerContainerStats` |

**Service / Pkgs（`internal/service/`、`pkgs/`）** — 格式：`{Resource}{Action}`（去掉类名前缀）

| 操作 | 命名模式 | 示例 |
|---|---|---|
| 列表/单条 | `{Resource}List` / `{Resource}Inspect` | `RouteList()`、`ImageInspect(id)` |
| 创建/更新/删除 | `{Resource}Create/Update/Delete` | `RouteCreate(req)`、`RouteDelete(id)` |
| 状态切换 | `{Resource}StatusPatch` | `RouteStatusPatch(id, status)` |
| 特殊操作 | `{Resource}{Verb}` | `ContainerAction(id, action)`、`WhitelistRevoke(...)` |

- **模块前缀**：`docker`、`swarm`、`apisix`、`account`、`system`、`filer`、`compose`
- **资源名**：单数形式，不重复模块语义
- **禁止**：`动词+资源` 旧式命名（`CreateRoute`、`ListContainers`、`apisixCreateRoute`）
- **注意**：类名为 `Docker` 时 `ContainerList` 不缩写为 `List`；类名为 `Apisix` 时 `RouteList` 不缩写为 `List`

### 类型定义

- 请求/响应结构体：定义在对应 handler 子包；业务模型：就近定义在对应 `pkgs` 文件
- 避免跨包重复定义语义相同结构体

### 配置结构体

- 顶层 `Config`（`config/types.go`），子配置：`Server`、`AgentConfig`、`ApisixConfig`、`DockerConfig`、`MarketplaceConfig`、`MemberConfig`
- 镜像仓库 `DockerRegistry`（含 `Name`、`URL`、`Username`、`Password`、`Description`）
- 字段使用指针类型 `*` 和 YAML 标签

---

## 5) 前端编码规范（Vue/Tailwind）

适用目录：`webview/src`

### 5.1 组件装饰器

使用 `vue-facing-decorator`（`@Component`、`@Inject`、`@Prop`、`@Ref`、`@Watch`、`@Provide`），类必须 `toNative()` 包装后导出。`@Inject` 从 `APP_STATE_KEY`/`APP_ACTIONS_KEY` 注入

### 5.2 状态管理

- 全局状态 `store/state.ts` 使用 `reactive` + Provide/Inject
- 键：`APP_STATE_KEY`（`app.state`）、`APP_ACTIONS_KEY`（`app.actions`）
- 权限：`permissionsLoaded`（布尔）、`permissions`（`string[]`，格式为 `"METHOD /api/path"`），通过 `hasPerm(module)` 检查
- 初始化 `initProvider()` 返回 `{ state, actions }`

### 5.3 API 服务层

`service/api.ts` 单例 `class ApiService`，`export default new ApiService()`。请求统一通过 `http`/`httpBlob`（`service/axios.ts`），前者类型安全已解包为 `APIResponse`，后者 Blob 下载专用

### 5.4 类型定义与命名（强制）

`service/types/` 按域拆分（`docker`、`swarm`、`apisix`、`compose`、`system`、`account`、`overview`、`filer`），`service/types.ts` 统一 `export *` 导出

| 场景 | 命名 | 示例 |
|---|---|---|
| 列表/概览 | `XxxInfo` | `DockerContainerInfo`、`SwarmNodeInfo` |
| 详情/单条查询 | `XxxDetail` | `DockerImageDetail`、`SwarmNodeDetail` |
| 创建请求 | `XxxCreate` | `DockerContainerCreate`、`ApisixRouteCreate` |
| 更新请求 | `XxxUpdate` | `ApisixConsumerUpdate` |
| 复用创建类型 | `type XxxCreate = XxxSpec` | `ApisixUpstreamCreate = ApisixUpstream` |
| 响应结果 | `XxxResult` | `AuthLoginResult`、`ApiTokenResult` |
| 枚举联合 | `XxxType` / `XxxMode` | `ApisixUpstreamType`、`ComposeDeployTarget` |

禁止：`VO`/`BO`/`DTO` 等后缀

### 5.5 方法命名规范（强制）

`ApiService` 方法命名：`domainResourceAction` 驼峰格式

| 操作 | 命名模式 | 示例 |
|---|---|---|
| 列表 | `domainResourceList(params?)` | `dockerContainerList()`、`apisixRouteList()` |
| 单条 | `domainResource(id)` | `dockerImage(id)`、`swarmNode(id)` |
| 创建/更新/删除 | `domainResourceCreate/Update/Delete` | `dockerContainerCreate(data)`、`dockerImageDelete(id)` |
| 操作/动作 | `domainResourceAction(id, action)` | `dockerContainerAction(id, 'start')` |
| 状态切换 | `domainResourceStatus(id, status)` | `apisixRouteStatus(id, 0)` |
| 统计/日志 | `domainResourceStats/Logs(id)` | `dockerContainerStats(id)` |

- **域名前缀**：`docker`、`swarm`、`apisix`、`account`、`system`、`filer`、`compose`
- **资源名**：单数形式
- **分组注释**：`// ==================== XXXX 相关 ====================`

### 5.6 卡片与标题栏

- 列表/详情页统一 `.card mb-4`
- 标题栏：`bg-slate-50 border-b border-slate-200 rounded-t-2xl px-4 md:px-6 py-3`，容器不写 `flex/justify-between`
- 必须提供桌面 `hidden md:flex` 与移动 `flex md:hidden` 双布局
- 详情页右侧仅保留刷新等功能按钮，**不添加返回按钮**

### 5.7 列表双视图（强制）

- 桌面：`hidden md:block overflow-x-auto` + `<table>`
- 移动：`md:hidden space-y-3 p-4` + 卡片列表（**`p-4` 不得省略**，卡片 `rounded-xl border border-slate-200 bg-white p-4`）

### 5.8 表格第一列布局（强制）

- `<td>` 设 `max-w-[280px]`（纯短 ID 列除外）
- 副信息（描述/host/ID）显示在主名称下方，不占独立列
- 外层 flex 加 `min-w-0`，图标 `flex-shrink-0`，文字容器 `min-w-0` + `truncate block`
- 副信息样式：`text-xs text-slate-400 truncate block mt-0.5`；等宽内容加 `font-mono`；空值加 `v-if`

```html
<td class="px-4 py-3 max-w-[280px]">
  <div class="flex items-center gap-2 min-w-0">
    <div class="w-8 h-8 rounded-lg bg-xxx flex items-center justify-center flex-shrink-0">
      <i class="fas fa-xxx text-white text-sm"></i>
    </div>
    <div class="min-w-0">
      <span class="font-medium text-slate-800 truncate block">{{ 主名称 }}</span>
      <span class="text-xs text-slate-400 truncate block mt-0.5">{{ 副信息 }}</span>
    </div>
  </div>
</td>
```

图标配色（色阶 `400`）：容器 `emerald`/`slate`（按状态）、镜像 `blue`、网络 `purple`、数据卷 `amber`、仓库 `blue-500`、Swarm 服务 `emerald`、节点 `blue`、路由 `indigo`、白名单 `amber`、消费者 `violet`、用户 `blue-500`。新模块选未用色（`rose`/`cyan`/`lime` 等）

### 5.9 操作按钮语义色

统一格式：`btn-icon text-{color}-600 hover:bg-{color}-50`

| 操作 | 配色 | 图标 |
|---|---|---|
| 创建/启动/激活 | `emerald` | `fa-plus` / `fa-play` |
| 编辑/重启 | `blue` | `fa-pen` / `fa-rotate` |
| 停止/排空/告警 | `amber` | `fa-stop` / `fa-arrow-down` |
| 删除/移除 | `red` | `fa-trash` / `fa-xmark` |
| 详情/日志/暂停 | `slate` | `fa-circle-info` / `fa-file-lines` / `fa-pause` |
| 统计/扩缩容 | `indigo` | `fa-chart-line` / `fa-up-right-and-down-left-from-center` |
| 终端 | `teal` | `fa-terminal` |
| 禁用/只读 | `slate-300 cursor-not-allowed` | — |

> APISIX 域编辑按钮允许 `indigo`（路由）/`violet`（消费者），其余模块统一 `blue`

### 5.10 表单与敏感字段

- 表单容器 `max-w-3xl space-y-4`
- label：`block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1`，input 通用 `.input`，help `text-xs text-slate-400 mt-1`
- 密钥/密码：后端敏感字段 `json:"-"`，前端 `type="password" autocomplete="new-password"`，留空保存=不修改，placeholder："留空保持不变"

### 5.11 统一工具与轮询

通用函数复用 `webview/src/helper/utils.ts`；轮询间隔用 `POLL_INTERVAL`，禁止硬编码

### 5.12 import 分组排序

`<script>` 内 import 按以下顺序，组间空一行，组内字母升序：
1. 第三方库  2. `@/store/...`  3. `@/router`  4. `@/service/...`  5. `@/helper/...`  6. `@/component/...`  7. `@/views/...`  8. 其余 `@/`

同模块普通导入在前、`type` 导入在后紧邻。批量整理：`cd webview && python3 sort-imports.py src`（支持 `--dry-run`）

### 5.13 终端能力

系统终端走 `helper/shell.ts`，容器终端走 `helper/container-exec.ts`，禁止页面直接创建 Terminal/WebSocket 实例

---

## 6) 路由与导航

- `/overview` 概览；`docker/overview`/`swarm/overview` 仅作组件不作独立菜单路由
- 系统模块：`/system/config`、`/system/audit`；用户管理：`/account/members`
- 折叠子菜单展开状态跟随当前路由（`@Watch` immediate）
- 侧边栏宽度 `w-16`（折叠）→ `w-64`（展开）
- 移动端：遮罩 + 抽屉式（`-translate-x-full lg:translate-x-0`），`toggleMobileSidebar`/`closeMobileSidebar`/`openMobileSidebar`；窗口 ≥ 1024px 自动关闭

---

## 7) 注册中心与服务初始化

启动顺序：`main → config.Load → registry.Init → server.NewApp`

可用性检查：`IsDockerAvailable`、`IsSwarmAvailable`、`IsApisixAvailable`

已用命名：`registry.DockerService`、`registry.SwarmService`、`registry.ApisixClient`

---

## 8) 安全基线（必须遵守）

1. 禁止硬编码密钥/密码/令牌
2. 敏感配置仅返回 `xxxSet` 布尔值
3. 文件系统操作防目录遍历；解压防 Zip Slip
4. WebSocket 必须经过认证链路
5. 关键资源（内置角色等）前后端双重校验

---

## 9) 质量门禁（提交前自检）

```bash
go test ./...                                        # 后端编译/测试
cd webview && npm run lint                           # 前端类型检查
cd webview && python3 sort-imports.py --dry-run src  # import 排序检查
```

- [ ] 编译通过；[ ] 无新增 lint 警告；[ ] 关键路径手动验证；[ ] 错误处理与日志符合规范；[ ] 未引入明文敏感信息

---

## 10) Git 约定

提交格式：`<type>: <subject>`（`feat`/`fix`/`refactor`/`style`/`docs`/`chore`）

分支：`main`（生产）、`dev`（开发）、`feature/<name>`、`fix/<name>`

---

本规范适用于本仓库所有 AI 代理协作与代码改动。如与用户当次明确需求冲突，按"指令优先级"处理并在输出中说明取舍。
