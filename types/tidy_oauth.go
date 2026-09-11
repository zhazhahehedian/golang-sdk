package types

import "encoding/json"

// OAuthTokens OAuth 授权或续签结果。续签后必须保存新的 RefreshToken。
type OAuthTokens struct {
	TokenType        string `json:"token_type"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

// OpenIDVerifyResult 已通过服务端校验的用户标识及可选 OAuth 令牌。
type OpenIDVerifyResult struct {
	OpenID string `json:"openid"`
	OAuthTokens
}

type OAuthProfile struct {
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserNickname  string `json:"userNickname"`
	UserAvatarURL string `json:"userAvatarURL"`
}

// OAuthUserDetail 不复用普通用户模型，避免把不同鉴权接口的字段混在一起。
type OAuthUserDetail struct {
	OId            string `json:"oId"`
	UserName       string `json:"userName"`
	UserNickname   string `json:"userNickname"`
	UserAvatarURL  string `json:"userAvatarURL"`
	UserOnlineFlag bool   `json:"userOnlineFlag"`
	OnlineMinute   int64  `json:"onlineMinute"`
	UserURL        string `json:"userURL"`
	UserCity       string `json:"userCity,omitempty"`
	UserPoint      int64  `json:"userPoint"`
	UserIntro      string `json:"userIntro"`
	// 接入文档示例使用字符串，API 表格使用数字；兼容两种 JSON 表示。
	UserNo             json.Number     `json:"userNo"`
	UserAppRole        json.Number     `json:"userAppRole"`
	SysMetal           json.RawMessage `json:"sysMetal"`
	FollowerCount      int             `json:"followerCount"`
	FollowingUserCount int             `json:"followingUserCount"`
	UserRole           string          `json:"userRole"`
	CardBg             string          `json:"cardBg"`
}

type OAuthMembership struct {
	Active     bool   `json:"active"`
	State      int    `json:"state"`
	LvCode     string `json:"lvCode"`
	ExpiresAt  int64  `json:"expiresAt"`
	ConfigJSON string `json:"configJson"`
}

type OAuthPagination struct {
	PaginationCurrentPageNum int   `json:"paginationCurrentPageNum"`
	PaginationPageSize       int   `json:"paginationPageSize"`
	PaginationRecordCount    int   `json:"paginationRecordCount"`
	PaginationPageCount      int   `json:"paginationPageCount"`
	PaginationPageNums       []int `json:"paginationPageNums"`
}

type OAuthPointRecord struct {
	OId           string `json:"oId"`
	FromID        string `json:"fromId"`
	ToID          string `json:"toId"`
	Sum           int64  `json:"sum"`
	Type          int    `json:"type"`
	Time          int64  `json:"time"`
	DataID        string `json:"dataId"`
	Memo          string `json:"memo"`
	SourceAppName string `json:"sourceAppName"`
	SourceScene   string `json:"sourceScene"`
	Operation     string `json:"operation"`
	Balance       int64  `json:"balance"`
	DisplayType   string `json:"displayType"`
	Description   string `json:"description"`
}

type OAuthPoints struct {
	UserID     string             `json:"userId"`
	UserPoint  int64              `json:"userPoint"`
	Records    []OAuthPointRecord `json:"records"`
	Pagination OAuthPagination    `json:"pagination"`
}

type OAuthArticle struct {
	OId                 string `json:"oId"`
	ArticleTitle        string `json:"articleTitle"`
	ArticlePermalink    string `json:"articlePermalink"`
	ArticleTags         string `json:"articleTags"`
	ArticleCreateTime   int64  `json:"articleCreateTime"`
	ArticleUpdateTime   int64  `json:"articleUpdateTime"`
	ArticleCommentCount int    `json:"articleCommentCount"`
	ArticleViewCount    int    `json:"articleViewCount"`
	ArticleType         int    `json:"articleType"`
	ArticlePerfect      int    `json:"articlePerfect"`
}

type OAuthArticles struct {
	UserID     string          `json:"userId"`
	Articles   []OAuthArticle  `json:"articles"`
	Pagination OAuthPagination `json:"pagination"`
}
