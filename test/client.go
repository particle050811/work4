package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const defaultBaseURL = "http://localhost:8888"

func getBaseURL() string {
	v := strings.TrimSpace(os.Getenv("BASE_URL"))
	if v == "" {
		return defaultBaseURL
	}
	return strings.TrimRight(v, "/")
}

func checkServerAvailable(client *http.Client, baseURL string) bool {
	resp, err := client.Get(baseURL + "/ping")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return true
}

func doJSON(client *http.Client, method, url string, body any, token string, out any) (int, string, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, "", err
		}
		reader = bytes.NewBuffer(b)
	}
	return doRequest(client, method, url, "application/json", token, reader, out)
}

func doRequest(client *http.Client, method, url, contentType, token string, body io.Reader, out any) (int, string, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, "", err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	raw := string(rawBytes)

	if out != nil && len(rawBytes) > 0 {
		if err := json.Unmarshal(rawBytes, out); err != nil {
			return resp.StatusCode, raw, fmt.Errorf("解析 JSON 失败: %w（响应体: %s）", err, truncate(raw, 200))
		}
	}
	return resp.StatusCode, raw, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func checkStaticAvailable(client *http.Client, baseURL, path string) (bool, int, error) {
	u := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return false, 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 300, resp.StatusCode, nil
}

func prepareVideoFile() (string, func(), error) {
	if p := strings.TrimSpace(os.Getenv("VIDEO_FILE")); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", nil, fmt.Errorf("VIDEO_FILE 不存在: %w", err)
		}
		return p, func() {}, nil
	}

	f, err := os.CreateTemp("", "fanone_test_*.mp4")
	if err != nil {
		return "", nil, err
	}
	path := f.Name()

	// 写入一些随机字节，保证上传的文件非空
	payload := bytes.Repeat([]byte("fanone-test-video\n"), 1024)
	if _, err := f.Write(payload); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", nil, err
	}

	return path, func() { _ = os.Remove(path) }, nil
}
