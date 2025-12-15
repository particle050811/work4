package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrTokenExpired Token 已过期
	ErrTokenExpired = errors.New("token 已过期")
	// ErrTokenInvalid Token 无效
	ErrTokenInvalid = errors.New("token 无效")
	// ErrTokenMalformed Token 格式错误
	ErrTokenMalformed = errors.New("token 格式错误")
)

// TokenType Token 类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims 自定义 JWT Claims
type Claims struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secretKey           []byte
	accessTokenExpiry   time.Duration
	refreshTokenExpiry  time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager() *JWTManager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET 环境变量未设置，请检查 .env 文件")
	}

	return &JWTManager{
		secretKey:          []byte(secret),
		accessTokenExpiry:  15 * time.Minute, // 访问令牌 15 分钟
		refreshTokenExpiry: 7 * 24 * time.Hour, // 刷新令牌 7 天
	}
}

// GenerateTokenPair 生成 access_token 和 refresh_token 对
func (m *JWTManager) GenerateTokenPair(userID uint, username string) (accessToken, refreshToken string, err error) {
	accessToken, err = m.generateToken(userID, username, AccessToken, m.accessTokenExpiry)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = m.generateToken(userID, username, RefreshToken, m.refreshTokenExpiry)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateToken 生成指定类型的 Token
func (m *JWTManager) generateToken(userID uint, username string, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "fanone-video-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ParseToken 解析 Token
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ValidateAccessToken 验证访问令牌
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateRefreshToken 验证刷新令牌
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// RefreshTokens 使用刷新令牌获取新的令牌对
func (m *JWTManager) RefreshTokens(refreshTokenString string) (newAccessToken, newRefreshToken string, err error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	return m.GenerateTokenPair(claims.UserID, claims.Username)
}

// 全局 JWT 管理器实例
var jwtManager *JWTManager

// GetJWTManager 获取 JWT 管理器实例（单例）
func GetJWTManager() *JWTManager {
	if jwtManager == nil {
		jwtManager = NewJWTManager()
	}
	return jwtManager
}
