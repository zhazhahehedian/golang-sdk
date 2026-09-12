package credit_test

import (
	"fmt"
	"net/url"

	"github.com/fishpioffical/golang-sdk/credit"
)

func ExampleClient_CashierURL() {
	client, err := credit.NewClient(credit.Config{AppID: "app_xxx", AppSecret: "sec_xxx"})
	if err != nil {
		panic(err)
	}
	address, err := client.CashierURL("000000000000000001")
	if err != nil {
		panic(err)
	}
	fmt.Println(address)
	// Output: https://credit.yuis.cc/paying?order_no=000000000000000001
}

func ExampleClient_VerifyNotification() {
	client, err := credit.NewClient(credit.Config{AppID: "app_xxx", AppSecret: "sec_xxx"})
	if err != nil {
		panic(err)
	}
	query, err := url.ParseQuery("order_no=000000000000000001&merchant_order_no=M202607020001&app_id=app_xxx&amount=10.00&fee_amount=0.00&payee_amount=0.00&status=success&sign_type=MD5&sign=2468b15f65e9a5bdf2e24558bf29544e")
	if err != nil {
		panic(err)
	}
	notification, err := client.VerifyNotification(query)
	if err != nil {
		panic(err)
	}
	// 实际应用中还应核对本地订单、金额并完成幂等更新。
	fmt.Println(notification.OrderNo, notification.Status, notification.Amount)
	// Output: 000000000000000001 success 10.00
}
