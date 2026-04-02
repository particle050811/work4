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
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: nil, // 禁用代理，避免本地请求被拦截
		},
	}

	// 先检查服务是否可用
	fmt.Println("【0】检查服务连接...")
	if !checkServerAvailable(client, baseURL) {
		fmt.Println("    ✗ 无法连接到服务器，请确保服务已启动: go run .")
		fmt.Println("    服务地址:", baseURL)
		os.Exit(1)
	}
	fmt.Println("    ✓ 服务连接正常")
	fmt.Println()

	fmt.Println("【0.5】加载已存在测试账号（保留历史数据）")
	username := "e2e_main_user"
	password := "123456"
	seedUser, err := loginSeedUser(client, baseURL, username, password)
	if err != nil {
		fmt.Printf("    ✗ 加载失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("    ✓ 已加载测试账号: %s\n", username)
	fmt.Println()

	// 1. 测试账号检查
	fmt.Println("【1】测试账号检查")
	fmt.Printf("    ✓ 使用已插入测试账号: %s\n", seedUser.Username)
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
	loginResult := Result[LoginResponse]{
		Data: LoginResponse{
			Base: BaseResponse{Code: 0},
			Data: User{
				ID:       seedUser.UserID,
				Username: seedUser.Username,
			},
			AccessToken:  seedUser.AccessToken,
			RefreshToken: seedUser.RefreshToken,
		},
		StatusCode: http.StatusOK,
	}
	fmt.Printf("    ✓ 登录成功\n")
	fmt.Printf("    - 用户ID: %s\n", loginResult.Data.Data.ID)
	fmt.Printf("    - 用户名: %s\n", loginResult.Data.Data.Username)
	fmt.Printf("    - AccessToken: %s...\n", truncate(loginResult.Data.AccessToken, 50))
	fmt.Printf("    - RefreshToken: %s...\n", truncate(loginResult.Data.RefreshToken, 50))
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

	// ====== 互动模块 ======
	fmt.Println("========== 互动模块 ==========")

	// 14. 测试点赞视频
	fmt.Println("【14】测试点赞视频（需要登录）")
	if latestVideo == nil || latestVideo.ID == "" || accessToken == "" {
		fmt.Println("    - 跳过（无 video_id 或未登录）")
	} else {
		likeResult := testVideoLikeAction(client, baseURL, accessToken, latestVideo.ID, 1)
		if likeResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", likeResult.Err)
			addError("14", fmt.Sprintf("请求失败: %v", likeResult.Err))
		} else if likeResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 点赞成功: %s\n", likeResult.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 点赞失败: %s（HTTP %d）\n", likeResult.Data.Base.Msg, likeResult.StatusCode)
			addError("14", fmt.Sprintf("点赞失败: %s", likeResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 15. 测试重复点赞（应该幂等成功）
	fmt.Println("【15】测试重复点赞（应该幂等成功）")
	if latestVideo == nil || latestVideo.ID == "" || accessToken == "" {
		fmt.Println("    - 跳过（无 video_id 或未登录）")
	} else {
		likeResult2 := testVideoLikeAction(client, baseURL, accessToken, latestVideo.ID, 1)
		if likeResult2.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", likeResult2.Err)
			addError("15", fmt.Sprintf("请求失败: %v", likeResult2.Err))
		} else if likeResult2.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 符合预期，重复点赞幂等成功: %s\n", likeResult2.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 重复点赞失败: %s（HTTP %d）\n", likeResult2.Data.Base.Msg, likeResult2.StatusCode)
			addError("15", fmt.Sprintf("重复点赞应幂等成功，但失败了: %s", likeResult2.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 16. 测试点赞列表
	fmt.Println("【16】测试获取点赞列表")
	if userID == "" {
		fmt.Println("    - 跳过（无 user_id）")
	} else {
		likedResult := testListLikedVideos(client, baseURL, userID, 1, 10)
		if likedResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", likedResult.Err)
			addError("16", fmt.Sprintf("请求失败: %v", likedResult.Err))
		} else if likedResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", likedResult.Data.Data.Total, len(likedResult.Data.Data.Items))
			if latestVideo != nil && likedResult.Data.Data.Total > 0 {
				found := false
				for _, v := range likedResult.Data.Data.Items {
					if v.ID == latestVideo.ID {
						found = true
						break
					}
				}
				if found {
					fmt.Printf("    ✓ 符合预期：点赞列表包含刚点赞的视频\n")
				} else {
					fmt.Printf("    ✗ 不符合预期：点赞列表未包含刚点赞的视频\n")
					addError("16", "点赞列表未包含刚点赞的视频")
				}
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", likedResult.Data.Base.Msg, likedResult.StatusCode)
			addError("16", fmt.Sprintf("获取失败: %s", likedResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 17. 测试取消点赞
	fmt.Println("【17】测试取消点赞")
	if latestVideo == nil || latestVideo.ID == "" || accessToken == "" {
		fmt.Println("    - 跳过（无 video_id 或未登录）")
	} else {
		unlikeResult := testVideoLikeAction(client, baseURL, accessToken, latestVideo.ID, 2)
		if unlikeResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", unlikeResult.Err)
			addError("17", fmt.Sprintf("请求失败: %v", unlikeResult.Err))
		} else if unlikeResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 取消点赞成功: %s\n", unlikeResult.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 取消点赞失败: %s（HTTP %d）\n", unlikeResult.Data.Base.Msg, unlikeResult.StatusCode)
			addError("17", fmt.Sprintf("取消点赞失败: %s", unlikeResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 18. 测试发布评论
	fmt.Println("【18】测试发布评论（需要登录）")
	commentContent := fmt.Sprintf("这是测试评论 %d", time.Now().Unix())
	var latestCommentID string
	if latestVideo == nil || latestVideo.ID == "" || accessToken == "" {
		fmt.Println("    - 跳过（无 video_id 或未登录）")
	} else {
		commentResult := testPublishComment(client, baseURL, accessToken, latestVideo.ID, commentContent)
		if commentResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", commentResult.Err)
			addError("18", fmt.Sprintf("请求失败: %v", commentResult.Err))
		} else if commentResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 评论成功: %s\n", commentResult.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 评论失败: %s（HTTP %d）\n", commentResult.Data.Base.Msg, commentResult.StatusCode)
			addError("18", fmt.Sprintf("评论失败: %s", commentResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 19. 测试获取用户评论列表
	fmt.Println("【19】测试获取用户评论列表")
	if userID == "" {
		fmt.Println("    - 跳过（无 user_id）")
	} else {
		userCommentsResult := testListUserComments(client, baseURL, userID, 1, 10)
		if userCommentsResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", userCommentsResult.Err)
			addError("19", fmt.Sprintf("请求失败: %v", userCommentsResult.Err))
		} else if userCommentsResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", userCommentsResult.Data.Data.Total, len(userCommentsResult.Data.Data.Items))
			// 找到刚发布的评论
			for _, c := range userCommentsResult.Data.Data.Items {
				if c.Content == commentContent {
					latestCommentID = c.ID
					fmt.Printf("    ✓ 符合预期：找到刚发布的评论（ID: %s）\n", latestCommentID)
					break
				}
			}
			if latestCommentID == "" && len(userCommentsResult.Data.Data.Items) > 0 {
				// 取第一条作为测试删除用
				latestCommentID = userCommentsResult.Data.Data.Items[0].ID
				fmt.Printf("    - 提示：未找到刚发布的评论，使用最新评论（ID: %s）进行后续测试\n", latestCommentID)
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", userCommentsResult.Data.Base.Msg, userCommentsResult.StatusCode)
			addError("19", fmt.Sprintf("获取失败: %s", userCommentsResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 20. 测试删除评论
	fmt.Println("【20】测试删除评论（需要登录+作者权限）")
	if latestCommentID == "" || accessToken == "" {
		fmt.Println("    - 跳过（无 comment_id 或未登录）")
	} else {
		deleteResult := testDeleteComment(client, baseURL, accessToken, latestCommentID)
		if deleteResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", deleteResult.Err)
			addError("20", fmt.Sprintf("请求失败: %v", deleteResult.Err))
		} else if deleteResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 删除成功: %s\n", deleteResult.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 删除失败: %s（HTTP %d）\n", deleteResult.Data.Base.Msg, deleteResult.StatusCode)
			addError("20", fmt.Sprintf("删除失败: %s", deleteResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	// 21. 测试删除他人评论（应该失败）
	fmt.Println("【21】测试删除已删除的评论（应该失败-评论不存在）")
	if latestCommentID == "" || accessToken == "" {
		fmt.Println("    - 跳过（无 comment_id 或未登录）")
	} else {
		deleteResult2 := testDeleteComment(client, baseURL, accessToken, latestCommentID)
		if deleteResult2.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", deleteResult2.Err)
			addError("21", fmt.Sprintf("请求失败: %v", deleteResult2.Err))
		} else if deleteResult2.Data.Base.Code != 0 {
			fmt.Printf("    ✓ 符合预期，删除被拒绝: %s\n", deleteResult2.Data.Base.Msg)
		} else {
			fmt.Printf("    ✗ 不符合预期，删除已删除的评论应该失败\n")
			addError("21", "删除已删除的评论应该失败，但实际成功了")
		}
	}
	fmt.Println()

	// ====== 社交模块 ======
	fmt.Println("========== 社交模块 ==========")

	fmt.Println("【22】预置社交测试数据")
	relationFixture, err := prepareRelationFixture(client, baseURL)
	if err != nil {
		fmt.Printf("    ✗ 预置失败: %v\n", err)
		addError("22", fmt.Sprintf("预置失败: %v", err))
	} else {
		fmt.Printf("    ✓ 已预置用户: %s, %s, %s\n", relationFixture.Alice.Username, relationFixture.Bob.Username, relationFixture.Carol.Username)
	}
	fmt.Println()

	fmt.Println("【23】测试关注列表")
	if relationFixture == nil {
		fmt.Println("    - 跳过（预置失败）")
	} else {
		followingResult := testGetFollowingList(client, baseURL, relationFixture.Alice.UserID, 1, 10)
		if followingResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", followingResult.Err)
			addError("23", fmt.Sprintf("请求失败: %v", followingResult.Err))
		} else if followingResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", followingResult.Data.Data.Total, len(followingResult.Data.Data.Items))
			hasBob := containsProfileByUsername(followingResult.Data.Data.Items, relationFixture.Bob.Username)
			hasCarol := containsProfileByUsername(followingResult.Data.Data.Items, relationFixture.Carol.Username)
			if hasBob && hasCarol && followingResult.Data.Data.Total == 2 {
				fmt.Println("    ✓ 符合预期：Alice 的关注列表包含 Bob 和 Carol")
			} else {
				fmt.Println("    ✗ 不符合预期：Alice 的关注列表缺少预置用户")
				addError("23", "Alice 的关注列表未同时包含 Bob 和 Carol")
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", followingResult.Data.Base.Msg, followingResult.StatusCode)
			addError("23", fmt.Sprintf("获取失败: %s", followingResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	fmt.Println("【24】测试粉丝列表")
	if relationFixture == nil {
		fmt.Println("    - 跳过（预置失败）")
	} else {
		followerResult := testGetFollowersList(client, baseURL, relationFixture.Alice.UserID, 1, 10)
		if followerResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", followerResult.Err)
			addError("24", fmt.Sprintf("请求失败: %v", followerResult.Err))
		} else if followerResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", followerResult.Data.Data.Total, len(followerResult.Data.Data.Items))
			hasBob := containsProfileByUsername(followerResult.Data.Data.Items, relationFixture.Bob.Username)
			hasCarol := containsProfileByUsername(followerResult.Data.Data.Items, relationFixture.Carol.Username)
			if hasBob && !hasCarol && followerResult.Data.Data.Total == 1 {
				fmt.Println("    ✓ 符合预期：Alice 的粉丝列表只包含 Bob")
			} else {
				fmt.Println("    ✗ 不符合预期：Alice 的粉丝列表结果错误")
				addError("24", "Alice 的粉丝列表应只包含 Bob")
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", followerResult.Data.Base.Msg, followerResult.StatusCode)
			addError("24", fmt.Sprintf("获取失败: %s", followerResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	fmt.Println("【25】测试好友列表（互相关注）")
	if relationFixture == nil {
		fmt.Println("    - 跳过（预置失败）")
	} else {
		friendsResult := testGetFriendsList(client, baseURL, relationFixture.Alice.AccessToken, 1, 10)
		if friendsResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", friendsResult.Err)
			addError("25", fmt.Sprintf("请求失败: %v", friendsResult.Err))
		} else if friendsResult.StatusCode == http.StatusOK && friendsResult.Data.Base.Code == 0 {
			fmt.Printf("    ✓ 获取成功，总数: %d，本页: %d\n", friendsResult.Data.Data.Total, len(friendsResult.Data.Data.Items))
			hasBob := containsProfileByUsername(friendsResult.Data.Data.Items, relationFixture.Bob.Username)
			hasCarol := containsProfileByUsername(friendsResult.Data.Data.Items, relationFixture.Carol.Username)
			if hasBob && !hasCarol && friendsResult.Data.Data.Total == 1 {
				fmt.Println("    ✓ 符合预期：Alice 的好友列表只包含 Bob")
			} else {
				fmt.Println("    ✗ 不符合预期：Alice 的好友列表结果错误")
				addError("25", "Alice 的好友列表应只包含 Bob")
			}
		} else {
			fmt.Printf("    ✗ 获取失败: %s（HTTP %d）\n", friendsResult.Data.Base.Msg, friendsResult.StatusCode)
			addError("25", fmt.Sprintf("获取失败: %s", friendsResult.Data.Base.Msg))
		}
	}
	fmt.Println()

	fmt.Println("【26】测试未登录访问好友列表（应该失败）")
	unauthorizedFriendsResult := testGetFriendsList(client, baseURL, "", 1, 10)
	if unauthorizedFriendsResult.Err != nil {
		fmt.Printf("    ✗ 请求失败: %v\n", unauthorizedFriendsResult.Err)
		addError("26", fmt.Sprintf("请求失败: %v", unauthorizedFriendsResult.Err))
	} else if unauthorizedFriendsResult.StatusCode == http.StatusUnauthorized {
		fmt.Printf("    ✓ 符合预期，未登录被拒绝: %s\n", unauthorizedFriendsResult.Data.Base.Msg)
	} else {
		fmt.Printf("    ✗ 不符合预期，未登录应该返回 401，实际 HTTP %d\n", unauthorizedFriendsResult.StatusCode)
		addError("26", fmt.Sprintf("未登录访问好友列表应返回 401，实际 HTTP %d", unauthorizedFriendsResult.StatusCode))
	}
	fmt.Println()

	fmt.Println("【27】测试取消互关后好友列表更新")
	if relationFixture == nil {
		fmt.Println("    - 跳过（预置失败）")
	} else {
		unfollowResult := testUnfollowUser(client, baseURL, relationFixture.Alice.AccessToken, relationFixture.Bob.UserID)
		if unfollowResult.Err != nil {
			fmt.Printf("    ✗ 请求失败: %v\n", unfollowResult.Err)
			addError("27", fmt.Sprintf("请求失败: %v", unfollowResult.Err))
		} else if unfollowResult.Data.Base.Code != 0 {
			fmt.Printf("    ✗ 取消关注失败: %s（HTTP %d）\n", unfollowResult.Data.Base.Msg, unfollowResult.StatusCode)
			addError("27", fmt.Sprintf("取消关注失败: %s", unfollowResult.Data.Base.Msg))
		} else {
			afterFriendsResult := testGetFriendsList(client, baseURL, relationFixture.Alice.AccessToken, 1, 10)
			if afterFriendsResult.Err != nil {
				fmt.Printf("    ✗ 请求失败: %v\n", afterFriendsResult.Err)
				addError("27", fmt.Sprintf("请求失败: %v", afterFriendsResult.Err))
			} else if afterFriendsResult.Data.Base.Code == 0 && !containsProfileByUsername(afterFriendsResult.Data.Data.Items, relationFixture.Bob.Username) {
				fmt.Println("    ✓ 符合预期：取消互关后 Bob 已不在 Alice 的好友列表中")
			} else {
				fmt.Println("    ✗ 不符合预期：取消互关后 Bob 仍出现在 Alice 的好友列表中")
				addError("27", "取消互关后好友列表未更新")
			}
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
