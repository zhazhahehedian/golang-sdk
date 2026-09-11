package sdk

import (
	"encoding/json"
	"fmt"

	"github.com/fishpioffical/golang-sdk/types"
	"github.com/imroc/req/v3"
)

// apiResponse 保持 SDK 的 code/msg 返回约定，同时将 HTTP 和无效 JSON 作为错误返回。
func apiResponse[T any](request *req.Request, method, endpoint string) (*types.ApiResponse[T], error) {
	result := new(types.ApiResponse[T])
	resp, err := request.SetSuccessResult(result).SetErrorResult(result).Send(method, endpoint)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, fmt.Errorf("API request: HTTP %d", resp.StatusCode)
	}
	var envelope struct {
		Code *int `json:"code"`
	}
	if json.Unmarshal(resp.Bytes(), &envelope) != nil || envelope.Code == nil {
		return nil, fmt.Errorf("invalid API response (HTTP %d)", resp.StatusCode)
	}
	// 显式 JSON 解码，兼容未标注 Content-Type 的 JSON 响应。
	if err := json.Unmarshal(resp.Bytes(), result); err != nil {
		return nil, err
	}
	return result, nil
}
