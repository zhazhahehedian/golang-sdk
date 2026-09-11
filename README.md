# FishPi Golang SDK

摸鱼派社区 Golang SDK，提供主要 API 接口封装。

已补齐开放 API V2.3.5 的 OAuth 用户资源、文章草稿、复读机转录站和专栏封面接口。
接口覆盖范围、待补齐项及验证边界见 [接手开发清单](docs/api-v2.3.5.md)，用法见 [新增接口示例](docs/v2.3.5-usage.md)。

## 特性

- ✅ **主要API支持** - HTTP 接口按业务域封装，已完成项见下方接口清单
- ✅ **类型安全** - 使用go-enum自动生成枚举类型
- ✅ **灵活配置** - 支持多种ConfigProvider（内存/文件）
- ✅ **WebSocket支持** - 泛型架构，支持聊天室、私聊、用户通知、帖子页事件
- ✅ **自动重连** - 可配置重连策略（固定延迟/指数退避）
- ✅ **心跳机制** - 可选的自定义心跳配置
- ✅ **消息解析** - 完整的WebSocket消息解析器
- ✅ **并发安全** - 细粒度锁优化，线程安全的实现
- ✅ **结构化日志** - 使用slog.Logger，支持自定义日志级别
- ✅ **错误处理** - 完整的错误处理和包装

## 存在的问题

- [ ] 聊天消息结构没有和原始结构对比

## 目录

