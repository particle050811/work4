// Package service 业务逻辑层
package service

import "errors"

// 用户模块错误
var (
	ErrUserExists    = errors.New("用户名已存在")
	ErrUserNotFound  = errors.New("用户不存在")
	ErrPasswordWrong = errors.New("密码错误")
	ErrTokenExpired  = errors.New("令牌已过期")
	ErrTokenInvalid  = errors.New("令牌无效")
)

// 视频模块错误
var (
	ErrVideoNotFound = errors.New("视频不存在")
)

// 互动模块错误
var (
	ErrCommentNotFound  = errors.New("评论不存在")
	ErrNoPermission     = errors.New("无权限操作")
	ErrCommentTooLong   = errors.New("评论内容过长")
	ErrCommentEmpty     = errors.New("评论内容不能为空")
)

// 社交模块错误
var (
	ErrCannotFollowSelf = errors.New("不能关注自己")
	ErrFollowNotFound   = errors.New("关注关系不存在")
)
