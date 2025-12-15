# 互动模块实现详解

> 日期：2025-12-15
> 主题：点赞、评论的实现模式与权限校验

## 1. 模块概览

互动模块包含 5 个接口：

| 接口 | 方法 | 路径 | 认证 |
|------|------|------|------|
| 点赞/取消点赞 | POST | `/api/v1/interaction/like` | 需要 |
| 点赞列表 | GET | `/api/v1/interaction/like/list` | 不需要 |
| 发布评论 | POST | `/api/v1/interaction/comment` | 需要 |
| 评论列表 | GET | `/api/v1/interaction/comment/list` | 不需要 |
| 删除评论 | POST | `/api/v1/interaction/comment/delete` | 需要 |

## 2. 点赞模块设计

### 2.1 数据模型

```go
// biz/dal/model/like.go
type VideoLike struct {
    ID        uint           `gorm:"primaryKey"`
    UserID    uint           `gorm:"uniqueIndex:idx_user_video;not null"`
    VideoID   uint           `gorm:"uniqueIndex:idx_user_video;index;not null"`
    CreatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`  // 软删除
}
```

**设计要点**：
- `uniqueIndex:idx_user_video`：联合唯一索引，防止重复点赞
- 软删除：取消点赞不物理删除，便于恢复和数据分析

### 2.2 点赞操作的事务处理

```go
// biz/handler/v1/interaction_service.go:87-133
var likeDelta int64
err = store.WithTx(func(txStore *dal.Store) error {
    // 1. 查询点赞记录（包含软删除的）
    like, err := db.GetVideoLikeUnscoped(txStore, userID, videoID)
    if err != nil {
        return err
    }

    if actionType == 1 { // 点赞
        if like == nil {
            // 不存在 → 创建新记录
            newLike := &model.VideoLike{UserID: userID, VideoID: videoID}
            db.CreateVideoLike(txStore, newLike)
            likeDelta = 1
        } else if like.DeletedAt.Valid {
            // 已软删除 → 恢复记录
            db.RestoreVideoLike(txStore, like.ID)
            likeDelta = 1
        }
        // 已点赞 → 幂等返回，likeDelta = 0
    } else { // 取消点赞
        if like != nil && !like.DeletedAt.Valid {
            // 存在且未删除 → 软删除
            db.SoftDeleteVideoLike(txStore, like.ID)
            likeDelta = -1
        }
        // 不存在或已删除 → 幂等返回，likeDelta = 0
    }

    // 2. 更新视频点赞数（增量更新）
    if likeDelta != 0 {
        db.IncreaseVideoLikeCount(txStore, videoID, likeDelta)
    }
    return nil
})
```

### 2.3 增量更新视频点赞数

```go
// biz/dal/db/like_dao.go
func IncreaseVideoLikeCount(store DBProvider, videoID uint, delta int64) error {
    return store.DB().Model(&model.Video{}).
        Where("id = ?", videoID).
        Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}
```

**为什么用 `gorm.Expr` 而非先查后改？**

```go
// 错误示范（并发不安全）
video := GetVideo(id)
video.LikeCount += 1
db.Save(video)  // 并发时会丢失更新