- [接口](#接口)
- [安装](#安装)
- [快速开始](#快速开始)
- [WebSocket](#websocket)
- [类型安全](#类型安全)
- [开发](#开发)
- [许可证](#许可证)
- [相关链接](#相关链接)
- [贡献](#贡献)

## 接口
- [ ] 鉴权
    - [x] 获取apiKey
    - [x] 查询用户信息
    - [ ] 注册用户
        - [ ] 预注册账号 POST /register
        - [ ] 验证手机验证码 GET /verify
        - [ ] 注册账号 POST /register2
- [x] OpenID 接入
    - [x] 获取授权链接
    - [x] 签名校验获取openid
    - [x] 获取用户信息
- [x] OAuth 用户资源
    - [x] 带 scope 的授权链接及回跳校验令牌解析
    - [x] 续签访问令牌
    - [x] 基础资料、详细资料、VIP 状态
    - [x] 积分记录、公开发帖记录及分页
- [x] 文章草稿
    - [x] 查询列表、保存、查询详情、删除
- [x] 复读机转录站
    - [x] 查询列表、随机获取、上传、点赞或取消点赞
- [x] 专栏封面
    - [x] 更新或清空封面
- [ ] 新增
    - [x] 获取用户VIP信息 通用 GET /api/membership/{userId}
    - [x] 获取操作日志 通用 GET /logs/more
- [ ] 杂项
    - [x] 勋章链接生成
    - [ ] 客户端版本解析
- [ ] 通用
    - [x] 通过API累计用户的在线时间 WS
    - [x] 查询成员信息
    - [x] 用户名联想
    - [x] 用户常用表情
    - [x] 获取活跃度
    - [x] 获取签到状态
    - [x] 领取昨日活跃奖励
    - [x] 查询在昨日奖励领取状态
    - [x] 举报
    - [x] 查询最近注册的20个用户
    - [x] 转账
    - [x] 关注用户
    - [x] 取关用户
    - [x] 修改信息 通用 POST /api/settings/profiles
    - [x] 修改头像 通用 POST /api/settings/avatar
    - [x] 查询指定用户积分 GET /user/{username}/point
    - [x] 查询用户勋章 GET /user/{用户名}/medal
- [x] 通知
    - [x] 通知计数
    - [x] 通知详情
    - [x] 批量已读类型的通知
    - [x] 已读所有消息
- [ ] 聊天室
    - [x] 获取发送弹幕的价格
    - [x] 连接聊天室 WS
    - [x] 聊天室地址API
    - [x] 聊天历史消息
        - [x] 通过聊天消息的oId获取前后消息
    - [x] 发送消息
        - [x] 弹幕
        - [x] 红包
    - [x] 撤回消息
    - [x] 获取消息markdown
    - [x] 打开红包
    - [x] ~~获取表情包（已废弃，见表情包V2章节）~~
        - [x] 默认表情包
    - [x] ~~同步表情包（已废弃，见表情包V2章节）~~
    - [x] 获取禁言中的成员列表（思过崖）
- [x] 表情包V2
    - [x] 获取分组列表
    - [x] 获取分组内表情
    - [x] 上传URL到“全部”分组（一键上传失败）
    - [x] 创建分组
    - [x] 更新分组
    - [x] 删除分组
    - [x] 分组添加已有表情
    - [x] 分组添加URL表情
    - [x] 从分组移除表情
    - [x] 更新表情项（重命名/排序）
    - [x] 迁移旧表情到v2
- [ ] 图床
    - [x] 上传图片
    - [ ] 限制
- [ ] 帖子
    - [x] 发帖
    - [x] 更新帖子
    - [x] 帖子列表
        - [x] 特别注意
        - [x] 最近
        - [x] 按标签
        - [x] 按领域
    - [x] 获取指定帖子
    - [x] 获取指定用户的帖子列表
    - [x] 给文章点赞
    - [x] 感谢文章
    - [x] 获取帖子的评论列表
    - [x] 评论/回复
    - [x] 更新评论
    - [x] 给评论点赞
    - [x] 感谢评论
    - [x] 删除评论
    - [x] 获取帖子当前正在阅读的人数
        - [x] /article-channel WSS
    - [x] 收藏帖子
    - [x] 取消收藏帖子
    - [x] 关注帖子
    - [x] 取消关注帖子
    - [x] 打赏帖子
    - [x] 获取帖子的Markdown原文
- [x] 清风明月
    - [x] 获取清风明月列表
    - [x] 发布清风明月
    - [x] 获取指定用户的清风明月列表
- [ ] 私信
    - [x] 消息通知 /user-channel WSS
    - [x] 获取用户私聊历史消息
    - [x] 标记用户消息已读
    - [x] 获取私聊用户列表以及第一条消息
    - [x] 获取未读消息
    - [x] 撤回私聊消息
- [ ] 敏感操作
    - [ ] 永久注销删除用户
- [ ] 金手指
    - [ ] 注意
    - [x] 上传摸鱼大闯关关卡数据
    - [x] 查询用户最近登录的IP地址
    - [x] ~~添加勋章（已废弃，见新版勋章）~~
    - [x] ~~移除勋章（已废弃，见新版勋章）~~
    - [x] ~~移除勋章（通过userId）（已废弃，见新版勋章）~~
    - [x] 查询用户背包
    - [x] 调整用户背包
    - [x] 调整用户积分
    - [x] 获取用户活跃度
    - [x] 领取指定用户的昨日活跃奖励
    - [x] 发送指定用户系统通知（缺少金手指，未测试） 
- [x] 新版勋章系统（Medal v2）
  - [x] 用户侧API（需要登录）
    - [x] 读取当前登录用户的所有勋章列表
    - [x] 调整单个勋章顺序（上移/下移）
    - [x] 调整当前登录用户隐藏/显示勋章
    - [x] 读取指定用户当前展示中的勋章列表（用于主页展示）
  - [x] 管理侧API（需要管理员/金手指权限）
    - [x] 分页读取全部勋章列表(medal_admin_read)
    - [x] 按关键词搜索勋章列表(medal_admin_read)
    - [x] 读取指定勋章详细信息(medal_admin_read)
    - [x] 删除指定勋章(medal_admin_write)
    - [x] 编辑指定勋章(medal_admin_write)
    - [x] 新建勋章(medal_admin_write)
    - [x] 给指定用户发指定勋章(medal_admin_write)
    - [x] 给指定用户移除指定勋章(medal_admin_write)
    - [x] 读取指定勋章拥有的用户和拥有总数（分页）(medal_admin_read)
    - [x] 读取指定用户的所有勋章(medal_admin_read)
  - [x] description/data（变量渲染说明补充）
  - [x] 勋章生成页（SVG）
    - [x] 通过勋章ID显示勋章
    - [x] 通过勋章名称显示勋章
  - [x] 贴emoji
    - [x] 帖子：获取指定帖子
    - [x] 帖子：获取帖子的评论列表
    - [x] 帖子：给帖子添加表情
    - [x] 帖子：给评论添加表情
    - [x] 聊天室：聊天历史消息
    - [x] 聊天室：给聊天室消息添加表情
    - [x] 可用emoji值
    - [x] 聊天室WebSocket新增消息类型
    - [x] 帖子页 /article-channel 新增消息类型

## 安装

```bash
go get github.com/fishpioffical/golang-sdk
```

## 快速开始

### 基础使用

```go
package main

import (
    "fmt"
    "log"

    "github.com/fishpioffical/golang-sdk/sdk"
)

func main() {
    // 使用 API Key 创建 SDK 实例
    fishpi := sdk.NewSDKWithAPIKey("your-api-key")

    // 获取当前登录用户信息
    userInfo, err := fishpi.GetApiUser()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("用户: %s\n", userInfo.Data.UserName)
}
```

### 配置方式

```go
import (
    "github.com/fishpioffical/golang-sdk/config"
    "github.com/fishpioffical/golang-sdk/sdk"
)

// 方式1: API Key（推荐）
fishpi := sdk.NewSDKWithAPIKey("your-api-key")

// 方式2: JSON 文件配置
provider := config.NewFileJsonProvider("config.json")
fishpi := sdk.NewSDK(provider)

// 方式3: YAML 文件配置
provider := config.NewFileYamlProvider("config.yaml")
fishpi := sdk.NewSDK(provider)

// 方式4: 内存配置
provider := config.NewMemoryConfigProvider(&config.Config{
    BaseUrl: "https://fishpi.cn",
    ApiKey:  "your-api-key",
})
fishpi := sdk.NewSDK(provider)
```

### SDK 选项

```go
// 启用请求日志
fishpi := sdk.NewSDK(provider, 
    sdk.WithLogDir("./logs"),
)

// 自定义 JSON 解析器（用于调试）
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
fishpi := sdk.NewSDK(provider,
    sdk.WithCustomUnmarshaler(logger),
)
```

### 完整示例

查看 `examples/usage/main.go` 获取完整的使用示例，包括：

- **配置管理**: 使用 YAML 配置文件
- **日志配置**: 使用 devslog 美化日志输出
- **API 调用**: 用户信息、文章、清风明月等功能
- **WebSocket**: 聊天室、私聊、通知、帖子页事件的实时通信
- **错误处理**: 完整的错误处理示例

运行示例：

1. 增加本地配置 `_tmp/config.yaml`

```yaml
baseUrl: https://fishpi.cn
userAgent: your-user-agent-here

apiKey: your-api-key-here
username: your-username-here
password: your-password-here
passwordMd5: your-md5-password-here
mfaCode: your-mfa-code-here
totp: your-totp-secret-here

pointGoldFingerKey: "your-point-gold-finger-key-here"
livenessGoldFingerKey: "your-liveness-gold-finger-key-here"
gameGoldFingerKey: "your-game-gold-finger-key-here"
queryGoldFingerKey: "your-query-gold-finger-key-here"
metalGoldFingerKey: "your-metal-gold-finger-key-here"
itemGoldFingerKey: "your-item-gold-finger-key-here"
medalReadFingerKey: "your-medal-read-finger-key-here"
medalWriteFingerKey: "your-medal-write-finger-key-here"
notificationFingerKey: "your-notification-finger-key-here"

```

2. 运行代码

```bash
cd examples/usage
go run main.go
```

> **⚠️ 特别说明**
> 
> 1. 示例代码需要配置 `config.yaml` 文件，包含 `apiKey` 和 `baseUrl`
> 2. WebSocket 连接需要有效的 API Key
> 3. 部分功能（如金手指）需要特殊权限
> 4. 建议先在摸鱼派社区获取 API Key：https://fishpi.cn/settings/account

## WebSocket

基于泛型的 WebSocket 客户端，支持自动重连、心跳、自定义日志等企业级特性。

**核心特性**：
- 🔧 **泛型架构** - 类型安全的消息处理
- 🔄 **自动重连** - 默认启用，支持指数退避/固定延迟策略  
- 💓 **心跳机制** - 可配置心跳间隔和自定义消息
- 🔒 **并发安全** - 细粒度锁优化
- ⚙️ **函数式选项** - 灵活配置
- 📝 **结构化日志** - 使用 slog.Logger

**快速开始**：

```go
// 1. 聊天室（自动重连）
node, err := fishpi.GetChatroomNode()
if err != nil {
    log.Fatal(err)
}
ws := fishpi.NewChatroomWebSocket(node.Data)
ws.OnMessage(func(msg *types.ChatroomMsg) {
    fmt.Printf("收到消息: %s\n", msg.Type)
})
if err := ws.Connect(); err != nil {
    log.Fatal(err)
}
if _, err := fishpi.PostChatroomSend("Hello!"); err != nil {
    log.Fatal(err)
}

// 2. 私聊
privateWS := fishpi.NewPrivateChatWebSocket("target-user-name")
privateWS.OnMessage(func(msg *types.ChatChannelMsg) {
    fmt.Printf("[私聊] %s\n", msg.Content)
})
if err := privateWS.Connect(); err != nil {
    log.Fatal(err)
}

// 3. 用户通知（带心跳）
noticeWS := fishpi.NewUserNotificationWebSocket(
    sdk.WithHeartbeat[types.UserChannelMsg](30*time.Second, func() []byte {
        return []byte(`{"type":"ping"}`)
    }),
)
noticeWS.OnMessage(func(msg *types.UserChannelMsg) {
    fmt.Printf("[通知] %s\n", msg.Command)
})
if err := noticeWS.Connect(); err != nil {
    log.Fatal(err)
}

// 4. 帖子页事件
articleWS := fishpi.NewArticleChannelWebSocket("article-id", types.ArticleTypeNormal)
articleWS.OnMessage(func(msg *types.ArticleChannelMsg) {
    fmt.Printf("[帖子事件] %s\n", msg.Type)
})
if err := articleWS.Connect(); err != nil {
    log.Fatal(err)
}
```

**高级配置**：

```go
ws := fishpi.NewChatroomWebSocket("wss://...",
    // 重连策略
    sdk.WithReconnectStrategy[types.ChatroomMsg](&sdk.ExponentialBackoffStrategy{
        BaseDelay:  1 * time.Second,
        MaxDelay:   60 * time.Second,
        Multiplier: 2.0,
    }),
    // 最大重连次数（0=无限）
    sdk.WithMaxReconnectAttempts[types.ChatroomMsg](10),
    // 重连失败回调
    sdk.WithReconnectFailedCallback[types.ChatroomMsg](func(attempts int, err error) {
        log.Printf("重连失败: %v", err)
    }),
    // 自定义日志
    sdk.WithLogger[types.ChatroomMsg](customLogger),
)
```

> 📖 **详细 API 文档**: 查看 [examples/usage/main.go](./examples/usage/main.go) 获取所有 API 的实际使用示例

## 类型安全

SDK 使用 `go-enum` 自动生成所有枚举类型：

```go
// 枚举使用
articleType := types.ArticleListTypeHot
fmt.Println(articleType.String()) // "hot"

// 解析枚举
parsed, err := types.ParseArticleListType("hot")
```

## 开发

```bash
# 克隆项目
git clone https://github.com/fishpioffical/golang-sdk

# 安装依赖
go mod download

# 生成枚举代码
go generate ./...
```

## 许可证

Apache 2.0

## 相关链接

- [摸鱼派社区](https://fishpi.cn)
- [API文档 v2.2.8](https://fishpi.cn/article/1636516552191)
- [TypeScript SDK 2026.01.09](https://github.com/FishPiOffical/fishpi.js/commit/3980e4768afba4d181781d45078e6ca629244af1)

## 贡献

欢迎提交Issue和Pull Request！
