package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Database       DatabaseConfig       `yaml:"database"`
	JWT            JWTConfig            `yaml:"jwt"`
	Domain         DomainConfig         `yaml:"domain"`
	TencentCloud   TencentCloudConfig   `yaml:"tencent_cloud"`
	Email          EmailConfig          `yaml:"email"`
	WechatMP       WechatMPConfig       `yaml:"wechat_miniprogram"`
	AlipayMP       AlipayMPConfig       `yaml:"alipay_miniprogram"`
	XhsMP          XhsMPConfig          `yaml:"xhs_miniprogram"`
	WechatPay      WechatPayConfig      `yaml:"wechat_pay"`
	AlipayPay      AlipayPayConfig      `yaml:"alipay_pay"`
	CORS           CORSConfig           `yaml:"cors"`
	Log            LogConfig            `yaml:"log"`
	ProductCategories []ProductCategory `yaml:"product_categories"`
}

type ServerConfig struct {
	HTTPPort     int    `yaml:"http_port"`
	GRPCPort     int    `yaml:"grpc_port"`
	Mode         string `yaml:"mode"`
	FrontendPort int    `yaml:"frontend_port"`
	BackendPort  int    `yaml:"backend_port"`
}

type DatabaseConfig struct {
	Driver     string      `yaml:"driver"`
	SQLitePath string      `yaml:"sqlite_path"`
	MySQL      MySQLConfig `yaml:"mysql"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type DomainConfig struct {
	Web string `yaml:"web"`
	API string `yaml:"api"`
	H5  string `yaml:"h5"`
	IP  string `yaml:"ip"`
}

type TencentCloudConfig struct {
	COS COSConfig `yaml:"cos"`
	SMS SMSConfig `yaml:"sms"`
}

type COSConfig struct {
	Enabled   bool   `yaml:"enabled"`
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	BaseURL   string `yaml:"base_url"`
}

type SMSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	AppID      string `yaml:"app_id"`
	SignName   string `yaml:"sign_name"`
	TemplateID string `yaml:"template_id"`
}

type EmailConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"from_name"`
}

type WechatMPConfig struct {
	AppID     string          `yaml:"app_id"`
	AppSecret string          `yaml:"app_secret"`
	Pages     WechatMPPages   `yaml:"pages"`
}

type WechatMPPages struct {
	Home    string `yaml:"home"`
	Product string `yaml:"product"`
	Cart    string `yaml:"cart"`
	Order   string `yaml:"order"`
	User    string `yaml:"user"`
}

type AlipayMPConfig struct {
	AppID          string         `yaml:"app_id"`
	PrivateKey     string         `yaml:"private_key"`
	AlipayPublicKey string        `yaml:"alipay_public_key"`
	Pages          AlipayMPPages  `yaml:"pages"`
}

type AlipayMPPages struct {
	Home    string `yaml:"home"`
	Product string `yaml:"product"`
	Cart    string `yaml:"cart"`
	Order   string `yaml:"order"`
	User    string `yaml:"user"`
}

type XhsMPConfig struct {
	AppID     string     `yaml:"app_id"`
	AppSecret string     `yaml:"app_secret"`
	Pages     XhsMPPages `yaml:"pages"`
}

type XhsMPPages struct {
	Home    string `yaml:"home"`
	Product string `yaml:"product"`
	Cart    string `yaml:"cart"`
	Order   string `yaml:"order"`
	User    string `yaml:"user"`
}

type WechatPayConfig struct {
	Enabled   bool   `yaml:"enabled"`
	MchID     string `yaml:"mch_id"`
	APIKey    string `yaml:"api_key"`
	CertPath  string `yaml:"cert_path"`
	KeyPath   string `yaml:"key_path"`
	NotifyURL string `yaml:"notify_url"`
}

type AlipayPayConfig struct {
	Enabled             bool   `yaml:"enabled"`
	AppID               string `yaml:"app_id"`
	PrivateKeyPath      string `yaml:"private_key_path"`
	AlipayPublicKeyPath string `yaml:"alipay_public_key_path"`
	NotifyURL           string `yaml:"notify_url"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

type LogConfig struct {
	Level    string `yaml:"level"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

type ProductCategory struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
}

