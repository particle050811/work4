package main

// 响应结构

type BaseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type RegisterResponse struct {
	Base BaseResponse `json:"base"`
}

type LoginResponse struct {
	Base         BaseResponse `json:"base"`
	Data         User         `json:"data"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Base         BaseResponse `json:"base"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type GetUserInfoResponse struct {
	Base BaseResponse `json:"base"`
	Data User         `json:"data"`
}

type PublishVideoResponse struct {
	Base BaseResponse `json:"base"`
}

type Video struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	VideoURL     string `json:"video_url"`
	CoverURL     string `json:"cover_url"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	VisitCount   int64  `json:"visit_count"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	DeletedAt    string `json:"deleted_at"`
}

type VideoListWithTotal struct {
	Items []Video `json:"items"`
	Total int64   `json:"total"`
}

type ListPublishedVideosResponse struct {
	Base BaseResponse       `json:"base"`
	Data VideoListWithTotal `json:"data"`
}

type SearchVideosResponse struct {
	Base BaseResponse       `json:"base"`
	Data VideoListWithTotal `json:"data"`
}

type GetHotVideosResponse struct {
	Base BaseResponse       `json:"base"`
	Data VideoListWithTotal `json:"data"`
}

type VideoComment struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Content   string `json:"content"`
	LikeCount int64  `json:"like_count"`
	CreatedAt string `json:"created_at"`
}

type VideoCommentList struct {
	Items []VideoComment `json:"items"`
	Total int64          `json:"total"`
}

type ListVideoCommentsResponse struct {
	Base BaseResponse     `json:"base"`
	Data VideoCommentList `json:"data"`
}

// 请求结果封装，包含错误信息
type Result[T any] struct {
	Data       T
	StatusCode int
	RawBody    string
	Err        error
}
