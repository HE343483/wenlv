// Package repository 提供数据访问层。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"wenlv-backend/model"
)

// ErrUserNotFound 用户不存在。
var ErrUserNotFound = errors.New("user not found")

// UserRepo 用户数据访问。
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 构造用户仓储。
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 新建用户,返回含自增 ID 的实体。
func (r *UserRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}

// FindByUsername 按用户名查询。
func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 按主键查询。
func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}