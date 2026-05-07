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
- **同步更新文档与 Skills**：修改代码时，必须同步更新所有相关的说明文档和 skills 文件，确保文档与代码始终一致

---

## 3) 项目架构

### 分层依赖方向

`config → internal/registry → pkgs → internal/service → internal/server`

- `pkgs/`：原生客户端层，返回 SDK/底层类型
- `internal/service/`：业务组合、类型转换、参数校验
- `internal/server/`：HTTP 入口，解析请求 → 调用 service → 返回响应
- `internal/registry/`：外部服务初始化、生命周期管理与可用性检查

### 禁止

- `pkgs/` 依赖 `internal/`
- `handler` 中堆叠业务逻辑
- `service/handler` 直接从配置创建外部客户端

### 内聚与耦合

- 同领域功能聚合同包（`pkgs/docker/`、`pkgs/swarm/`），类型就近定义，不集中 `types.go`
- 层间通过接口或注入解耦；前端组件通过 `Provide/Inject` 获取全局状态
- 判定：改一个功能只需改一处；若需同时改多个包同名函数，说明内聚不足

---

## 4) 代码变更同步文档规范

修改代码时，必须同步更新所有相关的说明文档和 skills 文件，确保文档与代码始终一致。

### 适用范围

所有涉及以下变更的场景：

- 新增/修改/删除 API 路由或参数
- 新增/修改/删除 数据结构（Request/Response 字段）
- 新增/修改/删除 业务逻辑或工作流
- 新增/修改/删除 配置项或环境变量
- 新增/修改/删除 Shell 脚本中的命令或参数

### 需要同步更新的文件

| 代码变更位置 | 需同步更新的文档 |
|---|---|
| `internal/server/ctrl_*.go`（路由/参数） | `skills/isrvd/docs/*.md` 对应章节 |
| `pkgs/*/`（数据结构） | `skills/isrvd/docs/*.md` 中的字段表 |
| `internal/server/route.go`（路由注册） | `skills/isrvd/SKILL.md` API 速查表 |
| 部署/运维相关变更 | `skills/isrvd/scripts/*.sh` |
| 新功能/新模块 | `skills/isrvd/SKILL.md` 决策树 |

### 执行步骤

1. **识别影响范围**：修改代码后，明确哪些文档和 skills 文件会受影响
2. **同步更新文档**：
   - API 变更 → 更新 `docs/*.md` 中对应的路由、请求体、响应字段
   - 数据结构变更 → 更新 `docs/*.md` 中的字段表，确保与 Go struct 定义一致
   - 新增路由 → 在 `SKILL.md` API 速查表中添加条目
   - 脚本相关变更 → 更新 `scripts/*.sh` 中的对应命令
3. **验证一致性**：确认文档中的字段名、类型、路径与代码完全匹配

### 注意事项

- 文档中标注"只读"的字段（如 `create_time`、`update_time`、`id`）也必须列出
- 请求体和响应体中的字段必须与 Go struct 的 json tag 完全一致
- 路由路径必须与 `ctrl_*.go` 中的注册路径完全一致

---

## 5) 后端编码规范（Go）

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

## 6) 前端编码规范（Vue/Tailwind）

适用目录：`webview/src`

### 6.1 组件装饰器

使用 `vue-facing-decorator`（`@Component`、`@Inject`、`@Prop`、`@Ref`、`@Watch`、`@Provide`），类必须 `toNative()` 包装后导出。`@Inject` 从 `APP_STATE_KEY`/`APP_ACTIONS_KEY` 注入

### 6.2 状态管理

- 全局状态 `store/state.ts` 使用 `reactive` + Provide/Inject
- 键：`APP_STATE_KEY`（`app.state`）、`APP_ACTIONS_KEY`（`app.actions`）
- 权限：`permissionsLoaded`（布尔）、`permissions`（`string[]`，格式为 `"METHOD /api/path"`），通过 `hasPerm(module)` 检查
- 初始化 `initProvider()` 返回 `{ state, actions }`

### 6.3 API 服务层

`service/api.ts` 单例 `class ApiService`，`export default new ApiService()`。请求统一通过 `http`/`httpBlob`（`service/axios.ts`），前者类型安全已解包为 `APIResponse`，后者 Blob 下载专用

### 6.4 类型定义与命名（强制）

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

### 6.5 方法命名规范（强制）

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

### 6.6 卡片与标题栏

- 列表/详情页统一 `.card mb-4`
- 标题栏：`bg-slate-50 border-b border-slate-200 rounded-t-2xl px-4 md:px-6 py-3`，容器不写 `flex/justify-between`
- 必须提供桌面 `hidden md:flex` 与移动 `flex md:hidden` 双布局
- 详情页右侧仅保留刷新等功能按钮，**不添加返回按钮**

**toolbar 图标与标题（强制）**：

