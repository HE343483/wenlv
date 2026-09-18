package service

import (
	"errors"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// ScenicService 景点查询业务。
type ScenicService struct {
	repo *repository.ScenicRepo
}

// NewScenicService 构造景点服务。
func NewScenicService(repo *repository.ScenicRepo) *ScenicService {
	return &ScenicService{repo: repo}
}

// List 分页查询景点。
func (s *ScenicService) List(opts repository.QueryOptions) ([]model.ScenicSpot, int64, error) {
	return s.repo.List(opts)
}

// Get 查询景点详情。
func (s *ScenicService) Get(id uint) (*model.ScenicSpot, error) {
	v, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return v, err
}

// FoodService 美食查询业务。
type FoodService struct {
	repo *repository.FoodRepo
}

// NewFoodService 构造美食服务。
func NewFoodService(repo *repository.FoodRepo) *FoodService {
	return &FoodService{repo: repo}
}

// List 分页查询美食。
func (s *FoodService) List(opts repository.QueryOptions) ([]model.Food, int64, error) {
	return s.repo.List(opts)
}

// Get 查询美食详情。
func (s *FoodService) Get(id uint) (*model.Food, error) {
	v, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return v, err
}

// RouteService 路线查询业务。
type RouteService struct {
	repo *repository.RouteRepo
}

// NewRouteService 构造路线服务。
func NewRouteService(repo *repository.RouteRepo) *RouteService {
	return &RouteService{repo: repo}
}

// List 分页查询路线。
func (s *RouteService) List(opts repository.QueryOptions) ([]model.Route, int64, error) {
	return s.repo.List(opts)
}

// Get 查询路线详情。
func (s *RouteService) Get(id uint) (*model.Route, error) {
	v, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return v, err
}