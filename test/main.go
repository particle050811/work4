package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// 错误收集器
var errors []string

func addError(testName, msg string) {
	errors = append(errors, fmt.Sprintf("【%s】%s", testName, msg))
}

func main() {
	fmt.Println("========== FanOne 视频平台 API 测试 ==========")
	fmt.Println()

	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}

	// 先检查服务是否可用
	fmt.Println("【0】检查服务连接...")
	if !checkServerAvailable(client, baseURL) {
		fmt.Println("    ✗ 无法连接到服务器，请确保服务已启动: go run .")
		fmt.Println("    服务地址:", baseURL)
		os.Exit(1)
	}
	fmt.Println("    ✓ 服务连接正常")
	fmt.Println()

	// 生成唯一用户名避免冲突
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	password := "123456"

	// 1. 测试注册
	fmt.Println("【1】测试用户注册")
	registerResult := testRegister(client, baseURL, username, password)
	if registerResult.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", registerResult.Err)
		addError("1", fmt.Sprintf("请求失败: %v", registerResult.Err))
	} else if registerResult.Data.Base.Code == 0 {
		fmt.Printf("    ✓ 注册成功: %s\n", registerResult.Data.Base.Msg)
	} else {
		fmt.Printf("    ✗ 注册失败: %s（HTTP %d）\n", registerResult.Data.Base.Msg, registerResult.StatusCode)
		addError("1", fmt.Sprintf("注册失败: %s", registerResult.Data.Base.Msg))
	}
	fmt.Println()

	// 2. 测试重复注册
	fmt.Println("【2】测试重复注册（应该失败）")
	registerResult2 := testRegister(client, baseURL, username, password)
	if registerResult2.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", registerResult2.Err)
		addError("2", fmt.Sprintf("请求失败: %v", registerResult2.Err))
	} else if registerResult2.Data.Base.Code != 0 {
		fmt.Printf("    ✓ 符合预期，重复注册被拒绝: %s\n", registerResult2.Data.Base.Msg)
	} else {
		fmt.Printf("    ✗ 不符合预期，重复注册应该失败\n")
		addError("2", "重复注册应该失败，但实际成功了")
	}
	fmt.Println()

	// 3. 测试登录
	fmt.Println("【3】测试用户登录")
	loginResult := testLogin(client, baseURL, username, password)
	if loginResult.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", loginResult.Err)
		addError("3", fmt.Sprintf("请求失败: %v", loginResult.Err))
	} else if loginResult.Data.Base.Code == 0 {
		fmt.Printf("    ✓ 登录成功\n")
		fmt.Printf("    - 用户ID: %s\n", loginResult.Data.Data.ID)
		fmt.Printf("    - 用户名: %s\n", loginResult.Data.Data.Username)
		fmt.Printf("    - AccessToken: %s...\n", truncate(loginResult.Data.AccessToken, 50))
		fmt.Printf("    - RefreshToken: %s...\n", truncate(loginResult.Data.RefreshToken, 50))
	} else {
		fmt.Printf("    ✗ 登录失败: %s（HTTP %d）\n", loginResult.Data.Base.Msg, loginResult.StatusCode)
		addError("3", fmt.Sprintf("登录失败: %s", loginResult.Data.Base.Msg))
	}
	fmt.Println()

	// 4. 测试错误密码登录
	fmt.Println("【4】测试错误密码登录（应该失败）")
	loginResult2 := testLogin(client, baseURL, username, "wrongpassword")
	if loginResult2.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", loginResult2.Err)
		addError("4", fmt.Sprintf("请求失败: %v", loginResult2.Err))
	} else if loginResult2.Data.Base.Code != 0 {
		fmt.Printf("    ✓ 符合预期，错误密码被拒绝: %s\n", loginResult2.Data.Base.Msg)
	} else {
		fmt.Printf("    ✗ 不符合预期，错误密码应该登录失败\n")
		addError("4", "错误密码应该登录失败，但实际成功了")
	}
	fmt.Println()

	// 5. 测试获取用户信息
	fmt.Println("【5】测试获取用户信息")
	if loginResult.Err == nil && loginResult.Data.Data.ID != "" {
		userInfoResult := testGetUserInfo(client, baseURL, loginResult.Data.Data.ID)
		if userInfoResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", userInfoResult.Err)
			addError("5", fmt.Sprintf("请求失败: %v", userInfoResult.Err))
		} else if userInfoResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功\n")
			fmt.Printf("    - 用户ID: %s\n", userInfoResult.Data.Data.ID)
			fmt.Printf("    - 用户名: %s\n", userInfoResult.Data.Data.Username)
			fmt.Printf("    - 创建时间: %s\n", userInfoResult.Data.Data.CreatedAt)
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", userInfoResult.Data.Base.Msg, userInfoResult.StatusCode)
			addError("5", fmt.Sprintf("获取失败: %s", userInfoResult.Data.Base.Msg))
		}
	} else {
		fmt.Println("    - 跳过（登录失败，无用户ID）")
	}
	fmt.Println()

	// 6. 测试刷新令牌
	fmt.Println("【6】测试刷新令牌")
	if loginResult.Err == nil && loginResult.Data.RefreshToken != "" {
		refreshResult := testRefreshToken(client, baseURL, loginResult.Data.RefreshToken)
		if refreshResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", refreshResult.Err)
			addError("6", fmt.Sprintf("请求失败: %v", refreshResult.Err))
		} else if refreshResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 刷新成功\n")
			fmt.Printf("    - 新 AccessToken: %s...\n", truncate(refreshResult.Data.AccessToken, 50))
			fmt.Printf("    - 新 RefreshToken: %s...\n", truncate(refreshResult.Data.RefreshToken, 50))
		} else {
			fmt.Printf("    ✗ 刷新失败: %s（HTTP %d）\n", refreshResult.Data.Base.Msg, refreshResult.StatusCode)
			addError("6", fmt.Sprintf("刷新失败: %s", refreshResult.Data.Base.Msg))
		}
	} else {
		fmt.Println("    - 跳过（登录失败，无 RefreshToken）")
	}
	fmt.Println()

	// 7. 测试无效令牌刷新
	fmt.Println("【7】测试无效令牌刷新（应该失败）")
	refreshResult2 := testRefreshToken(client, baseURL, "invalid_token")
	if refreshResult2.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", refreshResult2.Err)
		addError("7", fmt.Sprintf("请求失败: %v", refreshResult2.Err))
	} else if refreshResult2.Data.Base.Code != 0 {
		fmt.Printf("    ✓ 符合预期，无效令牌被拒绝: %s\n", refreshResult2.Data.Base.Msg)
	} else {
		fmt.Printf("    ✗ 不符合预期，无效令牌应该刷新失败\n")
		addError("7", "无效令牌应该刷新失败，但实际成功了")
	}
	fmt.Println()

	// ====== 视频模块 ======
	fmt.Println("========== 视频模块 ==========")

	accessToken := ""
	userID := ""
	if loginResult.Err == nil && loginResult.Data.Base.Code == 0 {
		accessToken = loginResult.Data.AccessToken
		userID = loginResult.Data.Data.ID
	}

	// 8. 测试投稿
	fmt.Println("【8】测试视频投稿（需要登录）")
	videoTitle := fmt.Sprintf("test video %d", time.Now().Unix())
	videoDesc := "这是一个用于自动化测试的投稿视频（内容为随机字节，不是真实视频）"
	videoFilePath, cleanup, err := prepareVideoFile()
	if err != nil {
		fmt.Printf("    ✗ 准备视频文件失败: %v\n", err)
		addError("8", fmt.Sprintf("准备视频文件失败: %v", err))
	} else {
		defer cleanup()
		publishResult := testPublishVideo(client, baseURL, accessToken, videoTitle, videoDesc, videoFilePath)
		if publishResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", publishResult.Err)
			addError("8", fmt.Sprintf("请求失败: %v", publishResult.Err))
		} else if publishResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 投稿成功: %s\n", publishResult.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 投稿失败: %s（HTTP %d）\n", publishResult.Data.Base.Msg, publishResult.StatusCode)
			addError("8", fmt.Sprintf("投稿失败: %s", publishResult.Data.Base.Msg))
			if publishResult.RawBody != "" {
				fmt.Printf("    - 响应体: %s\n", truncate(publishResult.RawBody, 200))
			}
		}
	}
	fmt.Println()

	// 9. 测试发布列表
	fmt.Println("【9】测试发布列表")
	var latestVideo *Video
	if userID == "" {
		fmt.Println("    - 跳过（登录失败，无 user_id）")
	} else {
		listResult := testListPublishedVideos(client, baseURL, userID, 1, 10)
		if listResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", listResult.Err)
			addError("9", fmt.Sprintf("请求失败: %v", listResult.Err))
		} else if listResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", listResult.Data.Data.Total, len(listResult.Data.Data.Items))
			if len(listResult.Data.Data.Items) > 0 {
				latestVideo = &listResult.Data.Data.Items[0]
				fmt.Printf("    - 最新视频ID: %s\n", latestVideo.ID)
				fmt.Printf("    - 标题: %s\n", latestVideo.Title)
				fmt.Printf("    - 视频URL: %s\n", latestVideo.VideoURL)
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", listResult.Data.Base.Msg, listResult.StatusCode)
			addError("9", fmt.Sprintf("获取失败: %s", listResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 10. 测试视频文件是否可访问
	fmt.Println("【10】测试视频文件可访问（静态资源）")
	if latestVideo == nil || latestVideo.VideoURL == "" {
		fmt.Println("    - 跳过（无可用 video_url）")
	} else {
		ok, status, err := checkStaticAvailable(client, baseURL, latestVideo.VideoURL)
		if err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", err)
			addError("10", fmt.Sprintf("请求失败: %v", err))
		} else if ok {
			fmt.Printf("    ✓ 可访问（HTTP %d）\n", status)
		} else {
			fmt.Printf("    ✗ 不可访问（HTTP %d）\n", status)
			addError("10", fmt.Sprintf("视频文件不可访问，HTTP %d", status))
		}
	}
	fmt.Println()

	// 11. 测试搜索视频
	fmt.Println("【11】测试搜索视频")
	searchResult := testSearchVideos(client, baseURL, videoTitle, 1, 10, "", 0, 0, "latest")
	if searchResult.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", searchResult.Err)
		addError("11", fmt.Sprintf("请求失败: %v", searchResult.Err))
	} else if searchResult.Data.Base.Code == 0 {
		fmt.Printf("    ✓ 搜索成功，总数: %d，本页: %d\n", searchResult.Data.Data.Total, len(searchResult.Data.Data.Items))
		found := false
		for _, v := range searchResult.Data.Data.Items {
			if v.Title == videoTitle {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("    ✓ 符合预期：能搜到刚投稿的视频\n")
		} else {
			fmt.Printf("    ✗ 不符合预期：未搜到刚投稿的视频\n")
			addError("11", "未搜到刚投稿的视频")
		}
	} else {
		fmt.Printf("    ✗ 搜索失败: %s（HTTP %d）\n", searchResult.Data.Base.Msg, searchResult.StatusCode)
		addError("11", fmt.Sprintf("搜索失败: %s", searchResult.Data.Base.Msg))
	}
	fmt.Println()

	// 12. 测试热门排行榜（要求 Redis 缓存）
	fmt.Println("【12】测试热门排行榜")
	hotResult := testGetHotVideos(client, baseURL, 1, 10)
	if hotResult.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", hotResult.Err)
		addError("12", fmt.Sprintf("请求失败: %v", hotResult.Err))
	} else if hotResult.Data.Base.Code == 0 {
		fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", hotResult.Data.Data.Total, len(hotResult.Data.Data.Items))
	} else {
		fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", hotResult.Data.Base.Msg, hotResult.StatusCode)
		addError("12", fmt.Sprintf("获取失败: %s", hotResult.Data.Base.Msg))
	}
	fmt.Println()

	// 13. 测试视频评论列表（新视频应为空）
	fmt.Println("【13】测试视频评论列表（新视频应为空）")
	if latestVideo == nil || latestVideo.ID == "" {
		fmt.Println("    - 跳过（无 video_id）")
	} else {
		commentsResult := testListVideoComments(client, baseURL, latestVideo.ID, 1, 10)
		if commentsResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", commentsResult.Err)
			addError("13", fmt.Sprintf("请求失败: %v", commentsResult.Err))
		} else if commentsResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", commentsResult.Data.Data.Total, len(commentsResult.Data.Data.Items))
			if commentsResult.Data.Data.Total == 0 {
				fmt.Printf("    ✓ 符合预期：新视频评论为空\n")
			} else {
				fmt.Printf("    - 提示：该视频已有评论（可能是历史数据）\n")
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", commentsResult.Data.Base.Msg, commentsResult.StatusCode)
			addError("13", fmt.Sprintf("获取失败: %s", commentsResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 输出测试汇总
	fmt.Println("========== 测试完成 ==========")
	if len(errors) == 0 {
		fmt.Println("✓ 所有测试通过！")
	} else {
		fmt.Printf("✗ 共 %d 个错误：\n", len(errors))
		fmt.Println()
		for i, e := range errors {
			fmt.Printf("  %d. %s\n", i+1, e)
		}
	}
}
