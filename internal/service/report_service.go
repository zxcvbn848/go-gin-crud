package service

import (
	"context"
	"time"

	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/repository"
)

const (
	// dateLayout 報表對外的日期格式
	dateLayout = "2006-01-02"
	// defaultDailyDays 未指定區間時的預設天數
	defaultDailyDays = 30
	// maxDailyDays 區間上限。
	//
	// 沒有上限的話，一個 from=1970-01-01 的請求就會掃全表並產生上萬列回應。
	maxDailyDays = 365
	// defaultAuthorLimit 作者排行未指定時的筆數
	defaultAuthorLimit = 10
)

type ReportService interface {
	Overview(ctx context.Context) (*dto.ReportOverviewResponse, error)
	Daily(ctx context.Context, req dto.ReportDailyRequest) (*dto.ReportDailyResponse, error)
	TopAuthors(ctx context.Context, req dto.ReportAuthorsRequest) (*dto.ReportAuthorsResponse, error)
}

// reportService 組裝跨 model 的報表。
//
// 各表的查詢留在各自的 repository，這裡只負責串起來與轉成回應格式 ——
// 刻意不建 ReportRepository，那會長成什麼表都碰的 god object。
type reportService struct {
	userRepo    repository.UserRepository
	postRepo    repository.PostRepository
	productRepo repository.ProductRepository
	bookRepo    repository.BookRepository
	reportCache cache.ReportCache
}

func NewReportService(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	productRepo repository.ProductRepository,
	bookRepo repository.BookRepository,
	reportCache cache.ReportCache,
) ReportService {
	return &reportService{
		userRepo:    userRepo,
		postRepo:    postRepo,
		productRepo: productRepo,
		bookRepo:    bookRepo,
		reportCache: reportCache,
	}
}

// cacheGet 讀快取。錯誤只記錄不往上拋 —— 快取壞掉不該讓報表整支失敗，
// 這是既有的降級慣例（Redis 熔斷時 ErrBreakerOpen 也走這條路）。
func cacheLog(err error, what string) {
	if err != nil {
		logger.Log.WithError(err).Warnf("報表快取%s失敗，改查 DB", what)
	}
}

func (s *reportService) Overview(ctx context.Context) (*dto.ReportOverviewResponse, error) {
	if s.reportCache != nil {
		cached, err := s.reportCache.GetOverview(ctx)
		cacheLog(err, "讀取")
		if cached != nil {
			return cached, nil
		}
	}

	users, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	posts, err := s.postRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	products, err := s.productRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	books, err := s.bookRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	stockValue, err := s.productRepo.SumStockValue(ctx)
	if err != nil {
		return nil, err
	}

	resp := &dto.ReportOverviewResponse{
		Users:             users,
		Posts:             posts,
		Products:          products,
		Books:             books,
		ProductStockValue: stockValue,
	}

	if s.reportCache != nil {
		cacheLog(s.reportCache.SetOverview(ctx, resp), "寫入")
	}
	return resp, nil
}

// resolveRange 決定查詢區間。
//
// 回傳的 to 是「隔天零點」的半開區間上界，查詢用 created_at < to ——
// 這樣才涵蓋 to 當天的所有時刻，而不是只到 00:00:00。
func resolveRange(req dto.ReportDailyRequest) (from, to time.Time, fromStr, toStr string, err error) {
	now := time.Now()

	toDay := now
	if req.To != "" {
		toDay, err = time.Parse(dateLayout, req.To)
		if err != nil {
			return
		}
	}
	fromDay := toDay.AddDate(0, 0, -(defaultDailyDays - 1))
	if req.From != "" {
		fromDay, err = time.Parse(dateLayout, req.From)
		if err != nil {
			return
		}
	}

	// truncate 到日界線，避免傳進來的時分秒影響分組
	fromDay = time.Date(fromDay.Year(), fromDay.Month(), fromDay.Day(), 0, 0, 0, 0, now.Location())
	toDay = time.Date(toDay.Year(), toDay.Month(), toDay.Day(), 0, 0, 0, 0, now.Location())

	if toDay.Before(fromDay) {
		fromDay, toDay = toDay, fromDay
	}
	// 夾住區間長度，否則 from 給很早的日期就會掃全表
	if toDay.Sub(fromDay) > maxDailyDays*24*time.Hour {
		fromDay = toDay.AddDate(0, 0, -maxDailyDays)
	}

	return fromDay, toDay.AddDate(0, 0, 1), fromDay.Format(dateLayout), toDay.Format(dateLayout), nil
}