// 正确做法（原子操作）
db.Update("like_count", gorm.Expr("like_count + ?", 1))
// 生成 SQL: UPDATE videos SET like_count = like_count + 1 WHERE id = ?
```

### 2.4 幂等性设计

| 场景 | 处理方式 | likeDelta |
|------|----------|-----------|
| 首次点赞 | 创建记录 | +1 |
| 重复点赞 | 不操作 | 0 |
| 取消后再点赞 | 恢复软删除记录 | +1 |
| 首次取消 | 软删除记录 | -1 |
| 重复取消 | 不操作 | 0 |
| 未点赞时取消 | 不操作 | 0 |

## 3. 评论模块设计

### 3.1 发布评论的事务

```go
// biz/handler/v1/interaction_service.go:303-314
err = store.WithTx(func(txStore *dal.Store) error {
    comment := &model.Comment{
        UserID:  userID,
        VideoID: videoID,
        Content: content,
    }
    if err := db.CreateComment(txStore, comment); err != nil {
        return err
    }
    // 同时更新视频评论数
    return db.IncreaseVideoCommentCount(txStore, videoID, 1)
})
```

### 3.2 删除评论的权限校验

```go
// biz/handler/v1/interaction_service.go:441-447
// 权限校验：只能删除自己的评论
if comment.UserID != userID {
    c.JSON(consts.StatusForbidden, &v1.DeleteCommentResponse{
        Base: response.Forbidden("无权删除他人评论"),
    })
    return
}
```

## 4. 权限校验对比

| 操作 | 是否需要校验 | 原因 |
|------|-------------|------|
| 点赞/取消点赞 | 否 | 查询条件天然包含 `userID`，只能操作自己的记录 |
| 删除评论 | 是 | 请求参数是 `comment_id`，需验证评论归属 |
| 发布评论 | 否 | 创建的记录 `UserID` 就是当前用户 |

**点赞为何不需要权限校验？**

```go
// 查询条件已包含当前用户 ID
like, err := db.GetVideoLikeUnscoped(txStore, userID, videoID)
// 只会查到当前用户对该视频的点赞，无法操作他人数据
```

## 5. Redis 热榜同步

点赞和评论都会更新 Redis 热榜缓存：

```go
const hotVideosKey = "fanone:video:hot:zset"

// 点赞权重为 3
if likeDelta != 0 {
    scoreDelta := float64(likeDelta * 3)
    store.Redis().ZIncrBy(ctx, hotVideosKey, scoreDelta, fmt.Sprintf("%d", videoID))
}

// 评论权重为 2
store.Redis().ZIncrBy(ctx, hotVideosKey, 2.0, fmt.Sprintf("%d", videoID))
```

**热榜分数公式**：`score = like_count*3 + comment_count*2 + visit_count*1`

## 6. 数据流总结

```
点赞流程：
┌──────────────────────────────────────────────────────────┐
│  POST /api/v1/interaction/like                           │
│  ├─ JWT 中间件提取 userID                                 │
│  ├─ 参数校验 (video_id, action_type)                     │
│  ├─ 检查视频是否存在                                      │
│  └─ 事务开始 ─────────────────────────────────────────── │
│       ├─ 查询 VideoLike 记录 (Unscoped)                  │
│       ├─ 根据 action_type 创建/恢复/软删除                │
│       ├─ 计算 likeDelta (+1/-1/0)                        │
│       └─ UPDATE videos SET like_count += likeDelta       │
│     事务提交 ────────────────────────────────────────────│
│  └─ ZINCRBY fanone:video:hot:zset (非事务，失败不阻塞)    │
└──────────────────────────────────────────────────────────┘

删除评论流程：
┌──────────────────────────────────────────────────────────┐
│  POST /api/v1/interaction/comment/delete                 │
│  ├─ JWT 中间件提取 userID                                 │
│  ├─ 参数校验 (comment_id)                                │
│  ├─ 查询评论                                              │
│  ├─ 权限校验：comment.UserID == userID ?                 │
│  │   └─ 不匹配 → 403 Forbidden                           │
│  └─ 事务开始 ─────────────────────────────────────────── │
│       ├─ 软删除评论                                       │
│       └─ UPDATE videos SET comment_count -= 1            │
│     事务提交 ────────────────────────────────────────────│
│  └─ ZINCRBY fanone:video:hot:zset -2                     │
└──────────────────────────────────────────────────────────┘
```

## 7. 关键代码位置

| 功能 | 文件 | 行号 |
|------|------|------|
| 点赞操作 Handler | `biz/handler/v1/interaction_service.go` | 26-158 |
| 删除评论权限校验 | `biz/handler/v1/interaction_service.go` | 441-447 |
| VideoLike 模型 | `biz/dal/model/like.go` | 9-17 |
| 点赞 DAO | `biz/dal/db/like_dao.go` | 全文件 |
| 评论 DAO | `biz/dal/db/comment_dao.go` | 51-90 |
| 认证中间件配置 | `biz/router/v1/middleware.go` | 95-118 |

## 8. 推荐阅读

- [GORM 软删除文档](https://gorm.io/zh_CN/docs/delete.html#%E8%BD%AF%E5%88%A0%E9%99%A4)
- [GORM 事务处理](https://gorm.io/zh_CN/docs/transactions.html)
- [Redis ZINCRBY 命令](https://redis.io/commands/zincrby/)
