package service

import (
	"testing"
	"time"

	"go-gin-crud/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveRangeDefaultsToRecentDays 沒給區間時預設最近 30 天
func TestResolveRangeDefaultsToRecentDays(t *testing.T) {
	from, to, fromStr, toStr, err := resolveRange(dto.ReportDailyRequest{})
	require.NoError(t, err)

	// to 是隔天零點的半開上界，所以整段長度是 defaultDailyDays 天
	assert.Equal(t, float64(defaultDailyDays), to.Sub(from).Hours()/24)
	assert.Equal(t, time.Now().Format(dateLayout), toStr, "未指定 to 時應為今天")
	assert.NotEmpty(t, fromStr)
}

// TestResolveRangeUpperBoundIsExclusiveNextDay to 必須是「隔天零點」。
//
// 若上界用 to 當天的 00:00:00，該日全天的資料都會被漏掉 —— 這是
// 半開區間最容易寫錯的地方。
func TestResolveRangeUpperBoundIsExclusiveNextDay(t *testing.T) {
	_, to, _, toStr, err := resolveRange(dto.ReportDailyRequest{
		From: "2026-01-01", To: "2026-01-31",
	})
	require.NoError(t, err)

	assert.Equal(t, "2026-01-31", toStr, "對外顯示的仍是使用者給的那天")
	assert.Equal(t, "2026-02-01", to.Format(dateLayout), "查詢用的上界應為隔天零點")
	assert.Equal(t, 0, to.Hour())
}

// TestResolveRangeSwapsReversedInput from 比 to 晚時自動對調，而不是回傳空結果
func TestResolveRangeSwapsReversedInput(t *testing.T) {
	from, _, fromStr, toStr, err := resolveRange(dto.ReportDailyRequest{
		From: "2026-03-31", To: "2026-03-01",
	})
	require.NoError(t, err)

	assert.Equal(t, "2026-03-01", fromStr)
	assert.Equal(t, "2026-03-31", toStr)
	assert.Equal(t, "2026-03-01", from.Format(dateLayout))
}

// TestResolveRangeClampsToMaxDays 區間過長時夾住。
//
// 沒有上限的話，一個 from=1970-01-01 的請求就會掃全表並產生上萬列回應。
func TestResolveRangeClampsToMaxDays(t *testing.T) {
	from, to, _, _, err := resolveRange(dto.ReportDailyRequest{
		From: "1970-01-01", To: "2026-01-01",
	})
	require.NoError(t, err)

	days := to.Sub(from).Hours() / 24
	assert.LessOrEqual(t, days, float64(maxDailyDays+1), "應被夾在 maxDailyDays 內")

	// 夾的是 from 那一端，to 不動
	assert.Equal(t, "2026-01-02", to.Format(dateLayout), "上界仍是使用者給的 to 的隔天")
	assert.Equal(t, "2025-01-01", from.Format(dateLayout), "from 應被推到 to 往前 maxDailyDays 天")
}

// TestResolveRangeRejectsBadFormat 格式錯誤要回錯，不是靜默當成預設值
func TestResolveRangeRejectsBadFormat(t *testing.T) {
	_, _, _, _, err := resolveRange(dto.ReportDailyRequest{From: "2026/01/01"})
	assert.Error(t, err)

	_, _, _, _, err = resolveRange(dto.ReportDailyRequest{To: "not-a-date"})
	assert.Error(t, err)
}

// TestResolveRangeTruncatesToMidnight 傳進來的時分秒不該影響分組邊界
func TestResolveRangeTruncatesToMidnight(t *testing.T) {
	from, _, _, _, err := resolveRange(dto.ReportDailyRequest{From: "2026-05-10", To: "2026-05-20"})
	require.NoError(t, err)

	assert.Equal(t, 0, from.Hour())
	assert.Equal(t, 0, from.Minute())
	assert.Equal(t, 0, from.Second())
}

// TestNormalizeDate 兩種驅動的日期輸出都要正規化成 YYYY-MM-DD。
//
// MySQL 的 DSN 開了 parseTime=True，DATE(created_at) 會變成完整的
// RFC3339；sqlite 直接回日期字串。對外契約只有一種格式。
func TestNormalizeDate(t *testing.T) {
	assert.Equal(t, "2026-08-05", normalizeDate("2026-08-05T00:00:00+08:00"), "MySQL + parseTime")
	assert.Equal(t, "2026-08-05", normalizeDate("2026-08-05"), "sqlite")
	assert.Equal(t, "", normalizeDate(""), "空值不該 panic")
	assert.Equal(t, "2026-08", normalizeDate("2026-08"), "比日期短的字串原樣回傳")
}
