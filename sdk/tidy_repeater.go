package sdk

import (
	"errors"
	"strings"

	"github.com/fishpioffical/golang-sdk/types"
)

// GetRepeaterItems 获取转录列表，contentType 可为 joke/kfc/fish，空字符串查询全部。
func (s *FishPiSDK) GetRepeaterItems(contentType string) (*types.ApiResponse[*types.RepeaterItemsData], error) {
	r := s.client.R()
	if contentType != "" {
		r.SetQueryParam("repeaterContentType", contentType)
	}
	return apiResponse[*types.RepeaterItemsData](r, "GET", "/api/repeater/items")
}

// GetRepeaterNext 随机获取一条转录；excludeID 为空时不排除指定内容。
func (s *FishPiSDK) GetRepeaterNext(contentType, excludeID string) (*types.ApiResponse[*types.RepeaterItemData], error) {
	r := s.client.R()
	if contentType != "" {
		r.SetQueryParam("repeaterContentType", contentType)
	}
	if excludeID != "" {
		r.SetQueryParam("excludeId", excludeID)
	}
	return apiResponse[*types.RepeaterItemData](r, "GET", "/api/repeater/next")
}

// PostRepeater 上传转录内容，服务端验证类型及 2 到 500 字符限制。
func (s *FishPiSDK) PostRepeater(contentType, content string) (*types.ApiResponse[*types.RepeaterItemData], error) {
	r := s.client.R().SetBodyJsonMarshal(map[string]string{"apiKey": s.GetAPIKey(), "repeaterContentType": contentType, "repeaterContent": content})
	return apiResponse[*types.RepeaterItemData](r, "POST", "/api/repeater")
}

// PostRepeaterLike 切换点赞状态：已点赞时再次调用会取消点赞。
func (s *FishPiSDK) PostRepeaterLike(id string) (*types.ApiResponse[*types.RepeaterLikeData], error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("repeater id is required")
	}
	r := s.client.R().SetPathParam("id", id).SetBodyJsonMarshal(map[string]string{"apiKey": s.GetAPIKey()})
	return apiResponse[*types.RepeaterLikeData](r, "POST", "/api/repeater/{id}/like")
}
