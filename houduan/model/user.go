package model

import "time"

// User 用户账号。
type User struct {
	ID           uint      `gorm:"primaryKey;comment:用户ID" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null;comment:用户名(唯一)" json:"username"`
	PasswordHash string    `gorm:"size:255;not null;comment:密码哈希(bcrypt)" json:"-"`
	AvatarURL    string    `gorm:"size:512;comment:头像URL" json:"avatar_url"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
