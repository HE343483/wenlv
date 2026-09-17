// Package service 承载业务逻辑。
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
)

// TokenPair 一次认证返回的 Token 组。
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access 过期秒数
}

// Redis Key 前缀,用于 token 有效期控制。
const (
	keyAccessJTI  = "auth:access:%s"  // access jti → uid
	keyRefreshJTI = "auth:refresh:%s" // refresh jti → uid
)

// ErrCredential 用户名或密码错误。
var ErrCredential = errors.New("用户名或密码错误")

// ErrUsernameTaken 用户名已存在。
var ErrUsernameTaken = errors.New("用户名已被占用")

// ErrRefreshInvalid 刷新 Token 无效或已过期。
var ErrRefreshInvalid = errors.New("刷新令牌无效或已过期")

// AuthService 认证业务。
type AuthService struct {
	repo *repository.UserRepo
	rdb  *redis.Client
	jwt  *pkg.JWTManager
	ttl  struct {
		accessTTL  time.Duration
		refreshTTL time.Duration
	}
}

// NewAuthService 构造认证服务。
func NewAuthService(repo *repository.UserRepo, rdb *redis.Client, jwt *pkg.JWTManager, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{repo: repo, rdb: rdb, jwt: jwt,
		ttl: struct {
			accessTTL  time.Duration
			refreshTTL time.Duration
		}{accessTTL: accessTTL, refreshTTL: refreshTTL},
	}
}

// Register 注册新用户。
func (s *AuthService) Register(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return errors.New("用户名和密码不能为空")
	}
	if _, err := s.repo.FindByUsername(username); err == nil {
		return ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return err
	}
	hash, err := pkg.HashPassword(password)
	if err != nil {
		return err
	}
	u := &model.User{Username: username, PasswordHash: hash}
	return s.repo.Create(u)
}

// Login 校验账号并签发双 Token,同时写入 Redis 控制有效期。
func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, *TokenPair, error) {
	u, err := s.repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrCredential
		}
		return nil, nil, err
	}
	if !pkg.CheckPassword(u.PasswordHash, password) {
		return nil, nil, ErrCredential
	}
	pair, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

// Refresh 用 refresh Token 换取新 Token 对(轮换)。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.Validate(refreshToken, string(pkg.TokenRefresh))
	if err != nil {
		return nil, ErrRefreshInvalid
	}
	// 校验 Redis 中的 refresh 会话仍存在
	jtiKey := fmt.Sprintf(keyRefreshJTI, claims.ID)
	uid, err := s.rdb.Get(ctx, jtiKey).Result()
	if err != nil || uid != strconv.FormatUint(uint64(claims.UID), 10) {
		return nil, ErrRefreshInvalid
	}
	// 会话轮换:删除旧 refresh,再签发新 Token
	if err := s.revokeRefresh(ctx, claims.ID); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, claims.UID)
}

// Logout 登出:注销 access 与 refresh 会话。
func (s *AuthService) Logout(ctx context.Context, accessJTI, refreshJTI string) error {
	if accessJTI != "" {
		if err := s.rdb.Del(ctx, fmt.Sprintf(keyAccessJTI, accessJTI)).Err(); err != nil {
			return err
		}
	}
	if refreshJTI != "" {
		if err := s.rdb.Del(ctx, fmt.Sprintf(keyRefreshJTI, refreshJTI)).Err(); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAccess 校验 access Token 与 Redis 会话,注入 userId。
func (s *AuthService) ValidateAccess(ctx context.Context, accessToken string) (uint, error) {
	claims, err := s.jwt.Validate(accessToken, string(pkg.TokenAccess))
	if err != nil {
		return 0, err
	}
	uid, err := s.rdb.Get(ctx, fmt.Sprintf(keyAccessJTI, claims.ID)).Result()
	if err != nil || uid != strconv.FormatUint(uint64(claims.UID), 10) {
		return 0, pkg.ErrInvalidToken
	}
	return claims.UID, nil
}

// GetUser 按 ID 查询用户。
func (s *AuthService) GetUser(ctx context.Context, uid uint) (*model.User, error) {
	return s.repo.FindByID(uid)
}

// ParseAccessJTI 仅解析 access Token 的 jti,用于登出,不校验会话。
func (s *AuthService) ParseAccessJTI(accessToken string) (string, error) {
	claims, err := s.jwt.Validate(accessToken, string(pkg.TokenAccess))
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// ParseRefreshJTI 仅解析 refresh Token 的 jti,用于登出。
func (s *AuthService) ParseRefreshJTI(refreshToken string) (string, error) {
	claims, err := s.jwt.Validate(refreshToken, string(pkg.TokenRefresh))
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// issueTokens 生成双 Token 并写入 Redis。
func (s *AuthService) issueTokens(ctx context.Context, uid uint) (*TokenPair, error) {
	access, accessJTI, err := s.jwt.GenerateAccess(uid)
	if err != nil {
		return nil, err
	}
	refresh, refreshJTI, err := s.jwt.GenerateRefresh(uid)
	if err != nil {
		return nil, err
	}
	uidStr := strconv.FormatUint(uint64(uid), 10)
	if err := s.rdb.Set(ctx, fmt.Sprintf(keyAccessJTI, accessJTI), uidStr, s.ttl.accessTTL).Err(); err != nil {
		return nil, err
	}
	if err := s.rdb.Set(ctx, fmt.Sprintf(keyRefreshJTI, refreshJTI), uidStr, s.ttl.refreshTTL).Err(); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.ttl.accessTTL.Seconds()),
	}, nil
}

func (s *AuthService) revokeRefresh(ctx context.Context, jti string) error {
	return s.rdb.Del(ctx, fmt.Sprintf(keyRefreshJTI, jti)).Err()
}