- 图标：`w-9 h-9 rounded-lg`（不得用 `w-10`/`w-8` 或 `rounded-xl`）
- 标题：`<h1 class="text-lg font-semibold text-slate-800 truncate">`（不得用 `h3`/`h2`/`text-base`）
- 副标题：`<p class="text-xs text-slate-500 truncate">`
- 移动端左侧容器：`flex items-center gap-3 min-w-0 flex-1`，图标加 `flex-shrink-0`，文字容器加 `min-w-0`
- 桌面端右侧按钮区：`flex items-center gap-2 flex-shrink-0`（刷新等功能按钮）

**图标样式（强制）**：

- 形状：`rounded-lg`（禁止 `rounded-full`/`rounded-2xl`）
- 尺寸：`w-16 h-16`（空状态/登录）、`w-12 h-12`（加载）、`w-10 h-10`（移动卡片）、`w-9 h-9`（toolbar）、`w-8 h-8`（桌面表格）

### 6.7 列表双视图（强制）

- 桌面：`hidden md:block overflow-x-auto` + `<table>`
- 移动：`md:hidden space-y-3 p-4` + 卡片列表（**`p-4` 不得省略**，卡片 `rounded-xl border border-slate-200 bg-white p-4 transition-all hover:shadow-sm`）

**移动端卡片顶部结构（强制）**：

```html
<div class="flex items-center gap-3 min-w-0 flex-1 mb-3">
  <div class="w-10 h-10 rounded-lg bg-xxx flex items-center justify-center flex-shrink-0">
    <i class="fas fa-xxx text-white text-base"></i>
  </div>
  <div class="min-w-0">
    <span class="font-medium text-slate-800 text-sm truncate block">{{ 主名称 }}</span>
    <span class="text-xs text-slate-400 truncate block mt-0.5">{{ 副信息 }}</span>
  </div>
</div>
```

- 卡片图标：`w-10 h-10 rounded-lg`
- 主名称：`font-medium text-slate-800 text-sm truncate block`
- 副信息：`text-xs text-slate-400 truncate block mt-0.5`

### 6.8 移动端卡片属性行对齐（强制）

卡片内每条属性行由 `<标签 span>` + `<值>` 组成，**行间距统一 `mb-3`**：

- 纯文本：`flex items-center gap-2 mb-3`，值用 `text-slate-500`
- badge/code：`flex items-start gap-2 mb-3`，标签加 `mt-0.5`
- badge 形状：`rounded` 或 `rounded-lg`（禁止 `rounded-full`）

```html
<!-- 纯文本 -->
<div class="flex items-center gap-2 mb-3">
  <span class="text-xs text-slate-400 flex-shrink-0">创建</span>
  <span class="text-xs text-slate-500">{{ value }}</span>
</div>

<!-- badge -->
<div class="flex items-start gap-2 mb-3">
  <span class="text-xs text-slate-400 flex-shrink-0 mt-0.5">状态</span>
  <span class="inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-medium ...">{{ value }}</span>
</div>

<!-- code -->
<div class="flex items-start gap-2 mb-3">
  <span class="text-xs text-slate-400 flex-shrink-0 mt-0.5">路径</span>
  <code class="text-xs bg-slate-100 px-2 py-0.5 rounded break-all">{{ value }}</code>
</div>
```

### 6.9 表格第一列布局（强制）

- `<td>` 设 `max-w-[280px]`（纯短 ID 列除外）
- 副信息（描述/host/ID）显示在主名称下方，不占独立列
- 外层 flex 加 `min-w-0`，图标 `flex-shrink-0`，文字容器 `min-w-0` + `truncate block`
- 副信息样式：`text-xs text-slate-400 truncate block mt-0.5`；等宽内容加 `font-mono`；空值加 `v-if`
- **主名称直接用 `<span class="font-medium text-slate-800 truncate block">`，不得用额外 flex 容器包裹**

**桌面端表格文字颜色规范**：

- 数据列：`text-sm text-slate-600`
- 副信息行：`text-xs text-slate-400`
- 操作按钮列：`flex justify-end items-center gap-1`
- badge 形状：`rounded` 或 `rounded-lg`（禁止 `rounded-full`）

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

### 6.9.1 状态文字颜色（强制）

**状态值优先用文字颜色区分，不用 badge**；仅枚举型分类字段（驱动、类型等）才用 badge。

| 状态值 | 颜色 | 示例 |
|---|---|---|
| 正常/运行/就绪（running / ready / active / enabled） | `text-emerald-600 font-medium` | 容器 running、节点 ready/active |
| 异常/停止/下线（stopped / down / error） | `text-red-500 font-medium` | 节点 down |
| 警告/排空/暂停（drain / paused / warning） | `text-amber-600 font-medium` | 节点 drain |
| 其他/未知 | `text-slate-500` | — |

