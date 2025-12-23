// Package service 业务逻辑层
package service

import (
	"context"
	"fmt"

	"video-platform/biz/dal"
	"video-platform/biz/dal/db"
	"video-platform/biz/dal/model"
	"video-platform/pkg/auth"
)

// UserService 用户服务
type UserService struct {
	store *dal.Store
}

// NewUserService 创建用户服务实例
func NewUserService(store *dal.Store) *UserService {
	return &UserService{store: store}
}

// RegisterResult 注册结果
type RegisterResult struct {
	User *model.User
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, username, password string) (*RegisterResult, error) {
	// 检查用户名是否已存在
	exists, err := db.UserExists(s.store, username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// 密码哈希
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		Username: username,
		Password: hashedPassword,
	}
	if err := db.CreateUser(s.store, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &RegisterResult{User: user}, nil
}

// LoginResult 登录结果
type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	// 查找用户
	user, err := db.GetUserByUsername(s.store, username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 验证密码
	if !auth.CheckPassword(password, user.Password) {
		return nil, ErrPasswordWrong
	}

	// 生成 JWT Token 对
	jwtMgr := auth.GetJWTManager()
	accessToken, refreshToken, err := jwtMgr.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshTokenResult 刷新令牌结果
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResult, error) {
	jwtMgr := auth.GetJWTManager()
	newAccessToken, newRefreshToken, err := jwtMgr.RefreshTokens(refreshToken)
	if err != nil {
		if err == auth.ErrTokenExpired {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	return &RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	user, err := db.GetUserByID(s.store, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(ctx context.Context, userID uint, avatarURL string) (*model.User, error) {
	if err := db.UpdateUserAvatar(s.store, userID, avatarURL); err != nil {
		return nil, fmt.Errorf("更新头像失败: %w", err)
	}

	user, err := db.GetUserByID(s.store, userID)
	if err != nil {
		return nil, fmt.Errorf("查询更新后的用户信息失败: %w", err)
	}
	return user, nil
}
