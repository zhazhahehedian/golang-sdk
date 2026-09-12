package credit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultBaseURL = "https://credit.yuis.cc"
const maxResponseBytes = 2 << 20

// Config 仅用于积分服务，不与社区 SDK 的配置或凭据共享。
type Config struct {
	AppID     string
	AppSecret string
	BaseURL   string // 默认使用当前文档站点；兼容自行部署或平台迁移。
	UserAgent string
}

// Client 创建后可并发调用。不要在使用期间修改传入 HTTP 客户端的 Transport/Jar。
type Client struct {
	config       Config
	merchantHTTP *http.Client
	cashierHTTP  *http.Client
}

type Option func(*clientOptions)
type clientOptions struct{ httpClient *http.Client }

// WithHTTPClient 使用指定超时、Transport 及收银台 CookieJar。
// SDK 复制 http.Client，不修改原对象；商户请求移除 Jar，所有请求禁用重定向。
// Jar 应仅包含调用者已合法获取的积分服务登录会话。不要在 Transport 中记录密钥。
func WithHTTPClient(client *http.Client) Option {
	return func(options *clientOptions) { options.httpClient = client }
}

// NewClient 创建使用独立应用凭据的积分客户端。
func NewClient(config Config, options ...Option) (*Client, error) {
	if strings.TrimSpace(config.AppID) == "" || strings.TrimSpace(config.AppSecret) == "" {
		return nil, errors.New("credit: app ID and app secret are required")
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	u, err := url.Parse(config.BaseURL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, errors.New("credit: invalid base URL")
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	if config.UserAgent == "" {
		config.UserAgent = "FishPi-Go-SDK/credit"
	}
	opts := clientOptions{httpClient: &http.Client{Timeout: 30 * time.Second}}
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	if opts.httpClient == nil {
		return nil, errors.New("credit: HTTP client is required")
	}
	merchant, cashier := *opts.httpClient, *opts.httpClient
	merchant.Jar = nil
	merchant.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	cashier.CheckRedirect = merchant.CheckRedirect
	return &Client{config: config, merchantHTTP: &merchant, cashierHTTP: &cashier}, nil
}

// APIError 表示非 2xx 或业务 code 非 0；平台 message 保留在 Response 中。
// Error 不包含响应正文、请求体或密钥。
type APIError struct {
	HTTPStatusCode int
	Code           int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("credit: API error (HTTP %d, code %d)", e.HTTPStatusCode, e.Code)
}

// RequestError 表示请求、读取或响应解析失败。可用 errors.Is 判断 context 取消/超时。
type RequestError struct {
	Operation      string
	HTTPStatusCode int
	cause          error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("credit: %s failed (HTTP %d)", e.Operation, e.HTTPStatusCode)
}
func (e *RequestError) Unwrap() error { return e.cause }

func send[T any](ctx context.Context, c *Client, transport *http.Client, method, endpoint string, payload any) (*Response[T], error) {
	if ctx == nil {
		return nil, errors.New("credit: context is required")
	}
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, &RequestError{Operation: "encode request", cause: err}
		}
		body = bytes.NewReader(encoded)
	}
	r, err := http.NewRequestWithContext(ctx, method, c.config.BaseURL+endpoint, body)
	if err != nil {
		return nil, &RequestError{Operation: "create request", cause: err}
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("User-Agent", c.config.UserAgent)
	if payload != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	resp, err := transport.Do(r)
	if err != nil {
		return nil, &RequestError{Operation: "request", cause: err}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, &RequestError{Operation: "read response", HTTPStatusCode: resp.StatusCode, cause: err}
	}
	if len(raw) > maxResponseBytes {
		return nil, &RequestError{Operation: "response size limit", HTTPStatusCode: resp.StatusCode}
	}
	var envelope struct {
		Code    *int            `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &envelope) != nil || envelope.Code == nil {
		return nil, &RequestError{Operation: "decode response", HTTPStatusCode: resp.StatusCode}
	}
	result := &Response[T]{Code: *envelope.Code, Message: envelope.Message}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.Code != 0 {
		return result, &APIError{HTTPStatusCode: resp.StatusCode, Code: result.Code}
	}
	if len(envelope.Data) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Data), []byte("null")) || json.Unmarshal(envelope.Data, &result.Data) != nil {
		return nil, &RequestError{Operation: "decode response data", HTTPStatusCode: resp.StatusCode}
	}
	return result, nil
}

// CreateOrder 创建待支付订单并返回收银台链接，不执行用户支付。
func (c *Client) CreateOrder(ctx context.Context, request CreateOrderRequest) (*Response[*CreatedOrder], error) {
	if strings.TrimSpace(request.OrderName) == "" || strings.TrimSpace(request.MerchantOrderNo) == "" {
		return nil, errors.New("credit: order name and merchant order number are required")
	}
	if request.Amount <= 0 {
		return nil, errors.New("credit: amount must be a positive integer FPC")
	}
	payload := struct {
		CreateOrderRequest
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}{request, c.config.AppID, c.config.AppSecret}
	return send[*CreatedOrder](ctx, c, c.merchantHTTP, http.MethodPost, "/api/v1/merchant/orders", payload)
}

// IssuePoints 发放积分；应用必须通过发积分审核。调用成功后仍应检查入账状态。
func (c *Client) IssuePoints(ctx context.Context, request IssuePointsRequest) (*Response[*IssuedPoints], error) {
	if strings.TrimSpace(request.RecipientUsername) == "" || strings.TrimSpace(request.MerchantOrderNo) == "" {
		return nil, errors.New("credit: recipient username and merchant order number are required")
	}
	if request.Amount <= 0 {
		return nil, errors.New("credit: amount must be a positive integer FPC")
	}
	payload := struct {
		IssuePointsRequest
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}{request, c.config.AppID, c.config.AppSecret}
	return send[*IssuedPoints](ctx, c, c.merchantHTTP, http.MethodPost, "/api/v1/merchant/point-issues", payload)
}

// GetCashierOrder 需要 WithHTTPClient 提供积分服务已登录用户的 CookieJar。
// 不使用应用密钥冒充用户登录，也不把收银台查询当作商户支付确认接口。
func (c *Client) GetCashierOrder(ctx context.Context, orderNo string) (*Response[*CashierOrder], error) {
	if strings.TrimSpace(orderNo) == "" || orderNo == "." || orderNo == ".." {
		return nil, errors.New("credit: order number is required")
	}
	return send[*CashierOrder](ctx, c, c.cashierHTTP, http.MethodGet, "/api/v1/payment/orders/"+url.PathEscape(orderNo), nil)
}

// CashierURL 构造收银台页面地址，不发送请求。优先使用下单返回的 PaymentURL。
func (c *Client) CashierURL(orderNo string) (string, error) {
	if strings.TrimSpace(orderNo) == "" {
		return "", errors.New("credit: order number is required")
	}
	return c.config.BaseURL + "/paying?" + url.Values{"order_no": {orderNo}}.Encode(), nil
}
