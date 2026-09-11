package types

// PostArticleDraftRequest 保存完整草稿内容。字符串空值会发送，可用于清空内容。
// 可选数字和布尔值使用指针：nil 不发送，指向 0/false 则显式发送。
type PostArticleDraftRequest struct {
	ArticleDraftID             string `json:"articleDraftId,omitempty"`
	ArticleTitle               string `json:"articleTitle"`
	ArticleContent             string `json:"articleContent"`
	ArticleDraftThoughtContent string `json:"articleDraftThoughtContent"`
	ArticleTags                string `json:"articleTags"`
	ArticleType                *int   `json:"articleType,omitempty"` // 6 表示长文章
	ColumnID                   string `json:"columnId"`
	ColumnTitle                string `json:"columnTitle"`
	ChapterNo                  string `json:"chapterNo"`
	ArticleRewardContent       string `json:"articleRewardContent"`
	ArticleRewardPoint         *int   `json:"articleRewardPoint,omitempty"`
	ArticleQnAOfferPoint       *int   `json:"articleQnAOfferPoint,omitempty"`
	ArticleCommentable         *bool  `json:"articleCommentable,omitempty"`
	ArticleAnonymous           *bool  `json:"articleAnonymous,omitempty"`
	ArticleNotifyFollowers     *bool  `json:"articleNotifyFollowers,omitempty"`
	ArticleShowInList          *int   `json:"articleShowInList,omitempty"`
	ArticleStatement           *int   `json:"articleStatement,omitempty"`
}

// ArticleDraft 列表和保存响应不包含正文；详情响应才包含三个正文字段。
type ArticleDraft struct {
	OId                         string `json:"oId"`
	ArticleDraftTitle           string `json:"articleDraftTitle"`
	ArticleDraftSummary         string `json:"articleDraftSummary"`
	ArticleDraftContent         string `json:"articleDraftContent,omitempty"`
	ArticleDraftThoughtContent  string `json:"articleDraftThoughtContent,omitempty"`
	ArticleDraftTags            string `json:"articleDraftTags"`
	ArticleDraftType            int    `json:"articleDraftType"`
	ArticleDraftColumnID        string `json:"articleDraftColumnId"`
	ArticleDraftColumnTitle     string `json:"articleDraftColumnTitle"`
	ArticleDraftChapterNo       string `json:"articleDraftChapterNo"`
	ArticleDraftRewardContent   string `json:"articleDraftRewardContent,omitempty"`
	ArticleDraftRewardPoint     int    `json:"articleDraftRewardPoint"`
	ArticleDraftQnAOfferPoint   int    `json:"articleDraftQnAOfferPoint"`
	ArticleDraftCommentable     bool   `json:"articleDraftCommentable"`
	ArticleDraftAnonymous       bool   `json:"articleDraftAnonymous"`
	ArticleDraftNotifyFollowers bool   `json:"articleDraftNotifyFollowers"`
	ArticleDraftShowInList      int    `json:"articleDraftShowInList"`
	ArticleDraftStatement       int    `json:"articleDraftStatement"`
	ArticleDraftUpdatedTime     int64  `json:"articleDraftUpdatedTime"`
}

type ArticleDraftListData struct {
	Drafts []ArticleDraft `json:"drafts"`
}

type ArticleDraftData struct {
	Draft *ArticleDraft `json:"draft"`
}

type DeleteArticleDraftData struct {
	ID string `json:"id"`
}
