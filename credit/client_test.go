package credit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const createResponse = `{"code":0,"data":{"order_no":"000000000000000001","payment_url":"https://credit.yuimeta.com/paying?order_no=000000000000000001","expires_at":"2026-07-05T10:30:00Z"}}`
const issueResponse = `{"code":0,"data":{"order_no":"000000000000000002","status":"success","settlement_status":"completed","amount":"5.00","recipient_username":"fishpi_user"}}`

func testClient(t *testing.T, handler http.HandlerFunc, options ...Option) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := NewClient(Config{AppID: "app_xxx", AppSecret: "sec_xxx", BaseURL: server.URL, UserAgent: "credit-test"}, options...)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func respond(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

func TestCreateOrderContract(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != "POST" || r.URL.Path != "/api/v1/merchant/orders" || r.URL.RawQuery != "" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("User-Agent") != "credit-test" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("unexpected request headers")
		}
		var body map[string]any
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		if err := decoder.Decode(&body); err != nil {
			t.Error(err)
		}
		want := map[string]any{"app_id": "app_xxx", "app_secret": "sec_xxx", "order_name": "鱼排周边", "merchant_order_no": "M202607050001", "payer_oid": "1780718050570", "amount": json.Number("10"), "notify_url": "https://example.com/notify", "return_url": "https://example.com/return", "remark": "备注"}
		if !reflect.DeepEqual(body, want) {
			t.Error("create-order JSON differs from document")
		}
		respond(w, createResponse)
	})
	r, err := c.CreateOrder(context.Background(), CreateOrderRequest{OrderName: "鱼排周边", MerchantOrderNo: "M202607050001", PayerOID: "1780718050570", Amount: 10, NotifyURL: "https://example.com/notify", ReturnURL: "https://example.com/return", Remark: "备注"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Code != 0 || r.Data.OrderNo != "000000000000000001" || r.Data.PaymentURL != "https://credit.yuimeta.com/paying?order_no=000000000000000001" || r.Data.ExpiresAt.Format(time.RFC3339) != "2026-07-05T10:30:00Z" {
		t.Fatal("response fields were lost or rewritten")
	}
	if calls.Load() != 1 {
		t.Fatal("order was retried")
	}
}

func TestIssuePointsAndOptionalFields(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		d := json.NewDecoder(r.Body)
		d.UseNumber()
		if err := d.Decode(&body); err != nil {
			t.Error(err)
		}
		if r.URL.RawQuery != "" || r.Method != "POST" {
			t.Error("credentials should only be in JSON")
		}
		if r.URL.Path == "/api/v1/merchant/point-issues" {
			want := map[string]any{"app_id": "app_xxx", "app_secret": "sec_xxx", "recipient_username": "fishpi_user", "merchant_order_no": "P202607090001", "amount": json.Number("5"), "remark": "奖励"}
			if !reflect.DeepEqual(body, want) {
				t.Error("issue-points JSON differs from document")
			}
			respond(w, issueResponse)
		} else {
			for _, key := range []string{"payer_oid", "notify_url", "return_url", "remark"} {
				if _, ok := body[key]; ok {
					t.Errorf("optional %s should be omitted", key)
				}
			}
			if body["amount"] != json.Number("9223372036854775807") {
				t.Error("integer precision lost")
			}
			respond(w, createResponse)
		}
	})
	r, err := c.IssuePoints(context.Background(), IssuePointsRequest{RecipientUsername: "fishpi_user", MerchantOrderNo: "P202607090001", Amount: 5, Remark: "奖励"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Data.Amount != "5.00" || r.Data.Status != OrderSuccess || r.Data.SettlementStatus != SettlementCompleted || r.Data.RecipientUsername != "fishpi_user" {
		t.Fatal("issuance response mismatch")
	}
	if _, err := c.CreateOrder(context.Background(), CreateOrderRequest{OrderName: "order", MerchantOrderNo: "M1", Amount: math.MaxInt64}); err != nil {
		t.Fatal(err)
	}
}

func TestCashierSessionIsolation(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	original := &http.Client{Jar: jar, Timeout: time.Second}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" || r.Header.Get("Authorization") != "" {
			t.Error("unexpected query or Bearer credentials")
		}
		if r.Method == "GET" {
			if r.URL.EscapedPath() != "/api/v1/payment/orders/0001%2F%3F%23" {
				t.Errorf("order path not escaped: %s", r.URL.EscapedPath())
			}
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value != "user-session" {
				t.Error("cashier session missing")
			}
			raw, _ := io.ReadAll(r.Body)
			if len(raw) != 0 {
				t.Error("cashier GET must not send an app secret body")
			}
			respond(w, `{"code":0,"data":{"order_no":"0001","order_name":"周边","amount":"10.00","fee_amount":"0.00","payee_amount":"2.50","market_listing_id":null,"market_system_amount":"7.50","status":"refunded","settlement_status":"completed","notify_status":"success","notify_error":"","refund_status":"completed","refund_amount":"10.00","refund_error":"","return_url":"https://example.com/return","expires_at":"2026-07-02T10:30:00Z","paid_at":"2026-07-02T10:12:00Z","settled_at":null,"refunded_at":"2026-07-02T10:15:00Z","can_pay":false,"payer_mismatch":true,"merchant":{"app_name":"示例应用","homepage_url":"https://example.com","app_description":"应用简介","username":"fishpi_user"}}}`)
		} else {
			if r.Header.Get("Cookie") != "" {
				t.Error("user cookie leaked into merchant request")
			}
			respond(w, issueResponse)
		}
	}, WithHTTPClient(original))
	u, _ := url.Parse(c.config.BaseURL)
	jar.SetCookies(u, []*http.Cookie{{Name: "session", Value: "user-session"}})
	result, err := c.GetCashierOrder(context.Background(), "0001/?#")
	if err != nil {
		t.Fatal(err)
	}
	order := result.Data
	if order.MarketListingID != nil || order.SettledAt != nil || order.PaidAt == nil || order.RefundedAt == nil || order.PayeeAmount != "2.50" || order.CanPay || !order.PayerMismatch || order.RefundStatus != RefundCompleted || order.Merchant.AppName != "示例应用" {
		t.Fatal("cashier values or nullable fields lost")
	}
	if _, err := c.IssuePoints(context.Background(), IssuePointsRequest{RecipientUsername: "user", MerchantOrderNo: "P1", Amount: 5}); err != nil {
		t.Fatal(err)
	}
	if original.Jar != jar || original.CheckRedirect != nil || original.Timeout != time.Second {
		t.Error("caller HTTP client changed")
	}
}