func (s *reportService) Daily(ctx context.Context, req dto.ReportDailyRequest) (*dto.ReportDailyResponse, error) {
	from, to, fromStr, toStr, err := resolveRange(req)
	if err != nil {
		return nil, err
	}

	if s.reportCache != nil {
		cached, err := s.reportCache.GetDaily(ctx, fromStr, toStr)
		cacheLog(err, "讀取")
		if cached != nil {
			return cached, nil
		}
	}

	users, err := s.userRepo.CountDailyCreated(ctx, from, to)
	if err != nil {
		return nil, err
	}
	posts, err := s.postRepo.CountDailyCreated(ctx, from, to)
	if err != nil {
		return nil, err
	}
	products, err := s.productRepo.CountDailyCreated(ctx, from, to)
	if err != nil {
		return nil, err
	}

	resp := &dto.ReportDailyResponse{
		From:     fromStr,
		To:       toStr,
		Users:    toPoints(users),
		Posts:    toPoints(posts),
		Products: toPoints(products),
	}

	if s.reportCache != nil {
		cacheLog(s.reportCache.SetDaily(ctx, fromStr, toStr, resp), "寫入")
	}
	return resp, nil
}

// normalizeDate 把驅動回傳的日期正規化成 YYYY-MM-DD。
//
// DSN 開了 parseTime=True，所以 MySQL 的 DATE(created_at) 會被解析成
// time.Time 再字串化，掃進 string 欄位時變成完整的 RFC3339
// （"2026-08-05T00:00:00+08:00"），而對外契約是 YYYY-MM-DD。
//
// sqlite 的 DATE() 直接回 "2026-08-05"，兩種都只取前 10 碼即可。
//
// ponytail: 用長度截斷而不是解析再格式化 —— 兩種驅動的輸出前 10 碼都
// 剛好是日期，多一層 time.Parse 只是把同一件事寫得更長。
func normalizeDate(v string) string {
	if len(v) > len(dateLayout) {
		return v[:len(dateLayout)]
	}
	return v
}

// toPoints 轉成對外格式。
//
// 沒有資料的日期不補零 —— 補零要在應用層產生整段日期序列，回應也會膨脹；
// 圖表端補比較合適，它才知道要畫多長的軸。
func toPoints(rows []repository.DailyCount) []dto.DailyPoint {
	points := make([]dto.DailyPoint, 0, len(rows))
	for _, r := range rows {
		points = append(points, dto.DailyPoint{Date: normalizeDate(r.Date), Count: r.Count})
	}
	return points
}

func (s *reportService) TopAuthors(ctx context.Context, req dto.ReportAuthorsRequest) (*dto.ReportAuthorsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultAuthorLimit
	}

	if s.reportCache != nil {
		cached, err := s.reportCache.GetAuthors(ctx, limit)
		cacheLog(err, "讀取")
		if cached != nil {
			return cached, nil
		}
	}

	rows, err := s.postRepo.TopAuthors(ctx, limit)
	if err != nil {
		return nil, err
	}

	authors := make([]dto.AuthorRankItem, 0, len(rows))
	for _, r := range rows {
		authors = append(authors, dto.AuthorRankItem{
			AuthorID: r.AuthorID,
			Email:    r.Email,
			Posts:    r.Count,
		})
	}

	resp := &dto.ReportAuthorsResponse{Limit: limit, Authors: authors}

	if s.reportCache != nil {
		cacheLog(s.reportCache.SetAuthors(ctx, limit, resp), "寫入")
	}
	return resp, nil
}