var GlobalConfig *Config

func LoadConfig() error {
	// 从根目录加载配置文件
	configPath := filepath.Join("..", "config.yaml")
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	// 环境变量覆盖
	applyEnvOverrides(config)

	GlobalConfig = config
	return nil
}

func applyEnvOverrides(config *Config) {
	if port := os.Getenv("HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.HTTPPort = p
		}
	}
	if port := os.Getenv("BACKEND_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.BackendPort = p
		}
	}
	if driver := os.Getenv("DB_DRIVER"); driver != "" {
		config.Database.Driver = driver
	}
	if path := os.Getenv("SQLITE_PATH"); path != "" {
		config.Database.SQLitePath = path
	}
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		config.Database.MySQL.Host = host
	}
	if port := os.Getenv("MYSQL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.MySQL.Port = p
		}
	}
	if user := os.Getenv("MYSQL_USER"); user != "" {
		config.Database.MySQL.User = user
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		config.Database.MySQL.Password = password
	}
	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		config.Database.MySQL.Database = database
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}
	if secretID := os.Getenv("COS_SECRET_ID"); secretID != "" {
		config.TencentCloud.COS.SecretID = secretID
	}
	if secretKey := os.Getenv("COS_SECRET_KEY"); secretKey != "" {
		config.TencentCloud.COS.SecretKey = secretKey
	}
	if bucket := os.Getenv("COS_BUCKET"); bucket != "" {
		config.TencentCloud.COS.Bucket = bucket
	}
	if region := os.Getenv("COS_REGION"); region != "" {
		config.TencentCloud.COS.Region = region
	}
	if config.TencentCloud.COS.Bucket != "" && config.TencentCloud.COS.Region != "" {
		config.TencentCloud.COS.BaseURL = "https://" + config.TencentCloud.COS.Bucket + ".cos." + config.TencentCloud.COS.Region + ".myqcloud.com"
	}
	if config.TencentCloud.COS.SecretID != "" && config.TencentCloud.COS.SecretKey != "" && config.TencentCloud.COS.Bucket != "" {
		config.TencentCloud.COS.Enabled = true
	}
	if username := os.Getenv("EMAIL_USERNAME"); username != "" {
		config.Email.Username = username
	}
	if password := os.Getenv("EMAIL_PASSWORD"); password != "" {
		config.Email.Password = password
	}
	if appID := os.Getenv("WECHAT_MP_APPID"); appID != "" {
		config.WechatMP.AppID = appID
	}
	if appSecret := os.Getenv("WECHAT_MP_SECRET"); appSecret != "" {
		config.WechatMP.AppSecret = appSecret
	}
	if appID := os.Getenv("ALIPAY_MP_APPID"); appID != "" {
		config.AlipayMP.AppID = appID
	}
	if appID := os.Getenv("XHS_MP_APPID"); appID != "" {
		config.XhsMP.AppID = appID
	}
	if appSecret := os.Getenv("XHS_MP_SECRET"); appSecret != "" {
		config.XhsMP.AppSecret = appSecret
	}

	// 微信支付环境变量覆盖
	if mchID := os.Getenv("WECHAT_PAY_MCH_ID"); mchID != "" {
		config.WechatPay.MchID = mchID
	}
	if apiKey := os.Getenv("WECHAT_PAY_API_KEY"); apiKey != "" {
		config.WechatPay.APIKey = apiKey
	}
	if config.WechatPay.MchID != "" && config.WechatPay.APIKey != "" {
		config.WechatPay.Enabled = true
	}

	// 支付宝支付环境变量覆盖
	if appID := os.Getenv("ALIPAY_PAY_APPID"); appID != "" {
		config.AlipayPay.AppID = appID
	}
	if config.AlipayPay.AppID != "" {
		config.AlipayPay.Enabled = true
	}
}

func GetConfig() *Config {
	return GlobalConfig
}
