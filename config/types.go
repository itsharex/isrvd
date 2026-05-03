package config

// 配置结构
type Config struct {
	Server      *Server            `yaml:"server"`
	Agent       *AgentConfig       `yaml:"agent"`
	Apisix      *ApisixConfig      `yaml:"apisix"`
	Docker      *DockerConfig      `yaml:"docker"`
	Marketplace *MarketplaceConfig `yaml:"marketplace"`
	Links       []*LinkConfig      `yaml:"links"`
	Members     []*MemberConfig    `yaml:"members"`
}

// 服务器配置
type Server struct {
	Debug           bool   `yaml:"debug" json:"debug"`
	ListenAddr      string `yaml:"listenAddr" json:"listenAddr"`
	JWTSecret       string `yaml:"jwtSecret" json:"-"` // 敏感字段不序列化到 JSON
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
	// Permissions 各模块权限，key 为模块名，value 为 "r"（只读）或 "rw"（读写），空字符串或缺失表示无权限
	// 可用模块：overview, system, account, shell, filer, agent, apisix, docker, swarm, compose
	Permissions map[string]string `yaml:"permissions,omitempty" json:"permissions,omitempty"`
}
