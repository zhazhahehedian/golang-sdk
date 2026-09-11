package sdk

import (
	"fmt"
	"log/slog"

	"github.com/fishpioffical/golang-sdk/config"
	"github.com/imroc/req/v3"
)

// FishPiSDK SDK客户端
type FishPiSDK struct {
	configProvider config.Provider
	client         *req.Client
	oauthClient    *req.Client
	logDir         string
	logger         *slog.Logger
}

// NewSDK 创建新的SDK实例（使用ConfigProvider）
func NewSDK(configProvider config.Provider, options ...Option) *FishPiSDK {
	conf := configProvider.Get()

	if conf.BaseUrl == "" {
		conf.BaseUrl = BaseUrl
	}

	if conf.UserAgent == "" {
		conf.UserAgent = UserAgent
	}

	reqClient := req.NewClient().
		SetBaseURL(conf.BaseUrl).
		SetCommonHeader("User-Agent", conf.UserAgent)

	if conf.ApiKey != "" {
		reqClient.SetCommonQueryParam("apiKey", conf.ApiKey)
	}

	sdk := &FishPiSDK{
		configProvider: configProvider,
		client:         reqClient,
		// OAuth 凭据不经过 API Key 客户端的查询参数、Cookie 或请求转储日志。
		oauthClient: req.NewClient().SetRedirectPolicy(req.NoRedirectPolicy()),
	}

	// 应用选项
	for _, option := range options {
		option(sdk)
	}

	return sdk
}

// NewSDKWithAPIKey 使用API Key快速创建SDK实例（便捷方法）
func NewSDKWithAPIKey(apiKey string, domain ...string) *FishPiSDK {
	d := "fishpi.cn"
	if len(domain) > 0 && domain[0] != "" {
		d = domain[0]
	}

	conf := &config.Config{
		BaseUrl: fmt.Sprintf("https://%s", d),
		ApiKey:  apiKey,
	}

	provider := config.NewMemoryConfigProvider(conf)
	return NewSDK(provider)
}

// GetConfig 获取配置
func (s *FishPiSDK) GetConfig() *config.Config {
	return s.configProvider.Get()
}

// UpdateConfig 更新配置
func (s *FishPiSDK) UpdateConfig(config *config.Config) error {
	if err := s.configProvider.Update(config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// 更新 HTTP 客户端配置
	s.client.SetBaseURL(config.BaseUrl)

	if config.ApiKey != "" {
		s.client.SetCommonQueryParam("apiKey", config.ApiKey)
	}

	return nil
}

// GetUserAgent 获取User-Agent
func (s *FishPiSDK) GetUserAgent() string {
	return s.client.Headers.Get("User-Agent")
}

// GetAPIKey 获取当前API Key
func (s *FishPiSDK) GetAPIKey() string {
	return s.configProvider.Get().ApiKey
}
