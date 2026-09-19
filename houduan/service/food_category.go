package service

import (
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// FoodCategoryService 美食大类只读查询。
type FoodCategoryService struct{ repo *repository.FoodCategoryRepo }

// NewFoodCategoryService 构造美食大类服务。
func NewFoodCategoryService(repo *repository.FoodCategoryRepo) *FoodCategoryService {
	return &FoodCategoryService{repo: repo}
}

// Get 按类别键查询;不存在时返回 ErrNotFound。
func (s *FoodCategoryService) Get(key string) (*model.FoodCategory, error) {
	v, err := s.repo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrNotFound
	}
	return v, nil
}
