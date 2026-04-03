package models

// ServerConfig 服务器配置参数
type SystemConfig struct {
	SystemLogo      string `json:"systemLogo" yaml:"SystemLogo"`
	SystemIcon      string `json:"systemIcon" yaml:"SystemIcon"`
	SystemName      string `json:"systemName" yaml:"SystemName"`
	SystemCopyright string `json:"systemCopyright" yaml:"SystemCopyright"`
	SystemRecordNo  string `json:"systemRecordNo" yaml:"SystemRecordNo"`
	DefaultUsername string `json:"defaultUsername" yaml:"DefaultUsername"`
	DefaultPassword string `json:"defaultPassword" yaml:"DefaultPassword"`
}

type SafeConfig struct {
	LoginLockThreshold int  `json:"loginLockThreshold" yaml:"LoginLockThreshold"`
	LoginLockExpire    int  `json:"loginLockExpire" yaml:"LoginLockExpire"`
	LoginLockDuration  int  `json:"loginLockDuration" yaml:"LoginLockDuration"`
	MinPasswordLength  int  `json:"minPasswordLength" yaml:"MinPasswordLength"`
	RequireSpecialChar bool `json:"requireSpecialChar" yaml:"RequireSpecialChar"`
}

// CaptchaConfig 验证码配置参数
type CaptchaConfig struct {
	Open   bool `json:"open" yaml:"open"`
	Length int  `json:"length" yaml:"length"`
}

// ExportConfig 导出配置参数
type PlatformConfig struct {
	ExportAsyncThreshold      int  `json:"exportAsyncThreshold" yaml:"ExportAsyncThreshold"`
	ExportCleanDays           int  `json:"exportCleanDays" yaml:"ExportCleanDays"`
	CustomerExportWorkerCount int  `json:"customerExportWorkerCount" yaml:"CustomerExportWorkerCount"`
	Watermark                 bool `json:"watermark" yaml:"Watermark"`
}

// ConfigRequest 配置请求参数
type ConfigRequest struct {
	System        SystemConfig           `json:"system" yaml:"System"`
	Safe          SafeConfig             `json:"safe" yaml:"Safe"`
	Captcha       CaptchaConfig          `json:"captcha" yaml:"Captcha"`
	Platform      PlatformConfig         `json:"platform" yaml:"Platform"`
	Export        PlatformConfig         `json:"export,omitempty" yaml:"Export,omitempty"`
	CustomerExtra map[string]interface{} `json:"customerExtra,omitempty" yaml:"CustomerExtra,omitempty"`
}
