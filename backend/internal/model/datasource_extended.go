package model

// MongoDBDatasourceConfig MongoDB数据源配置
type MongoDBDatasourceConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Database     string `json:"database"`
	AuthDatabase string `json:"authDatabase"` // 认证数据库
	SSL          bool   `json:"ssl"`
}

// ElasticsearchDatasourceConfig Elasticsearch数据源配置
type ElasticsearchDatasourceConfig struct {
	Hosts    []string `json:"hosts"`    // 集群地址
	Username string   `json:"username"`
	Password string   `json:"password"`
	SSL      bool     `json:"ssl"`
	APIKey   string   `json:"apiKey"` // API Key认证
}

// APIDatasourceConfig API数据源配置
type APIDatasourceConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"` // GET, POST
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	AuthType    string            `json:"authType"`    // none, basic, bearer
	AuthToken   string            `json:"authToken"`
	ResponseType string            `json:"responseType"` // json, xml
}

// 扩展数据源类型常量
const (
	DatasourceTypeMongoDB        = "mongodb"
	DatasourceTypeElasticsearch  = "elasticsearch"
	DatasourceTypeAPI            = "api"
)
