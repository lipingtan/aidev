package config

var ExtConfig Extend

// Extend 扩展配置
//  extend:
//    demo:
//      name: demo-name
// 使用方法： config.ExtConfig......即可！！
type Extend struct {
	AMap AMap // 这里配置对应配置文件的结构即可
}

type AMap struct {
	Key string
}

// WeChatPayConfig 微信支付配置（从 settings.yml 加载）
type WeChatPayConfig struct {
	AppID            string `yaml:"app_id"`
	MchID            string `yaml:"mch_id"`
	APIKey           string `yaml:"api_key"`
	CertSerialNo     string `yaml:"cert_serial_no"`
	PrivateKeyPath   string `yaml:"private_key_path"`
	PubKeyPath       string `yaml:"pubkey_path"`
	PlatformCertPath string `yaml:"platform_cert_path"`
	NotifyURL        string `yaml:"notify_url"`
	RefundNotifyURL  string `yaml:"refund_notify_url"`
}
