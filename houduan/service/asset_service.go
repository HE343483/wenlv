package service

import (
	"errors"
	"strings"
	"time"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// 资产/打卡相关业务错误。
var (
	ErrInvalidTarget    = errors.New("invalid target type")
	ErrInvalidInput     = errors.New("invalid input")
	ErrDuplicatedCheckIn = errors.New("already checked in")
)

// parseVisitedAt 解析打卡时间字符串,失败或为空时使用当前时间。
func parseVisitedAt(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	for _, layout := range []string{
		time.RFC3339, "2006-01-02 15:04:05", "2006-01-02",
	} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t
		}
	}
	return time.Now()
}

// 业务层统一的"资源不存在"错误,便于 handler 映射 404。
var ErrNotFound = errors.New("resource not found")

// validTargetTypes 收藏/照片/评价允许绑定到的目标类型。
var validTargetTypes = map[string]bool{
	model.TargetScenic: true,
	model.TargetFood:   true,
	model.TargetRoute:  true,
}

// CheckTargetType 校验目标类型是否合法。
func CheckTargetType(t string) bool { return validTargetTypes[t] }

// FavoriteService 收藏业务。
type FavoriteService struct {
	repo *repository.FavoriteRepo
}

// NewFavoriteService 构造收藏服务。
func NewFavoriteService(repo *repository.FavoriteRepo) *FavoriteService {
	return &FavoriteService{repo: repo}
}

// Add 添加收藏。
func (s *FavoriteService) Add(userID uint, targetType string, targetID uint) error {
	if !CheckTargetType(targetType) {
		return ErrInvalidTarget
	}
	err := s.repo.Add(userID, targetType, targetID)
	if errors.Is(err, repository.ErrDuplicate) {
		return nil // 幂等:已收藏视为成功
	}
	return err
}

// Remove 取消收藏。
func (s *FavoriteService) Remove(userID uint, targetType string, targetID uint) error {
	if !CheckTargetType(targetType) {
		return ErrInvalidTarget
	}
	return s.repo.Remove(userID, targetType, targetID)
}

// List 查询用户收藏,返回收藏记录(含目标 ID)。
func (s *FavoriteService) List(userID uint, targetType string) ([]model.Favorite, error) {
	if targetType != "" && !CheckTargetType(targetType) {
		return nil, ErrInvalidTarget
	}
	return s.repo.ListByUser(userID, targetType)
}

// CheckInService 打卡业务。
type CheckInService struct {
	repo *repository.CheckInRepo
}

// NewCheckInService 构造打卡服务。
func NewCheckInService(repo *repository.CheckInRepo) *CheckInService {
	return &CheckInService{repo: repo}
}

// CheckIn 打卡指定景点。
func (s *CheckInService) CheckIn(userID, scenicID uint, photoURL, comment, visitedAt string) error {
	comment = strings.TrimSpace(comment)
	if scenicID == 0 {
		return ErrInvalidInput
	}
	c := &model.CheckIn{
		UserID:    userID,
		ScenicID:  scenicID,
		PhotoURL:  photoURL,
		Comment:   comment,
		VisitedAt: parseVisitedAt(visitedAt),
	}
	err := s.repo.Add(c)
	if errors.Is(err, repository.ErrDuplicate) {
		return ErrDuplicatedCheckIn
	}
	return err
}

// List 查询用户打卡记录。
func (s *CheckInService) List(userID uint) ([]model.CheckIn, error) {
	return s.repo.ListByUser(userID)
}

// Remove 删除打卡记录。
func (s *CheckInService) Remove(userID, checkinID uint) error {
	return s.repo.Remove(userID, checkinID)
}