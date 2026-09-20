package service

import (
	"time"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// CultureDailyService 每日蜀签展示服务。
// 轮换规则:取第 (dayOfYear + offset) % 总数 条(id 升序为固定顺序),
// 每天自动轮换一条;前端"换一条"按钮 offset+1 顺延取下一条,不与日期严格一一对应。
type CultureDailyService struct {
	repo *repository.CultureDailyRepo
}

// NewCultureDailyService 构造蜀签展示服务。
func NewCultureDailyService(repo *repository.CultureDailyRepo) *CultureDailyService {
	return &CultureDailyService{repo: repo}
}

// Daily 取当前应展示的蜀签;库内无数据时返回 nil(前端据此整体隐藏卡片)。
func (s *CultureDailyService) Daily(offset int) (*model.CultureDaily, error) {
	total, err := s.repo.Count()
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return nil, nil
	}
	idx := modIndex(int(time.Now().Local().YearDay())+offset, int(total))
	return s.repo.FindByIndex(idx)
}

// modIndex 非负取模(兼容负偏移,保证索引落在 [0, total) 内)。
func modIndex(v, total int) int {
	if total <= 0 {
		return 0
	}
	return ((v % total) + total) % total
}
