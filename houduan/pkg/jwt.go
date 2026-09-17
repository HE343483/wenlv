package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType 区分 access / refresh 两种 Token。
type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

// ErrInvalidToken Token 无效。
var ErrInvalidToken = errors.New("invalid token")

// Claims 自定义 JWT 载荷,含用户 ID 与 Token 类型。
type Claims struct {
	UID  uint      `json:"uid"`
	Type TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTConfig 签名与有效期配置。
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     int // 秒
	RefreshTTL    int // 秒
	Issuer        string
}

// JWTManager 负责签发与解析双 Token。
type JWTManager struct {
	cfg JWTConfig
}

// NewJWTManager 构造 JWT 管理器。
func NewJWTManager(cfg JWTConfig) *JWTManager {
	return &JWTManager{cfg: cfg}
}

// newToken 生成指定类型的 Token,返回 token 串与 jti。
func (m *JWTManager) newToken(uid uint, typ TokenType) (string, string, error) {
	jti, err := randomID()
	if err != nil {
		return "", "", err
	}
	secret := m.cfg.AccessSecret
	ttl := m.cfg.AccessTTL
	if typ == TokenRefresh {
		secret = m.cfg.RefreshSecret
		ttl = m.cfg.RefreshTTL
	}
	now := time.Now()
	claims := Claims{
		UID:  uid,
		Type: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Subject:   "wenlv-user",
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
		},
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tk.SignedString([]byte(secret))
	return signed, jti, err
}

// GenerateAccess 签发 access Token。
func (m *JWTManager) GenerateAccess(uid uint) (string, string, error) {
	return m.newToken(uid, TokenAccess)
}

// GenerateRefresh 签发 refresh Token。
func (m *JWTManager) GenerateRefresh(uid uint) (string, string, error) {
	return m.newToken(uid, TokenRefresh)
}

// Validate 校验 Token 签名、有效期与类型,返回解析后的 Claims。
func (m *JWTManager) Validate(token, expectedType string) (*Claims, error) {
	secret := m.cfg.AccessSecret
	if expectedType == string(TokenRefresh) {
		secret = m.cfg.RefreshSecret
	}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(m.cfg.Issuer))
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != TokenType(expectedType) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// randomID 生成随机 jti 标识。
func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}