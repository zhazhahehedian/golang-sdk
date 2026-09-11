package types

type RepeaterItem struct {
	OId                        string `json:"oId"`
	RepeaterContentType        string `json:"repeaterContentType"`
	RepeaterContentTypeLabel   string `json:"repeaterContentTypeLabel"`
	RepeaterContent            string `json:"repeaterContent"`
	RepeaterContentAuthorID    string `json:"repeaterContentAuthorId"`
	RepeaterContentAuthorName  string `json:"repeaterContentAuthorName"`
	RepeaterContentSource      string `json:"repeaterContentSource"`
	RepeaterContentLikeCount   int    `json:"repeaterContentLikeCount"`
	RepeaterContentLiked       bool   `json:"repeaterContentLiked"`
	RepeaterContentCreatedTime int64  `json:"repeaterContentCreatedTime"`
	RepeaterContentUpdatedTime int64  `json:"repeaterContentUpdatedTime"`
}

type RepeaterItemsData struct {
	Items []RepeaterItem `json:"items"`
}

type RepeaterItemData struct {
	Item *RepeaterItem `json:"item"`
}

type RepeaterLikeData struct {
	Liked                    bool `json:"liked"`
	RepeaterContentLikeCount int  `json:"repeaterContentLikeCount"`
}
