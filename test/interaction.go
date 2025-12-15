package main

import (
	"net/http"
	"net/url"
	"strconv"
)

func testListVideoComments(client *http.Client, baseURL, videoID string, pageNum, pageSize int) Result[ListVideoCommentsResponse] {
	u, err := url.Parse(baseURL + "/api/v1/video/comments")
	if err != nil {
		return Result[ListVideoCommentsResponse]{Err: err}
	}
	q := u.Query()
	q.Set("video_id", videoID)
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	var result ListVideoCommentsResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", "", nil, &result)
	return Result[ListVideoCommentsResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

// TODO: 待实现
// - testLikeVideo: 点赞视频
// - testUnlikeVideo: 取消点赞
// - testGetVideoLikes: 获取视频点赞列表
// - testCommentVideo: 发表评论
// - testDeleteComment: 删除评论
