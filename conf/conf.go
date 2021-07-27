package conf

type PcsResponseBase struct {
	ErrorCode int  	 `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	RequestID uint64 `json:"request_id"`
}

type CloudDiskResponseBase struct {
	ErrorCode int  	 `json:"errno"`
	ErrorMsg  string `json:"errmsg"`
	RequestID uint64 `json:"request_id"`
}

type TestDataConfig struct {
	ClientID string
	ClientSecret string
	RedirectUri string
	Code string
	AccessToken string
	RefreshToken string
	Dir string
	FsID uint64
	Path string
	LocalFilePath string
	TranscodingType string
}

const (
	BaiduOpenApiDomain = "https://openapi.baidu.com"
	OpenApiDomain = "https://pan.baidu.com"
	PcsDataDomain = "https://d.pcs.baidu.com"
)

// 测试参数
var TestData TestDataConfig
