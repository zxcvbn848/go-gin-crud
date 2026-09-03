package dto

// ReportOverviewResponse 總覽：各表筆數與商品庫存總值
type ReportOverviewResponse struct {
	Users             int64   `json:"users"`
	Posts             int64   `json:"posts"`
	Products          int64   `json:"products"`
	Books             int64   `json:"books"`
	ProductStockValue float64 `json:"product_stock_value"`
}

// DailyPoint 每日統計的單一資料點
type DailyPoint struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int64  `json:"count"`
}

// ReportDailyRequest 每日統計的查詢區間。
//
// 皆為選填：預設為「最近 30 天」。
type ReportDailyRequest struct {
	From string `form:"from" binding:"omitempty,datetime=2006-01-02"`
	To   string `form:"to" binding:"omitempty,datetime=2006-01-02"`
}

// ReportDailyResponse 每日新增數，三個表各一組序列
type ReportDailyResponse struct {
	From     string       `json:"from"`
	To       string       `json:"to"`
	Users    []DailyPoint `json:"users"`
	Posts    []DailyPoint `json:"posts"`
	Products []DailyPoint `json:"products"`
}

// ReportAuthorsRequest 作者排行的查詢條件
type ReportAuthorsRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// AuthorRankItem 作者排行的單一列
type AuthorRankItem struct {
	AuthorID uint   `json:"author_id"`
	Email    string `json:"email"`
	Posts    int64  `json:"posts"`
}

// ReportAuthorsResponse 發文數排行
type ReportAuthorsResponse struct {
	Limit   int              `json:"limit"`
	Authors []AuthorRankItem `json:"authors"`
}
