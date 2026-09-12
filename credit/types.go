package credit

import "time"

// Response 是积分服务的响应；错误说明字段为 message，而非社区 API 的 msg。
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
}

// CreateOrderRequest 创建待用户支付的订单，Amount 单位为整数 FPC。
// 商户单号由调用者持久化并保证业务唯一性；SDK 不生成或自动重试订单。
type CreateOrderRequest struct {
	OrderName       string `json:"order_name"`
	MerchantOrderNo string `json:"merchant_order_no"`
	PayerOID        string `json:"payer_oid,omitempty"`
	Amount          int64  `json:"amount"`
	NotifyURL       string `json:"notify_url,omitempty"`
	ReturnURL       string `json:"return_url,omitempty"`
	Remark          string `json:"remark,omitempty"`
}

type CreatedOrder struct {
	OrderNo    string    `json:"order_no"`
	PaymentURL string    `json:"payment_url"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// IssuePointsRequest 发放积分，需要应用单独通过发积分审核。
type IssuePointsRequest struct {
	RecipientUsername string `json:"recipient_username"`
	MerchantOrderNo   string `json:"merchant_order_no"`
	Amount            int64  `json:"amount"`
	Remark            string `json:"remark,omitempty"`
}

type IssuedPoints struct {
	OrderNo           string           `json:"order_no"`
	Status            OrderStatus      `json:"status"`
	SettlementStatus  SettlementStatus `json:"settlement_status"`
	Amount            string           `json:"amount"`
	RecipientUsername string           `json:"recipient_username"`
}

// CashierOrder 用于已登录用户的收银台展示，不是商户服务端的支付确认凭据。
// 金额保留服务端十进制字符串；平台单号始终保留字符串及前导零。
type CashierOrder struct {
	OrderNo            string           `json:"order_no"`
	OrderName          string           `json:"order_name"`
	Amount             string           `json:"amount"`
	FeeAmount          string           `json:"fee_amount"`
	PayeeAmount        string           `json:"payee_amount"`
	MarketListingID    *string          `json:"market_listing_id"`
	MarketSystemAmount string           `json:"market_system_amount"`
	Status             OrderStatus      `json:"status"`
	SettlementStatus   SettlementStatus `json:"settlement_status"`
	NotifyStatus       NotifyStatus     `json:"notify_status"`
	NotifyError        string           `json:"notify_error"`
	RefundStatus       RefundStatus     `json:"refund_status"`
	RefundAmount       string           `json:"refund_amount"`
	RefundError        string           `json:"refund_error"`
	ReturnURL          string           `json:"return_url"`
	ExpiresAt          time.Time        `json:"expires_at"`
	PaidAt             *time.Time       `json:"paid_at"`
	SettledAt          *time.Time       `json:"settled_at"`
	RefundedAt         *time.Time       `json:"refunded_at"`
	CanPay             bool             `json:"can_pay"`
	PayerMismatch      bool             `json:"payer_mismatch"`
	Merchant           CashierMerchant  `json:"merchant"`
}

type CashierMerchant struct {
	AppName        string `json:"app_name"`
	HomepageURL    string `json:"homepage_url"`
	AppDescription string `json:"app_description"`
	Username       string `json:"username"`
}

// Notification 经过应用绑定和签名校验的通知。验签不代表已完成本地业务核对。
// 必须再核对订单、金额，并幂等处理；只在业务处理成功后返回 2xx。
type Notification struct {
	OrderNo         string      `json:"order_no"`
	MerchantOrderNo string      `json:"merchant_order_no"`
	AppID           string      `json:"app_id"`
	Amount          string      `json:"amount"`
	FeeAmount       string      `json:"fee_amount"`
	PayeeAmount     string      `json:"payee_amount"`
	Status          OrderStatus `json:"status"`
	Sign            string      `json:"sign"`
	SignType        string      `json:"sign_type"`
}

// 状态使用开放字符串类型，保留服务端未来新增的值。
type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderProcessing OrderStatus = "processing"
	OrderSuccess    OrderStatus = "success"
	OrderSettling   OrderStatus = "settling"
	OrderRefunded   OrderStatus = "refunded"
	OrderExpired    OrderStatus = "expired"
	OrderFailed     OrderStatus = "failed"
)

// SettlementStatus 入账状态。文档仅给出了 completed，不推测其他枚举值。
type SettlementStatus string

const SettlementCompleted SettlementStatus = "completed"

type NotifyStatus string

const (
	NotifyNone    NotifyStatus = "none"
	NotifySkipped NotifyStatus = "skipped"
	NotifySuccess NotifyStatus = "success"
	NotifyFailed  NotifyStatus = "failed"
)

type RefundStatus string

const (
	RefundNone       RefundStatus = "none"
	RefundProcessing RefundStatus = "processing"
	RefundCompleted  RefundStatus = "completed"
	RefundFailed     RefundStatus = "failed"
)
