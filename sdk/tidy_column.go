package sdk

import (
	"errors"
	"strings"

	"github.com/fishpioffical/golang-sdk/types"
)

// PostColumnCover 更新自己的专栏封面。coverURL 为空时清空封面。
func (s *FishPiSDK) PostColumnCover(columnID, coverURL string) (*types.ApiResponse[*types.ColumnCoverData], error) {
	if strings.TrimSpace(columnID) == "" {
		return nil, errors.New("column id is required")
	}
	r := s.client.R().SetPathParam("id", columnID).SetBodyJsonMarshal(map[string]string{"apiKey": s.GetAPIKey(), "columnCoverURL": coverURL})
	return apiResponse[*types.ColumnCoverData](r, "POST", "/api/columns/{id}/cover")
}
