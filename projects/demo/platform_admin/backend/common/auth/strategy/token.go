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
	// 自定义过期时间（秒），0 表示使用全局配置
	ExpiresIn int64
}

// RefreshTokenOptions 生成 refresh_token 的参数
type RefreshTokenOptions struct {
	UserID       int64
	TenantID     int64
	UserPool     string
	TokenVersion int
	// 自定义过期时间（秒），0 表示使用全局配置
	ExpiresIn int64
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
	ttl := cfg.JWT.AccessTokenTTL
	if opts.ExpiresIn > 0 {
		ttl = time.Duration(opts.ExpiresIn) * time.Second
	}
	claims := &AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "access",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
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

// RefreshClaims refresh_token 的 JWT claims
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID       int64  `json:"user_id"`
	TenantID     int64  `json:"tenant_id"`
	UserPool     string `json:"user_pool,omitempty"`
	TokenVersion int    `json:"token_version,omitempty"`
}

// GenerateRefreshToken 生成 refresh_token
func GenerateRefreshToken(cfg *config.Config, opts *RefreshTokenOptions) (string, error) {
	now := time.Now()
	ttl := cfg.JWT.RefreshTokenTTL // 默认 7 天
	if opts.ExpiresIn > 0 {
		ttl = time.Duration(opts.ExpiresIn) * time.Second
	}
	claims := &RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "refresh",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:       opts.UserID,
		TenantID:     opts.TenantID,
		UserPool:     opts.UserPool,
		TokenVersion: opts.TokenVersion,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseRefreshToken 解析 refresh_token
func ParseRefreshToken(cfg *config.Config, tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "refresh_token 无效或已过期")
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "refresh_token 无效")
	}

	return claims, nil
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
