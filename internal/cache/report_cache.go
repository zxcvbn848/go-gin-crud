package cache

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"go-gin-crud/internal/dto"
	redispkg "go-gin-crud/internal/redis"

	"github.com/redis/go-redis/v9"
)

const (
	reportKeyPrefix = "report:"
	// reportTTL 報表允許的資料落後時間。
	//
	// 報表本來就是回顧性的，看到五分鐘前的數字沒有問題；而查詢成本高
	// （最慢的一支 500ms 以上），快取的效益比一般 CRUD 大得多。
	reportTTL = 5 * time.Minute
)

// ReportCache 報表快取。
//
// 三支報表的回應型別不同，用泛型的 get/set 收斂重複的序列化程式碼。
type ReportCache interface {
	GetOverview(ctx context.Context) (*dto.ReportOverviewResponse, error)
	SetOverview(ctx context.Context, v *dto.ReportOverviewResponse) error

	GetDaily(ctx context.Context, from, to string) (*dto.ReportDailyResponse, error)
	SetDaily(ctx context.Context, from, to string, v *dto.ReportDailyResponse) error

	GetAuthors(ctx context.Context, limit int) (*dto.ReportAuthorsResponse, error)
	SetAuthors(ctx context.Context, limit int, v *dto.ReportAuthorsResponse) error
}

type redisReportCache struct {
	client *redispkg.Client
}

func NewReportCache(client *redispkg.Client) ReportCache {
	if client == nil {
		return nil
	}
	return &redisReportCache{client: client}
}

// getJSON 讀取並反序列化。
//
// 快取未命中（redis.Nil）回傳 (nil, nil) —— 那不是錯誤，呼叫方直接去查 DB。
// 其他錯誤（含熔斷中的 ErrBreakerOpen）照實回報，service 層一樣會降級查 DB。
func getJSON[T any](ctx context.Context, c *redisReportCache, key string) (*T, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func setJSON[T any](ctx context.Context, c *redisReportCache, key string, v *T) error {
	if c == nil || c.client == nil || v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, raw, reportTTL).Err()
}

func (c *redisReportCache) GetOverview(ctx context.Context) (*dto.ReportOverviewResponse, error) {
	return getJSON[dto.ReportOverviewResponse](ctx, c, reportKeyPrefix+"overview")
}

func (c *redisReportCache) SetOverview(ctx context.Context, v *dto.ReportOverviewResponse) error {
	return setJSON(ctx, c, reportKeyPrefix+"overview", v)
}

// dailyKey 區間是查詢的一部分，必須進 key —— 否則不同區間會互相汙染
func dailyKey(from, to string) string {
	return reportKeyPrefix + "daily:" + from + ":" + to
}

func (c *redisReportCache) GetDaily(ctx context.Context, from, to string) (*dto.ReportDailyResponse, error) {
	return getJSON[dto.ReportDailyResponse](ctx, c, dailyKey(from, to))
}

func (c *redisReportCache) SetDaily(ctx context.Context, from, to string, v *dto.ReportDailyResponse) error {
	return setJSON(ctx, c, dailyKey(from, to), v)
}

func authorsKey(limit int) string {
	return reportKeyPrefix + "authors:" + strconv.Itoa(limit)
}

func (c *redisReportCache) GetAuthors(ctx context.Context, limit int) (*dto.ReportAuthorsResponse, error) {
	return getJSON[dto.ReportAuthorsResponse](ctx, c, authorsKey(limit))
}

func (c *redisReportCache) SetAuthors(ctx context.Context, limit int, v *dto.ReportAuthorsResponse) error {
	return setJSON(ctx, c, authorsKey(limit), v)
}
