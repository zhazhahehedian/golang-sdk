# 鱼排积分接入

本模块依据浏览器中打开的[应用对接文档](https://credit.yuis.cc/docs/merchant/)（页面标注最后更新 2026-07-21）实现，于 2026-09-12 核对。
[接口目录](https://credit.yuis.cc/docs/api/)列有更多后台接口，但没有给出完整的请求、响应和鉴权说明；本模块覆盖已详细公开的应用接入流程。

## 支持范围

| 方法 | 接口 / 功能 | 鉴权 |
| --- | --- | --- |
| `CreateOrder` | POST /api/v1/merchant/orders | JSON app_id / app_secret |
| `IssuePoints` | POST /api/v1/merchant/point-issues | JSON app_id / app_secret；应用需单独通过审核 |
| `CashierURL` | 构造 /paying?order_no=… | 不发请求；页面由用户登录支付 |
| `GetCashierOrder` | GET /api/v1/payment/orders/{order_no} | 积分服务已登录用户会话 |
| `VerifyNotification` | 平台 GET 支付 / 退款回调 | 应用密钥 MD5 验签及 app_id 校验 |

应用需上架后才能下单。集市支付由平台页面提供；SDK 不代替用户确认支付。收银台数据查询用于展示，商户应以验签后的通知确认结果。

平台默认地址取当前文档站点 `https://credit.yuis.cc`。文档中的支付链接示例使用 `credit.yuimeta.com`，SDK 保留服务端实际返回的 `PaymentURL`，不会替换其域名。需要切换服务地址时设置 `Config.BaseURL`。

## 创建客户端

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/fishpioffical/golang-sdk/credit"
)

client, err := credit.NewClient(credit.Config{
    AppID: os.Getenv("FISHPI_CREDIT_APP_ID"),
    AppSecret: os.Getenv("FISHPI_CREDIT_APP_SECRET"),
})
if err != nil {
    return err
}
```

本客户端与社区 `sdk.FishPiSDK` 独立。社区 API Key、OAuth access_token 不能充当应用密钥；两者也不会自动互相传递。无需修改原有 `config.Provider` 或 module 路径。

默认超时 30 秒，可通过 `credit.WithHTTPClient(&http.Client{Timeout: ...})` 配置 Transport 和超时。客户端创建后可并发调用；调用期间不要修改共用的 HTTP Transport 或会话配置。每个网络方法都接受 `context.Context`。

## 商户下单

```go
created, err := client.CreateOrder(context.Background(), credit.CreateOrderRequest{
    OrderName: "鱼排周边",
    MerchantOrderNo: merchantOrderNo, // 先在本地持久化，由业务保证唯一性
    PayerOID: payerOID,              // 是否必填取决于应用配置
    Amount: 10,                     // 正整数 FPC
    NotifyURL: "https://example.com/credit/notify",
    ReturnURL: "https://example.com/orders",
    Remark: "周边订单",
})
if err != nil {
    return err
}
// 将平台单号与本地商户单号关联保存，然后让用户打开平台返回的链接。
fmt.Println(created.Data.OrderNo, created.Data.PaymentURL)
```

`CreateOrder` 只生成订单和收银台链接，不执行支付。优先使用 `PaymentURL`；需要从已有单号构造链接时，可以调用 `client.CashierURL(orderNo)`。

## 发积分

```go
issued, err := client.IssuePoints(ctx, credit.IssuePointsRequest{
    RecipientUsername: "fishpi_user",
    MerchantOrderNo: rewardOrderNo,
    Amount: 5,
    Remark: "活动奖励",
})
if err != nil {
    return err
}
// code=0 表示接口成功，实际业务还应根据订单、入账状态处理。
if issued.Data.Status == credit.OrderSuccess &&
    issued.Data.SettlementStatus == credit.SettlementCompleted {
    // 在本地记录已发放，避免业务重入造成重复奖励。
}
```

请求金额使用 `int64`，必须大于 0。响应的 `Amount`、`PayeeAmount` 等使用字符串保留 `5.00` 这样的十进制表示，不经 float64 转换。平台单号保持字符串，保留前导零。

SDK 不自动生成商户单号、不自动重试下单或发积分。文档未给出完整的服务端幂等保证；出现超时等不确定结果时，应先核实平台订单和本地记录，不能把超时直接视作发放失败后再次发放。

## 支付和退款通知

平台在支付完成或退款完成后向 `notify_url` 发起 GET 请求，接收方返回 2xx 即表示处理成功。

```go
func verifyCallback(client *credit.Client, r *http.Request) (*credit.Notification, error) {
    if r.Method != http.MethodGet {
        return nil, errors.New("unexpected callback method")
    }
    query, err := url.ParseQuery(r.URL.RawQuery)
    if err != nil {
        return nil, err
    }
    return client.VerifyNotification(query)
}
```

验签规则按文档实现：

1. 使用 URL 解码一次后的参数，拒绝重复字段或多个值。
2. 去掉 `sign` 和 `sign_type`；剩余所有字段（包括空值、扩展字段）按字段名排序。
3. 按 `key=value` 用 `&` 拼接，**不再次 URL 编码**，末尾直接追加应用密钥。
4. 计算 UTF-8 字节的 MD5、小写十六进制，用恒定时间比较摘要。
5. 必须有完整通知字段，`sign_type` 必须为 `MD5`，`app_id` 必须匹配当前客户端。文档允许 `merchant_order_no` 为空，但该字段不能缺失。

建议使用 `url.ParseQuery` 并处理错误，而不是直接使用会丢弃解析错误的 `r.URL.Query()`。SDK 不修改传入参数。

验签后，业务仍需核对平台单号、商户单号和预期金额。`OrderSuccess` 与 `OrderRefunded` 必须分别处理，不能把“验签通过”统一当作发货条件。同一通知可能重复发送；在本地事务中完成幂等更新后再返回 2xx。通知文档未提供时间戳或 nonce，因此 SDK 不声称能单独防重放。

## 收银台数据

`GetCashierOrder` 需要用户已在积分服务登录。文档没有提供程序化登录 API，SDK 不推测 Cookie 名称或复制浏览器会话。

调用者可以使用自己已经建立的、针对积分服务域名的 `http.CookieJar`：

```go
cashierClient, err := credit.NewClient(credit.Config{
    AppID: appID,
    AppSecret: appSecret,
}, credit.WithHTTPClient(&http.Client{
    Timeout: 30 * time.Second,
    Jar: authenticatedJar,
}))
if err != nil {
    return err
}
order, err := cashierClient.GetCashierOrder(ctx, orderNo)
```

SDK 复制传入的 `http.Client` 配置，不修改原对象。CookieJar 仅用于收银台查询；商户 POST 请求不带该会话。收银台 GET 不带应用密钥。平台返回的空时间字段和 `market_listing_id=null` 保留为 nil。

## 错误与验证

积分响应使用 `code` / `data` / `message`。非 2xx 或 `code != 0` 返回 `*credit.APIError`，有合法响应时同时返回 `Response`，可读取其 `Message`。无效 JSON、响应超限、读取失败等返回 `*credit.RequestError`。错误字符串不包含请求正文、应用密钥或原始响应；平台原始 `Message` 仅在返回对象中供调用者按需处理。

网络错误支持 `errors.Is(err, context.Canceled)` / `context.DeadlineExceeded`。SDK 不自动跟随 HTTP 重定向，避免把商户密钥转发到重定向目标。模块没有原始请求日志；若提供自定义 Transport，应确保它不记录应用密钥。

本地验证涵盖文档请求/响应、独立 MD5 固定向量、空商户单号、中文及特殊字符、重复参数和篡改、Cookie 隔离、金额精度、状态及可空时间、错误响应、重定向、context 取消和并发请求。

```sh
go test -race ./...
go vet ./...
```

所有测试使用本地 `httptest` 和示例密钥。尚未用真实应用进行下单、发积分、收银台登录或回调验收。后台应用/商品管理、人工支付确认、用户转账及管理接口未按地址清单猜测实现。

本轮已通过：WSL Go 1.27.0 下全仓测试、竞争检测及静态检查；Go 1.24.0 下 `credit` 包测试。原有 CI 的 `go test -race ./...` 会自动包含新增包，无需额外依赖。
