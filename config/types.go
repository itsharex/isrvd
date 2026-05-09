package config

// 配置结构
type Config struct {
	Server      *ServerConfig      `yaml:"server"`
	Agent       *AgentConfig       `yaml:"agent"`
	Apisix      *ApisixConfig      `yaml:"apisix"`
	Docker      *DockerConfig      `yaml:"docker"`
	Marketplace *MarketplaceConfig `yaml:"marketplace"`
	Links       []*LinkConfig      `yaml:"links"`
	Members     []*MemberConfig    `yaml:"members"`
}

// 服务器配置
type ServerConfig struct {
	Debug           bool   `yaml:"debug" json:"debug"`
	ListenAddr      string `yaml:"listenAddr" json:"listenAddr"`
	JWTSecret       string `yaml:"jwtSecret" json:"-"`                 // 敏感字段不序列化到 JSON
	JWTExpiration   int64  `yaml:"jwtExpiration" json:"jwtExpiration"` // JWT 过期时间（秒），默认 86400
	MaxUploadSize   int64  `yaml:"maxUploadSize" json:"maxUploadSize"` // 文件上传最大大小（字节），默认 100MB
	ProxyHeaderName string `yaml:"proxyHeaderName" json:"proxyHeaderName"`
	RootDirectory   string `yaml:"rootDirectory" json:"rootDirectory"`
}

// Agent LLM 配置
type AgentConfig struct {
	Model   string `yaml:"model" json:"model"`     // 模型名称
	BaseURL string `yaml:"baseUrl" json:"baseUrl"` // LLM API 基础地址（OpenAI 兼容）
	APIKey  string `yaml:"apiKey" json:"-"`        // API 密钥（敏感字段不序列化到 JSON）
}

// Apisix 配置
type ApisixConfig struct {
	AdminURL string `yaml:"adminUrl" json:"adminUrl"` // Apisix Admin API 地址
	AdminKey string `yaml:"adminKey" json:"-"`        // Apisix Admin API Key（敏感字段不序列化到 JSON）
}

// Docker 配置
type DockerConfig struct {
	Host          string            `yaml:"host" json:"host"`                       // Docker 连接地址
	ContainerRoot string            `yaml:"containerRoot" json:"containerRoot"`     // 容器数据根目录
	Registries    []*DockerRegistry `yaml:"registries" json:"registries,omitempty"` // 镜像仓库配置列表
}

// 镜像仓库配置
type DockerRegistry struct {
	Name        string `yaml:"name" json:"name"`               // 仓库名称（用于显示）
	Description string `yaml:"description" json:"description"` // 仓库描述（可选）
	URL         string `yaml:"url" json:"url"`                 // 仓库地址，如 registry.example.com
	Username    string `yaml:"username" json:"username"`       // 用户名（可选）
	Password    string `yaml:"password" json:"-"`              // 密码（敏感字段不序列化到 JSON）
}

// 应用市场配置
type MarketplaceConfig struct {
	URL string `yaml:"url" json:"url"` // 应用市场站点地址，通过 iframe 嵌入
}

// 工具栏链接配置
type LinkConfig struct {
	Label string `yaml:"label" json:"label"` // 显示名称
	URL   string `yaml:"url" json:"url"`     // 链接地址
	Icon  string `yaml:"icon" json:"icon"`   // Font Awesome 图标类名（可选，如 fa-link）
}

// 成员配置
type MemberConfig struct {
	Username      string `yaml:"username" json:"username"`
	Password      string `yaml:"password" json:"-"` // 敏感字段不序列化到 JSON
	HomeDirectory string `yaml:"homeDirectory" json:"homeDirectory"`
	// Founder 创始人标志，创始人拥有所有模块的完整权限
	Founder bool `yaml:"founder" json:"founder"`
	// Description 成员描述信息（可选）
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Permissions 允许访问的路由列表，格式为 "METHOD /api/path"，如 "GET /api/docker/containers"
	Permissions []string `yaml:"permissions,omitempty" json:"permissions,omitempty"`
}
