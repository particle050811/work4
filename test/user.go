package main

import (
	"net/http"
	"net/url"
)

func testRegister(client *http.Client, baseURL, username, password string) Result[RegisterResponse] {
	body := map[string]string{
		"username": username,
		"password": password,
	}
	var result RegisterResponse
	status, raw, err := doJSON(client, http.MethodPost, baseURL+"/api/v1/user/register", body, "", &result)
	return Result[RegisterResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func testLogin(client *http.Client, baseURL, username, password string) Result[LoginResponse] {
	body := map[string]string{
		"username": username,
		"password": password,
	}
	var result LoginResponse
	status, raw, err := doJSON(client, http.MethodPost, baseURL+"/api/v1/user/login", body, "", &result)
	return Result[LoginResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func testGetUserInfo(client *http.Client, baseURL, userID string) Result[GetUserInfoResponse] {
	var result GetUserInfoResponse
	u := baseURL + "/api/v1/user/info?user_id=" + url.QueryEscape(userID)
	status, raw, err := doRequest(client, http.MethodGet, u, "", "", nil, &result)
	return Result[GetUserInfoResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}

func testRefreshToken(client *http.Client, baseURL, refreshToken string) Result[RefreshTokenResponse] {
	body := map[string]string{
		"refresh_token": refreshToken,
	}
	var result RefreshTokenResponse
	status, raw, err := doJSON(client, http.MethodPost, baseURL+"/api/v1/user/refresh", body, "", &result)
	return Result[RefreshTokenResponse]{Data: result, StatusCode: status, RawBody: raw, Err: err}
}
