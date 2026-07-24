package strategy

import (
	"time"

	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AccessTokenOptions 生成 access_token 的参数
type AccessTokenOptions struct {
	UserID       int64
	TenantID     int64
	Roles        []int64
	UserPool     string
	TokenVersion int
}

// GeneratePlatformToken 生成 platform_token
func GeneratePlatformToken(cfg *config.Config, userID int64, tenants []TenantInfo) (string, error) {
	now := time.Now()
	claims := &PlatformClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "platform",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.PlatformTokenTTL)),
		},
		UserID:  userID,
		Tenants: tenants,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateAccessToken 生成 access_token
func GenerateAccessToken(cfg *config.Config, opts *AccessTokenOptions) (string, error) {
	now := time.Now()
	claims := &AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "access",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.AccessTokenTTL)),
		},
		UserID:       opts.UserID,
		TenantID:     opts.TenantID,
		Roles:        opts.Roles,
		UserPool:     opts.UserPool,
		TokenVersion: opts.TokenVersion,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParsePlatformToken 解析 platform_token
func ParsePlatformToken(cfg *config.Config, tokenStr string) (*PlatformClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &PlatformClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "platform_token 无效或已过期")
	}

	claims, ok := token.Claims.(*PlatformClaims)
	if !ok || !token.Valid {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "platform_token 无效")
	}

	return claims, nil
}

// ParseAccessToken 解析 access_token
// 旧 token 不含 UserPool/TokenVersion 字段时，解析为零值（UserPool=""，TokenVersion=0），不报错
func ParseAccessToken(cfg *config.Config, tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "access_token 无效或已过期")
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "access_token 无效")
	}

	return claims, nil
}
