# V2.3.5 接手开发清单

核对日期：2026-09-11。代码基线：`e3a38b1`。来源：浏览器中已打开的
[开放 API V2.3.5](https://fishpi.cn/article/1636516552191) 和
[OAuth / OpenID 接入说明，2026-07-06 版](https://fishpi.cn/article/1782119519493)。

本清单记录与本地 SDK 的接口级差异。原有接口尚未逐字段重新验收，README 中原有勾选不代表本轮已完成真实服务验证。

## 本批实现

| API | SDK 方法 | 验证 |
| --- | --- | --- |
| GET /openid/login，增加 fishpi.scope | GetOpenIdUrlWithScopes | URL 参数编码及旧方法兼容 |
| POST /openid/verify，追加令牌字段 | PostOpenIdVerifyWithTokens | 纯文本、CRLF、冒号值、失败及旧方法兼容 |
| POST /openid/token | PostOpenIdToken | JSON、令牌轮换结果、日志及鉴权隔离 |
| GET /openid/user/profile | GetOAuthProfile | Bearer、模型解析 |
| GET /openid/user/detail | GetOAuthUserDetail | Bearer、详细资料独立模型 |
| GET /openid/user/membership | GetOAuthMembership | VIP 状态、毫秒时间戳 |
| GET /openid/user/points | GetOAuthPoints | 分页、积分流水及来源字段 |
| GET /openid/user/articles | GetOAuthArticles | 分页、公开发帖模型 |
| GET /api/article-drafts | GetArticleDrafts | 列表嵌套及时间戳 |
| POST /api/article-drafts | PostArticleDraft | 新建、更新、空字符串、显式 false/0 |
| GET /api/article-drafts/{id} | GetArticleDraft | 正文、思绪正文、打赏正文 |
| DELETE /api/article-drafts/{id} | DeleteArticleDraft | 方法、路径及删除 ID |
| GET /api/repeater/items | GetRepeaterItems | 筛选条件、列表 |
| GET /api/repeater/next | GetRepeaterNext | 排除 ID、空结果 |
| POST /api/repeater | PostRepeater | JSON 请求及内容响应 |
| POST /api/repeater/{id}/like | PostRepeaterLike | 点赞切换结果 |
| POST /api/columns/{columnId}/cover | PostColumnCover | 设置及空字符串清空封面 |

实现沿用现有 `ApiResponse[T]` 风格：业务失败检查 `Code` / `Msg`；新增接口的 HTTP 异常和无效响应返回 `error`。`Data` 在失败或无结果时可能为 nil。

普通新接口沿用 API Key 查询参数，POST 同时按文档携带 JSON `apiKey`。OAuth 使用独立客户端，不继承普通客户端的 API Key、Cookie、Authorization 或请求转储/自定义 JSON 诊断钩子；继承当前 BaseURL 与 User-Agent。OAuth 不跟随重定向、不自动重试续签、不自动持久化令牌。

旧 `GetOpenIdUrl` / `PostOpenIdVerify` 的签名保留。旧校验方法委托给新增解析器，只返回用户 ID，兼容没有令牌字段的旧响应；不再修改调用者 map，校验请求不再进入原始转储日志。调用者负责把回调与自己发起的登录会话、预期 return_to、nonce 绑定。

## 已识别但本批未实现的差异

| 模块 | 差异 / 下一步 |
| --- | --- |
| 鉴权 | 注册流程 `/captcha`、`/register`、`/verify`、`/register2` 原先未实现；需要独立的验证码、Cookie 会话流程 |
| 帖子 | 新增长篇列表 `/api/articles/recent/long`；类型 6；新增 `articleBypassCensor` 字段，历史思绪仍可读但不能新建 |
| 评论线程 | `/comment/thread/parents`、`/comment/thread/replies` 未实现 |
| 长篇段评 | `/comment/paragraph`、`/comment/paragraph/summary`、`/comment/paragraph/thread/parents` 未实现 |
| 历史版本 | `/article/{id}/revisions/list`、`/article/{id}/revisions/{revisionId}`、`/comment/{id}/revisions` 未实现 |
| 职业系统用户侧 | `/api/profession/me`、详情、记录、primary、skip、privacy、公开资料及 ranking 未实现；先核对 Cookie / CSRF 与 API Key 支持边界 |
| 职业系统管理侧 | 元数据、catalog、catalog-summary、catalog-detail、scheme-impact、配置导出/导入/预检、排序；职业定义、等级方案、自动化的 draft/publish/retire/copy/rollback/test 未实现 |
| 职业金手指 | `/api/gold-finger/profession/experience`、`/api/gold-finger/profession/query` 未实现 |
| 鱼游 | `/activities`、`/api/fish-games` 及编辑、投票、评论等接口未实现；需核对完整响应模型与权限 |
| 管理操作 | `/admin/user/{userId}/deactivate` 未实现；应先确认权限及停用语义 |
| 既有聊天室与 WebSocket | 已有自动重连及事件解析；README 已注明聊天结构待与真实数据对照，仍需协议回归 |
| 既有上传接口 | README 中上传限制尚未完成；需核对大小、类型和 multipart 字段 |

## 文档差异与真实验证边界

- OAuth 详细资料的 `userNo`、`userAppRole` 在专门接入说明中为字符串，在 API 表格中为数字示例，使用 `json.Number` 接受两者。`sysMetal` 保留为 `json.RawMessage`，兼容字符串或对象，避免强行复用普通用户模型。
- 举报接口的文字要求 form-urlencoded，但 curl 示例使用 JSON。原实现需要单独对照服务端或实际响应，本批不据此修改。
- 通用文档要求非例外接口间隔至少 30 秒；活跃度间隔至少 10 分钟；user-channel 需重连且不要发心跳。本批没有添加全局限流策略。
- 所有自动测试使用 `httptest` 本地服务器。已覆盖请求/响应契约和凭据隔离，不代表服务器已经接受过真实令牌，也不代表写操作已经在线验收。
- 后续真实验收应使用可控账号，分别验证 OAuth 授权/续签、草稿读写删除、复读机内容与点赞切换、专栏封面设置和清空；遵守频率限制并保留可核对的响应证据。

## 本地检查

本轮验证结果：WSL / Linux Go 1.27.0 下 `go test -race ./...` 和 `go vet ./...` 通过；Go 1.24.0 下 `go test ./...` 通过；原有 usage 示例编译通过。新增 GitHub Actions 配置尚未在远程运行。

在既有 WSL Go 环境中运行：

```sh
go test -race ./...
go vet ./...
```

原示例是独立 workspace，编译而不运行实际业务请求：

```sh
cd examples/usage
go build ./...
```

不改变 module 路径和既有导入方式。本批不包含版本发布。
