package credit

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func notificationFixture() url.Values {
	return url.Values{
		"order_no": {"000000000000000001"}, "merchant_order_no": {"M202607020001"},
		"app_id": {"app_xxx"}, "amount": {"10.00"}, "fee_amount": {"0.00"},
		"payee_amount": {"0.00"}, "status": {"success"}, "sign_type": {"MD5"},
		// 独立使用 .NET MD5/UTF-8 对文档签名串 + sec_xxx 计算的固定向量。
		"sign": {"2468b15f65e9a5bdf2e24558bf29544e"},
	}
}

func notificationClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClient(Config{AppID: "app_xxx", AppSecret: "sec_xxx"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNotificationKnownVectors(t *testing.T) {
	c := notificationClient(t)
	for _, tt := range []struct {
		name   string
		change func(url.Values)
	}{
		{"payment", func(url.Values) {}},
		{"refund and empty merchant number", func(q url.Values) {
			q.Set("status", "refunded")
			q.Set("merchant_order_no", "")
			q.Set("sign", "9a0c236c58032590ff99262083dfa5d9")
		}},
		{"extension and decoded characters", func(q url.Values) {
			q.Set("extra", "备注 A+B&=中文")
			q.Set("sign", "b1ea980fd4f058252c8ec2c28a64aaf0")
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			q := notificationFixture()
			tt.change(q)
			decoded, err := url.ParseQuery(q.Encode())
			if err != nil {
				t.Fatal(err)
			}
			before := decoded.Encode()
			n, err := c.VerifyNotification(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if n.OrderNo != "000000000000000001" || n.Amount != "10.00" || n.AppID != "app_xxx" || string(n.Status) != q.Get("status") || n.MerchantOrderNo != q.Get("merchant_order_no") {
				t.Fatal("notification fields lost")
			}
			if decoded.Encode() != before || !reflect.DeepEqual(q, decoded) {
				t.Error("input query mutated")
			}
		})
	}
}

func TestNotificationRejectsTampering(t *testing.T) {
	c := notificationClient(t)
	for _, key := range []string{"order_no", "merchant_order_no", "amount", "fee_amount", "payee_amount", "status"} {
		t.Run(key, func(t *testing.T) {
			q := notificationFixture()
			q.Set(key, q.Get(key)+"x")
			if n, err := c.VerifyNotification(q); !errors.Is(err, ErrInvalidSignature) || n != nil {
				t.Fatal("modified signed field accepted")
			}
		})
	}
	q := notificationFixture()
	q.Set("extra", "")
	if _, err := c.VerifyNotification(q); !errors.Is(err, ErrInvalidSignature) {
		t.Error("unknown empty fields must also be signed")
	}
	other, err := NewClient(Config{AppID: "app_xxx", AppSecret: "wrong-secret"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.VerifyNotification(notificationFixture()); !errors.Is(err, ErrInvalidSignature) {
		t.Error("wrong secret accepted")
	}
}

func TestNotificationRejectsMalformedInput(t *testing.T) {
	c := notificationClient(t)
	for _, change := range []func(url.Values){
		func(q url.Values) { q.Del("merchant_order_no") },
		func(q url.Values) { q.Set("order_no", "") },
		func(q url.Values) { q.Set("app_id", "different-app") },
		func(q url.Values) { q.Set("sign_type", "SHA256") },
		func(q url.Values) { q.Set("sign_type", "md5") },
		func(q url.Values) { q.Add("amount", "10.00") },
		func(q url.Values) { q.Add("sign", q.Get("sign")) },
		func(q url.Values) { q["amount"] = nil },
		func(q url.Values) { q["extra"] = []string{"a", "b"} },
		func(q url.Values) { q.Set("sign", strings.ToUpper(q.Get("sign"))) },
		func(q url.Values) { q.Set("sign", strings.Repeat("z", 32)) },
		func(q url.Values) { q.Set("sign", "bad") },
	} {
		q := notificationFixture()
		change(q)
		if n, err := c.VerifyNotification(q); err == nil || n != nil {
			t.Fatal("malformed notification accepted")
		}
	}
	if _, err := c.VerifyNotification(nil); !errors.Is(err, ErrInvalidNotification) {
		t.Error("nil query accepted")
	}
}
