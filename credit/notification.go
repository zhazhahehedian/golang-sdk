package credit

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
)

var ErrInvalidNotification = errors.New("credit: invalid notification")
var ErrInvalidSignature = errors.New("credit: invalid notification signature")

// VerifyNotification 校验平台 GET 通知。query 应来自 url.ParseQuery(r.URL.RawQuery)，
// 且调用者必须先处理 ParseQuery 的错误，以免忽略损坏的查询参数。
// 参数只解码一次；包括空值、未知扩展字段在内的所有字段均参与签名（sign/sign_type 除外）。
// 此方法不修改 query，不检查本地订单，也不负责去重或处理重放。
func (c *Client) VerifyNotification(query url.Values) (*Notification, error) {
	for _, values := range query {
		if len(values) != 1 {
			return nil, ErrInvalidNotification
		}
	}
	for _, key := range []string{"order_no", "merchant_order_no", "app_id", "amount", "fee_amount", "payee_amount", "status", "sign", "sign_type"} {
		values, ok := query[key]
		if !ok || (values[0] == "" && key != "merchant_order_no") {
			return nil, ErrInvalidNotification
		}
	}
	if query.Get("app_id") != c.config.AppID || query.Get("sign_type") != "MD5" {
		return nil, ErrInvalidNotification
	}
	signature := query.Get("sign")
	if len(signature) != md5.Size*2 || signature != strings.ToLower(signature) {
		return nil, ErrInvalidSignature
	}
	actual, err := hex.DecodeString(signature)
	if err != nil {
		return nil, ErrInvalidSignature
	}
	keys := make([]string, 0, len(query)-2)
	for key := range query {
		if key != "sign" && key != "sign_type" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var canonical strings.Builder
	for index, key := range keys {
		if index > 0 {
			canonical.WriteByte('&')
		}
		canonical.WriteString(key)
		canonical.WriteByte('=')
		canonical.WriteString(query.Get(key))
	}
	canonical.WriteString(c.config.AppSecret)
	// MD5 为平台规定的兼容协议，不替换成其他签名算法。
	expected := md5.Sum([]byte(canonical.String()))
	if subtle.ConstantTimeCompare(actual, expected[:]) != 1 {
		return nil, ErrInvalidSignature
	}
	return &Notification{
		OrderNo: query.Get("order_no"), MerchantOrderNo: query.Get("merchant_order_no"),
		AppID: query.Get("app_id"), Amount: query.Get("amount"), FeeAmount: query.Get("fee_amount"),
		PayeeAmount: query.Get("payee_amount"), Status: OrderStatus(query.Get("status")),
		Sign: signature, SignType: query.Get("sign_type"),
	}, nil
}