func TestFailuresAndNoRetry(t *testing.T) {
	tests := []struct {
		name, body string
		status     int
		apiErr     bool
		code       int
	}{
		{"business", `{"code":7,"message":"应用未上架","data":null}`, 200, true, 7},
		{"secret", `{"code":403,"message":"密钥错误"}`, 403, true, 403},
		{"throttle", `{"code":429,"message":"请求过快"}`, 429, true, 429},
		{"HTTP error with success code", `{"code":0,"data":{}}`, 500, true, 0},
		{"html", "<html>sec_xxx</html>", 500, false, 0},
		{"missing code", `{"data":{}}`, 200, false, 0},
		{"empty code", `{"code":null,"data":{}}`, 200, false, 0},
		{"missing data", `{"code":0}`, 200, false, 0},
		{"null data", `{"code":0,"data":null}`, 200, false, 0},
		{"bad data", `{"code":0,"data":[]}`, 200, false, 0},
		{"bad JSON", "{sec_xxx", 200, false, 0},
		{"oversized", strings.Repeat("x", maxResponseBytes+1), 200, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			})
			r, err := c.CreateOrder(context.Background(), CreateOrderRequest{OrderName: "order", MerchantOrderNo: "M1", Amount: 1})
			if err == nil {
				t.Fatal("expected failure")
			}
			if strings.Contains(err.Error(), "sec_xxx") {
				t.Error("response leaked into error string")
			}
			var apiErr *APIError
			if errors.As(err, &apiErr) != tt.apiErr {
				t.Fatalf("wrong error type: %T", err)
			}
			if tt.apiErr && (apiErr.HTTPStatusCode != tt.status || apiErr.Code != tt.code || r.Code != tt.code) {
				t.Error("error code or HTTP status lost")
			}
			if tt.name == "business" && r.Message != "应用未上架" {
				t.Error("message field lost")
			}
			if calls.Load() != 1 {
				t.Fatal("failed monetary request was retried")
			}
		})
	}
}