- 通配符或空值（如 host `*`）用 `text-slate-400`（无 `font-medium`）
- 有意义的 host/域名用 `text-teal-600 font-medium`
- 数值型强调（如运行中任务数）用 `text-emerald-600 font-medium`

### 6.9.2 枚举 badge 配色

枚举型分类字段（驱动、协议、类型等）使用 badge，颜色跟随模块主色：

| 模块/字段 | badge 配色 |
|---|---|
| 网络驱动（docker network driver） | `bg-purple-50 text-purple-700` |
| 路由上游（apisix upstream） | `bg-indigo-50 text-indigo-700` |
| 其他枚举 | 跟随模块主色，`bg-{color}-50 text-{color}-700` |

### 6.9.3 权限/角色图标颜色

| 角色/权限 | 图标 | 颜色 |
|---|---|---|
| 创始人/超级管理员 | `fa-crown` | `text-violet-400` |
| 有自定义权限 | `fa-key` | `text-amber-400` |

### 6.10 操作按钮语义色

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

> APISIX 域各资源编辑按钮颜色：路由 `indigo`、消费者 `violet`、SSL `cyan`、上游 `emerald`、插件配置 `rose`；其余模块（docker/swarm/account/filer/compose）统一 `blue`

**移动端操作按钮区（强制）**：

- 容器：`flex flex-wrap gap-1.5 pt-2 border-t border-slate-100`（**`gap-1.5` 不得用 `gap-1`**）

### 6.11 表单与敏感字段

- 表单容器 `max-w-3xl space-y-4`
- label：`block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1`，input 通用 `.input`，help `text-xs text-slate-400 mt-1`
- 密钥/密码：后端敏感字段 `json:"-"`，前端 `type="password" autocomplete="new-password"`，留空保存=不修改，placeholder："留空保持不变"

### 6.12 统一工具与轮询

通用函数复用 `webview/src/helper/utils.ts`；轮询间隔用 `POLL_INTERVAL`，禁止硬编码

### 6.13 import 分组排序

`<script>` 内 import 按以下顺序，组间空一行，组内字母升序：

1. 第三方库  2. `@/store/...`  3. `@/router`  4. `@/service/...`  5. `@/helper/...`  6. `@/component/...`  7. `@/views/...`  8. 其余 `@/`

同模块普通导入在前、`type` 导入在后紧邻。批量整理：`cd webview && python3 sort-imports.py src`（支持 `--dry-run`）

### 6.14 终端能力

系统终端走 `helper/shell.ts`，容器终端走 `helper/container-exec.ts`，禁止页面直接创建 Terminal/WebSocket 实例

---

## 7) 路由与导航

- `/overview` 概览；`docker/overview`/`swarm/overview` 仅作组件不作独立菜单路由
- 系统模块：`/system/config`、`/system/audit`；用户管理：`/account/members`
- 折叠子菜单展开状态跟随当前路由（`@Watch` immediate）
- 侧边栏宽度 `w-16`（折叠）→ `w-64`（展开）
- 移动端：遮罩 + 抽屉式（`-translate-x-full lg:translate-x-0`），`toggleMobileSidebar`/`closeMobileSidebar`/`openMobileSidebar`；窗口 ≥ 1024px 自动关闭

---

## 8) 注册中心与服务初始化

启动顺序：`main → config.Load → registry.Init → server.NewApp`

可用性检查：`IsDockerAvailable`、`IsSwarmAvailable`、`IsApisixAvailable`

已用命名：`registry.DockerService`、`registry.SwarmService`、`registry.ApisixClient`

---

## 9) 安全基线（必须遵守）

1. 禁止硬编码密钥/密码/令牌
2. 敏感配置仅返回 `xxxSet` 布尔值
3. 文件系统操作防目录遍历；解压防 Zip Slip
4. WebSocket 必须经过认证链路
5. 关键资源（内置角色等）前后端双重校验

---

## 10) 质量门禁（提交前自检）

```bash
go test ./...                                          # 后端编译/测试
cd webview && npm run lint                             # 前端类型检查
cd webview && npm run format                           # import 排序检查
cd webview && python3 script/review-style.py           # 前端样式一致性检查
```

- [ ] 编译通过；[ ] 无新增 lint 警告；[ ] 关键路径手动验证；[ ] 错误处理与日志符合规范；[ ] 未引入明文敏感信息；[ ] 相关文档与 skills 已同步更新

---

## 11) Git 约定

提交格式：`<type>: <subject>`（`feat`/`fix`/`refactor`/`style`/`docs`/`chore`）

分支：`main`（生产）、`dev`（开发）、`feature/<name>`、`fix/<name>`

---

本规范适用于本仓库所有 AI 代理协作与代码改动。如与用户当次明确需求冲突，按"指令优先级"处理并在输出中说明取舍。
