package types

type ColumnCover struct {
	OId            string `json:"oId"`
	ColumnTitle    string `json:"columnTitle"`
	ColumnCoverURL string `json:"columnCoverURL"`
	ColumnHasCover bool   `json:"columnHasCover"`
}

type ColumnCoverData struct {
	Column *ColumnCover `json:"column"`
}
