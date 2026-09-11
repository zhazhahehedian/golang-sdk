package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/fishpioffical/golang-sdk/types"
	"github.com/imroc/req/v3"
)

// GetOpenIdUrlWithScopes 生成带授权范围的登录链接，保留原 GetOpenIdUrl 签名。
func (s *FishPiSDK) GetOpenIdUrlWithScopes(realm, returnTo string, scopes ...string) string {
	addr := s.GetOpenIdUrl(realm, returnTo)
	if len(scopes) == 0 {
		return addr
	}
	return addr + "&fishpi.scope=" + url.QueryEscape(strings.Join(scopes, " "))
}

func (s *FishPiSDK) oauthRequest() *req.Request {
	return s.oauthClient.R().SetHeader("User-Agent", s.GetUserAgent())
}

func (s *FishPiSDK) oauthURL(endpoint string) string {
	return strings.TrimRight(s.GetConfig().BaseUrl, "/") + endpoint
}

// PostOpenIdVerifyWithTokens 校验回跳参数并读取 OAuth 令牌；不会修改传入 map。
// 调用者仍需校验回调与本地登录会话、return_to 和 nonce 的绑定。
func (s *FishPiSDK) PostOpenIdVerifyWithTokens(query map[string]string) (*types.OpenIDVerifyResult, error) {
	params := make(map[string]string, len(query))
	for key, value := range query {
		if strings.HasPrefix(key, "openid.") {
			params[key] = value
		}
	}
	params["openid.mode"] = "check_authentication"
	resp, err := s.oauthRequest().SetBodyJsonMarshal(params).Post(s.oauthURL("/openid/verify"))
	if err != nil {
		return nil, errors.New("openid verification request failed")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openid verification: HTTP %d", resp.StatusCode)
	}
	fields := make(map[string]string)
	for _, line := range strings.Split(resp.String(), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if _, exists := fields[key]; exists {
			return nil, errors.New("duplicate OpenID response field")
		}
		fields[key] = strings.TrimSpace(value)
	}
	if fields["is_valid"] != "true" {
		return nil, errors.New("用户信息验证失败")
	}
	identity, err := url.Parse(params["openid.identity"])
	if err != nil || identity.Host == "" || (identity.Scheme != "https" && identity.Scheme != "http") || identity.Path == "" || strings.HasSuffix(identity.Path, "/") {
		return nil, errors.New("invalid OpenID identity")
	}
	result := &types.OpenIDVerifyResult{OpenID: path.Base(identity.Path)}
	result.Scope = fields["scope"]
	result.TokenType = fields["token_type"]
	result.AccessToken = fields["access_token"]
	result.RefreshToken = fields["refresh_token"]
	for key, dest := range map[string]*int64{"expires_in": &result.ExpiresIn, "refresh_expires_in": &result.RefreshExpiresIn} {
		if value, ok := fields[key]; ok {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid OpenID response field: %s", key)
			}
			*dest = n
		}
	}
	return result, nil
}

// oauthResponse 不使用可记录原始数据的自定义反序列化器，避免令牌进入日志。
func oauthResponse[T any](request *req.Request, method, endpoint string) (*types.ApiResponse[T], error) {
	resp, err := request.Send(method, endpoint)
	if err != nil {
		return nil, errors.New("OAuth request failed")
	}
	var result types.ApiResponse[T]
	var envelope struct {
		Code *int `json:"code"`
	}
	body := resp.Bytes()
	if json.Unmarshal(body, &envelope) != nil || envelope.Code == nil || json.Unmarshal(body, &result) != nil {
		return nil, fmt.Errorf("invalid OAuth response (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &result, fmt.Errorf("OAuth request: HTTP %d", resp.StatusCode)
	}
	return &result, nil
}

// PostOpenIdToken 手动续签；不自动重试或保存令牌，避免重复使用已轮换的令牌。
func (s *FishPiSDK) PostOpenIdToken(refreshToken string) (*types.ApiResponse[*types.OAuthTokens], error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, errors.New("refresh token is required")
	}
	r := s.oauthRequest().SetBodyJsonMarshal(map[string]string{"grant_type": "refresh_token", "refresh_token": refreshToken})
	return oauthResponse[*types.OAuthTokens](r, "POST", s.oauthURL("/openid/token"))
}

func oauthGet[T any](s *FishPiSDK, token, endpoint string, page, size int) (*types.ApiResponse[T], error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("access token is required")
	}
	if page < 0 || size < 0 {
		return nil, errors.New("page and size must not be negative")
	}
	r := s.oauthRequest().SetHeader("Authorization", "Bearer "+token)
	if page > 0 {
		r.SetQueryParam("p", strconv.Itoa(page))
	}
	if size > 0 {
		r.SetQueryParam("size", strconv.Itoa(size))
	}
	return oauthResponse[T](r, "GET", s.oauthURL(endpoint))
}

// GetOAuthProfile 查询基础资料，需要 profile.read。
func (s *FishPiSDK) GetOAuthProfile(token string) (*types.ApiResponse[*types.OAuthProfile], error) {
	return oauthGet[*types.OAuthProfile](s, token, "/openid/user/profile", 0, 0)
}

// GetOAuthUserDetail 查询详细资料，需要 profile.detail.read。
func (s *FishPiSDK) GetOAuthUserDetail(token string) (*types.ApiResponse[*types.OAuthUserDetail], error) {
	return oauthGet[*types.OAuthUserDetail](s, token, "/openid/user/detail", 0, 0)
}

// GetOAuthMembership 查询 VIP 状态，需要 membership.read。
func (s *FishPiSDK) GetOAuthMembership(token string) (*types.ApiResponse[*types.OAuthMembership], error) {
	return oauthGet[*types.OAuthMembership](s, token, "/openid/user/membership", 0, 0)
}

// GetOAuthPoints 查询积分记录，需要 points.read。page/size 为 0 使用服务端默认值。
func (s *FishPiSDK) GetOAuthPoints(token string, page, size int) (*types.ApiResponse[*types.OAuthPoints], error) {
	if size > 200 {
		return nil, errors.New("points page size must not exceed 200")
	}
	return oauthGet[*types.OAuthPoints](s, token, "/openid/user/points", page, size)
}

// GetOAuthArticles 查询公开发帖记录，需要 articles.read。page/size 为 0 使用服务端默认值。
func (s *FishPiSDK) GetOAuthArticles(token string, page, size int) (*types.ApiResponse[*types.OAuthArticles], error) {
	if size > 100 {
		return nil, errors.New("articles page size must not exceed 100")
	}
	return oauthGet[*types.OAuthArticles](s, token, "/openid/user/articles", page, size)
}
