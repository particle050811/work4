package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func testPublishVideo(client *http.Client, baseURL, accessToken, title, description, filePath string) Result[PublishVideoResponse] {
	if strings.TrimSpace(accessToken) == "" {
		return Result[PublishVideoResponse]{Err: fmt.Errorf("缺少 access_token（请先登录）")}
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("title", title)
	_ = w.WriteField("description", description)

	f, err := os.Open(filePath)
	if err != nil {
		return Result[PublishVideoResponse]{Err: err}
	}
	defer f.Close()

	part, err := w.CreateFormFile("video", filepath.Base(filePath))
	if err != nil {
		return Result[PublishVideoResponse]{Err: err}
	}
	if _, err := io.Copy(part, f); err != nil {
		return Result[PublishVideoResponse]{Err: err}
	}
	if err := w.Close(); err != nil {
		return Result[PublishVideoResponse]{Err: err}
	}

	var result PublishVideoResponse
	status, raw, err := doRequest(client, http.MethodPost, baseURL+"/api/v1/video/publish", w.FormDataContentType(), accessToken, &buf, &result)
	return Result[PublishVideoResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func testListPublishedVideos(client *http.Client, baseURL, userID string, pageNum, pageSize int) Result[ListPublishedVideosResponse] {
	u, err := url.Parse(baseURL + "/api/v1/video/list")
	if err != nil {
		return Result[ListPublishedVideosResponse]{Err: err}
	}
	q := u.Query()
	q.Set("user_id", userID)
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	var result ListPublishedVideosResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", "", nil, &result)
	return Result[ListPublishedVideosResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func testSearchVideos(client *http.Client, baseURL, keywords string, pageNum, pageSize int, username string, fromDate, toDate int64, sortBy string) Result[SearchVideosResponse] {
	u, err := url.Parse(baseURL + "/api/v1/video/search")
	if err != nil {
		return Result[SearchVideosResponse]{Err: err}
	}
	q := u.Query()
	if strings.TrimSpace(keywords) != "" {
		q.Set("keywords", keywords)
	}
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	if strings.TrimSpace(username) != "" {
		q.Set("username", username)
	}
	if fromDate > 0 {
		q.Set("from_date", strconv.FormatInt(fromDate, 10))
	}
	if toDate > 0 {
		q.Set("to_date", strconv.FormatInt(toDate, 10))
	}
	if strings.TrimSpace(sortBy) != "" {
		q.Set("sort_by", sortBy)
	}
	u.RawQuery = q.Encode()

	var result SearchVideosResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", "", nil, &result)
	return Result[SearchVideosResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func testGetHotVideos(client *http.Client, baseURL string, pageNum, pageSize int) Result[GetHotVideosResponse] {
	u, err := url.Parse(baseURL + "/api/v1/video/hot")
	if err != nil {
		return Result[GetHotVideosResponse]{Err: err}
	}
	q := u.Query()
	q.Set("page_num", strconv.Itoa(pageNum))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	var result GetHotVideosResponse
	status, raw, err := doRequest(client, http.MethodGet, u.String(), "", "", nil, &result)
	return Result[GetHotVideosResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}
