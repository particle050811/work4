package main

import (
	"net/http"
	"net/url"
	"strconv"
)

// testFollowUser 关注用户
func testFollowUser(client *http.Client, baseURL, token, toUserID string) Result[RelationActionResponse] {
	body := map[string]any{
		"to_user_id":  toUserID,
		"action_type": 1,
	}
	var result RelationActionResponse
	status, raw, err := doJSON(client, http.MethodPost, baseURL+"/api/v1/relation/action", body, token, &result)
	return Result[RelationActionResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

// testUnfollowUser 取消关注
func testUnfollowUser(client *http.Client, baseURL, token, toUserID string) Result[RelationActionResponse] {
	body := map[string]any{
		"to_user_id":  toUserID,
		"action_type": 2,
	}
	var result RelationActionResponse
	status, raw, err := doJSON(client, http.MethodPost, baseURL+"/api/v1/relation/action", body, token, &result)
	return Result[RelationActionResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

// testGetFollowingList 获取关注列表
func testGetFollowingList(client *http.Client, baseURL, userID string, pageNum, pageSize int) Result[ListFollowingsResponse] {
	u, err := url.Parse(baseURL + "/api/v1/relation/following/list")
	if err != nil {
		return Result[ListFollowingsResponse]{Err: err}
	}
	q := u.Query()
	q.Set("user_id", userID)
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	var result ListFollowingsResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", "", nil, &result)
	return Result[ListFollowingsResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

// testGetFollowersList 获取粉丝列表
func testGetFollowersList(client *http.Client, baseURL, userID string, pageNum, pageSize int) Result[ListFollowersResponse] {
	u, err := url.Parse(baseURL + "/api/v1/relation/follower/list")
	if err != nil {
		return Result[ListFollowersResponse]{Err: err}
	}
	q := u.Query()
	q.Set("user_id", userID)
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	var result ListFollowersResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", "", nil, &result)
	return Result[ListFollowersResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

// testGetFriendsList 获取好友列表（互相关注）
func testGetFriendsList(client *http.Client, baseURL, token string, pageNum, pageSize int) Result[ListFriendsResponse] {
	u, err := url.Parse(baseURL + "/api/v1/relation/friend/list")
	if err != nil {
		return Result[ListFriendsResponse]{Err: err}
	}
	q := u.Query()
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	var result ListFriendsResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", token, nil, &result)
	return Result[ListFriendsResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func containsProfileByUsername(items []SocialProfile, username string) bool {
	for _, item := range items {
		if item.Username == username {
			return true
		}
	}
	return false
}
