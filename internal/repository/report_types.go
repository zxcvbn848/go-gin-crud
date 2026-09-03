package repository

// 報表用的聚合結果型別。
//
// 這些不是 model —— 它們是查詢的產物，沒有對應的表。放在 repository 層是因為
// 欄位名要跟 SQL 的 alias 對上（GORM 靠這個掃描結果）。
//
// 刻意不建 ReportRepository：報表的查詢天生跨多個 model，集中成一個 repo
// 會長成什麼都碰的 god object。這裡的做法是查詢歸各自的 model repo，
// 跨表的組裝歸 ReportService。

// DailyCount 每日計數，Date 為 YYYY-MM-DD
type DailyCount struct {
	Date  string `gorm:"column:date"`
	Count int64  `gorm:"column:count"`
}

// AuthorCount 作者發文數
type AuthorCount struct {
	AuthorID uint   `gorm:"column:author_id"`
	Email    string `gorm:"column:email"`
	Count    int64  `gorm:"column:count"`
}
