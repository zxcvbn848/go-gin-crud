package repository

import "gorm.io/gorm"

// WhereLike 輔助函數：為查詢添加 LIKE 條件
func WhereLike(query *gorm.DB, column string, search string) *gorm.DB {
	if search == "" {
		return query
	}
	searchPattern := "%" + search + "%"
	return query.Where(column+" LIKE ?", searchPattern)
}

// WhereLikeOr 輔助函數：為查詢添加多個 LIKE 條件（OR 關係）
func WhereLikeOr(query *gorm.DB, search string, columns ...string) *gorm.DB {
	if search == "" || len(columns) == 0 {
		return query
	}

	// 第一個條件使用 WhereLike
	query = WhereLike(query, columns[0], search)

	// 其餘條件使用 Or，重用 WhereLike 的邏輯
	searchPattern := "%" + search + "%"
	for i := 1; i < len(columns); i++ {
		query = query.Or(columns[i]+" LIKE ?", searchPattern)
	}

	return query
}
