package service

import (
	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// PoemService 景点诗词服务(诗词地图)。
type PoemService struct {
	repo *repository.PoemRepo
}

// NewPoemService 构造诗词服务。
func NewPoemService(repo *repository.PoemRepo) *PoemService {
	return &PoemService{repo: repo}
}

// ListBySpot 查询景点关联诗词(含全部多语字段)。
func (s *PoemService) ListBySpot(spotID uint) ([]model.Poem, error) {
	return s.repo.ListBySpot(spotID)
}
