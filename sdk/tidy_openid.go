package sdk

import (
	"net/url"

	"github.com/fishpioffical/golang-sdk/types"
)

// GetOpenIdUrl 获取登陆地址
func (s *FishPiSDK) GetOpenIdUrl(realm string, returnTo string) string {

	query := url.Values{}
	query.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	query.Set("openid.mode", "checkid_setup")
	query.Set("openid.return_to", returnTo)
	query.Set("openid.realm", realm)
	query.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	query.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")

	addr := url.URL{
		Path:     "/openid/login",
		RawQuery: query.Encode(),
	}

	return s.GetConfig().BaseUrl + addr.String()
}

// PostOpenIdVerify 验证用户信息
func (s *FishPiSDK) PostOpenIdVerify(query map[string]string) (*string, error) {
	result, err := s.PostOpenIdVerifyWithTokens(query)
	if err != nil {
		return nil, err
	}
	return &result.OpenID, nil
}

// GetUserInfoById 使用用户ID获取用户基础信息
func (s *FishPiSDK) GetUserInfoById(userId string) (*types.ApiResponse[*types.GetUserInfoByIdData], error) {
	response := new(types.ApiResponse[*types.GetUserInfoByIdData])

	_, err := s.client.R().
		SetQueryParam("userId", userId).
		SetSuccessResult(response).
		SetErrorResult(response).
		Get("/api/user/getInfoById")

	if err != nil {
		return nil, err
	}

	return response, nil
}
