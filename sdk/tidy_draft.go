package sdk

import (
	"errors"
	"strings"

	"github.com/fishpioffical/golang-sdk/types"
)

// GetArticleDrafts 查询当前用户的草稿列表，服务端最多返回 50 条。
func (s *FishPiSDK) GetArticleDrafts() (*types.ApiResponse[*types.ArticleDraftListData], error) {
	return apiResponse[*types.ArticleDraftListData](s.client.R(), "GET", "/api/article-drafts")
}

// PostArticleDraft 保存完整草稿；ArticleDraftID 为空时新建，否则更新。
func (s *FishPiSDK) PostArticleDraft(body *types.PostArticleDraftRequest) (*types.ApiResponse[*types.ArticleDraftData], error) {
	if body == nil {
		return nil, errors.New("draft request is required")
	}
	payload := struct {
		*types.PostArticleDraftRequest
		APIKey string `json:"apiKey"`
	}{body, s.GetAPIKey()}
	return apiResponse[*types.ArticleDraftData](s.client.R().SetBodyJsonMarshal(payload), "POST", "/api/article-drafts")
}

// GetArticleDraft 查询草稿详情（包含正文）。
func (s *FishPiSDK) GetArticleDraft(id string) (*types.ApiResponse[*types.ArticleDraftData], error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("draft id is required")
	}
	return apiResponse[*types.ArticleDraftData](s.client.R().SetPathParam("id", id), "GET", "/api/article-drafts/{id}")
}

// DeleteArticleDraft 删除当前用户的指定草稿。
func (s *FishPiSDK) DeleteArticleDraft(id string) (*types.ApiResponse[*types.DeleteArticleDraftData], error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("draft id is required")
	}
	return apiResponse[*types.DeleteArticleDraftData](s.client.R().SetPathParam("id", id), "DELETE", "/api/article-drafts/{id}")
}