func TestNoRedirect(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Location", "/unexpected")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	if _, err := c.IssuePoints(context.Background(), IssuePointsRequest{RecipientUsername: "user", MerchantOrderNo: "P1", Amount: 1}); err == nil {
		t.Error("redirect accepted")
	}
	if calls.Load() != 1 {
		t.Error("app secret could have followed a redirect")
	}
}

func TestConfigurationAndValidation(t *testing.T) {
	for _, cfg := range []Config{{}, {AppID: "app"}, {AppSecret: "secret"}, {AppID: "app", AppSecret: "secret", BaseURL: "file:///tmp/api"}, {AppID: "app", AppSecret: "secret", BaseURL: "https://user:secret@example.com"}, {AppID: "app", AppSecret: "secret", BaseURL: "https://example.com?secret=x"}, {AppID: "app", AppSecret: "secret", BaseURL: "https://example.com#fragment"}} {
		if _, err := NewClient(cfg); err == nil {
			t.Error("invalid config accepted")
		}
	}
	if _, err := NewClient(Config{AppID: "app", AppSecret: "secret"}, WithHTTPClient(nil)); err == nil {
		t.Error("nil HTTP client accepted")
	}
	c, err := NewClient(Config{AppID: "app", AppSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if c.config.BaseURL != DefaultBaseURL || c.merchantHTTP.Timeout != 30*time.Second {
		t.Fatal("wrong defaults")
	}
	u, err := c.CashierURL("0001&next=中文")
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(u)
	if len(parsed.Query()) != 1 || parsed.Query().Get("order_no") != "0001&next=中文" {
		t.Error("cashier URL was not encoded")
	}
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid input sent request") })
	checks := []func() error{
		func() error {
			_, e := c.CreateOrder(context.Background(), CreateOrderRequest{MerchantOrderNo: "M1", Amount: 1})
			return e
		},
		func() error {
			_, e := c.CreateOrder(context.Background(), CreateOrderRequest{OrderName: "order", Amount: 1})
			return e
		},
		func() error {
			_, e := c.CreateOrder(context.Background(), CreateOrderRequest{OrderName: "order", MerchantOrderNo: "M1"})
			return e
		},
		func() error {
			_, e := c.IssuePoints(context.Background(), IssuePointsRequest{MerchantOrderNo: "P1", Amount: 1})
			return e
		},
		func() error {
			_, e := c.IssuePoints(context.Background(), IssuePointsRequest{RecipientUsername: "user", Amount: 1})
			return e
		},
		func() error {
			_, e := c.IssuePoints(context.Background(), IssuePointsRequest{RecipientUsername: "user", MerchantOrderNo: "P1", Amount: -1})
			return e
		},
		func() error { _, e := c.GetCashierOrder(context.Background(), " "); return e },
		func() error { _, e := c.CashierURL(""); return e },
		func() error {
			_, e := c.CreateOrder(nil, CreateOrderRequest{OrderName: "order", MerchantOrderNo: "M1", Amount: 1})
			return e
		},
	}
	for i, check := range checks {
		if check() == nil {
			t.Errorf("case %d: expected validation error", i)
		}
	}
}

func TestCancellationAndConcurrentClients(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { respond(w, createResponse) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.CreateOrder(ctx, CreateOrderRequest{OrderName: "order", MerchantOrderNo: "M1", Amount: 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context cancellation lost: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := c.CreateOrder(context.Background(), CreateOrderRequest{OrderName: "order", MerchantOrderNo: fmt.Sprint(i), Amount: 1}); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
}
