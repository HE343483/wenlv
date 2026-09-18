package pkg

import "golang.org/x/crypto/bcrypt"

// HashPassword 对明文密码做 bcrypt 哈希,返回带盐的哈希串。
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword 校验明文密码是否匹配给定哈希。
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}