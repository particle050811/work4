# Go Context 用法详解

> 日期：2025-12-07
> 主题：context.Context 在 Hertz 项目中的使用

## 1. 概述

`context.Context` 是 Go 标准库中用于控制并发操作的核心类型，在 Hertz Handler 中作为第一个参数传递：

```go
func (s *UserService) Login(ctx context.Context, c *app.RequestContext, req *v1.LoginRequest) {
    // ctx: 控制流（超时、取消、传值）
    // c:   HTTP 交互（请求/响应）
}
```

## 2. 核心功能

### 2.1 传递请求级数据

```go
// 存储值
ctx = context.WithValue(ctx, "user_id", uint64(123))

// 读取值
userID := ctx.Value("user_id").(uint64)
```

### 2.2 控制超时

```go
// 设置 5 秒超时
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel() // 必须调用，释放资源

// 传递给数据库/HTTP 请求
result, err := db.QueryContext(ctx, "SELECT ...")
if err == context.DeadlineExceeded {
    // 超时处理
}
```

### 2.3 手动取消

```go
ctx, cancel := context.WithCancel(ctx)

go func() {
    if shouldStop {
        cancel()
    }
}()

select {
case <-ctx.Done():
    fmt.Println("被取消:", ctx.Err())
case result := <-doWork(ctx):
    fmt.Println("完成:", result)
}
```

### 2.4 设置截止时间

```go
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(ctx, deadline)
defer cancel()
```

## 3. WithValue 链式原理

### 3.1 问题：赋值会覆盖吗？

```go
ctx = context.WithValue(ctx, "user_id", uint64(123))
```

**不会覆盖**，`WithValue` 采用链式包装设计。

### 3.2 链表结构

```go
ctx1 := context.Background()
ctx2 := context.WithValue(ctx1, "user_id", uint64(123))
ctx3 := context.WithValue(ctx2, "request_id", "abc-123")
```

形成链表：

```
ctx3 → ctx2 → ctx1 → Background
 │       │
 │       └── user_id: 123
 └── request_id: "abc-123"
```

### 3.3 查找机制

调用 `ctx.Value(key)` 时，**从当前节点向上遍历**查找：

```go
// ctx3.Value("user_id") 的查找过程：
// 1. ctx3 没有 "user_id" → 继续向上
// 2. ctx2 有 "user_id" → 返回 123
```

### 3.4 同名 key 遮蔽

相同 key 再次 WithValue 会**遮蔽**旧值（不是覆盖）：

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "user_id", uint64(100))
ctx = context.WithValue(ctx, "user_id", uint64(200))

fmt.Println(ctx.Value("user_id")) // 200（找到最近的就返回）
```

### 3.5 设计优势

1. **不可变性**：每次 WithValue 返回新 context，原 context 不变
2. **并发安全**：多个 goroutine 可以基于同一个 ctx 派生不同分支
3. **请求链路追踪**：天然支持父子关系

## 4. 实际场景：HTTP 请求链路追踪

### 4.1 请求流程

```
HTTP请求 → 中间件 → Handler → Service → Repository → 数据库
```

### 4.2 完整示例

```go
// ============ 中间件层 ============
func TraceMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 生成请求唯一标识
        requestID := uuid.New().String()
        startTime := time.Now()

        // 派生新 context，注入追踪信息
        ctx = context.WithValue(ctx, "request_id", requestID)
        ctx = context.WithValue(ctx, "start_time", startTime)

        c.Next(ctx) // 传递给下一层
    }
}

func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        userID, _ := parseToken(c.GetHeader("Authorization"))

        // 继续派生，添加用户信息（不影响上层的 request_id）
        ctx = context.WithValue(ctx, "user_id", userID)

        c.Next(ctx)
    }
}

// ============ Handler 层 ============
func (h *VideoHandler) PublishVideo(ctx context.Context, c *app.RequestContext) {
    // 可以拿到所有上层注入的值
    requestID := ctx.Value("request_id").(string)
    userID := ctx.Value("user_id").(uint64)

    log.Printf("[%s] 用户 %d 开始投稿", requestID, userID)

    // 传递给 Service
    err := h.videoService.Publish(ctx, &req)
}

// ============ Service 层 ============
func (s *VideoService) Publish(ctx context.Context, req *PublishRequest) error {
    requestID := ctx.Value("request_id").(string)
    userID := ctx.Value("user_id").(uint64)

    // 设置数据库操作超时（派生带超时的 context）
    dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    // 传递给 Repository
    video, err := s.videoRepo.Create(dbCtx, userID, req)
    if err != nil {
        log.Printf("[%s] 投稿失败: %v", requestID, err)
        return err
    }

    // 异步处理：转码任务（派生新分支）
    go func() {
        // 新 goroutine 基于原 ctx 派生，不继承超时
        transcodeCtx := context.WithValue(ctx, "task", "transcode")
        s.transcodeVideo(transcodeCtx, video.ID)
    }()

    return nil
}

// ============ Repository 层 ============
func (r *VideoRepository) Create(ctx context.Context, userID uint64, req *PublishRequest) (*Video, error) {
    requestID := ctx.Value("request_id").(string)

    video := &Video{UserID: userID, Title: req.Title}

    // GORM 使用 ctx 控制超时
    err := r.db.WithContext(ctx).Create(video).Error
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            log.Printf("[%s] 数据库操作超时", requestID)
        }
        return nil, err
    }

    log.Printf("[%s] 视频创建成功: %d", requestID, video.ID)
    return video, nil
}
```

### 4.3 日志输出效果

```
[req-abc123] 用户 42 开始投稿
[req-abc123] 视频创建成功: 1001
[req-abc123] 转码任务启动
[req-def456] 用户 18 开始投稿    ← 另一个并发请求，互不干扰
[req-def456] 视频创建成功: 1002
```

### 4.4 Context 链路图

```
Background
    │
    ├── WithValue(request_id: "req-abc123")
    │       │
    │       └── WithValue(user_id: 42)
    │               │
    │               ├── WithTimeout(3s) → 数据库操作
    │               │
    │               └── WithValue(task: "transcode") → 异步转码
    │
    └── WithValue(request_id: "req-def456")  ← 另一个请求，完全独立
            │
            └── WithValue(user_id: 18)
```

## 5. 常用方法速查

| 方法 | 作用 |
|------|------|
| `context.Background()` | 创建根 context（main 函数或初始化时用） |
| `context.TODO()` | 占位符，不确定用什么时临时使用 |
| `context.WithValue(ctx, key, val)` | 附加键值对 |
| `context.WithTimeout(ctx, duration)` | 设置超时 |
| `context.WithDeadline(ctx, time)` | 设置截止时间 |
| `context.WithCancel(ctx)` | 创建可取消的 context |
| `ctx.Done()` | 返回关闭信号的 channel |
| `ctx.Err()` | 返回取消原因（Canceled 或 DeadlineExceeded） |
| `ctx.Deadline()` | 获取截止时间 |

## 6. 最佳实践

1. **ctx 作为第一个参数传递**
2. **不要把 ctx 存到 struct 里**
3. **WithTimeout/WithCancel 后必须 `defer cancel()`**
4. **用 WithValue 传递请求级数据，不要滥用**（如 request_id、user_id）
5. **异步 goroutine 派生独立分支**，避免继承不必要的超时

## 7. 在 Hertz 项目中的应用位置

| 层级 | 文件位置 | 用途 |
|------|----------|------|
| 中间件 | `biz/router/v1/middleware.go` | 注入 request_id、user_id |
| Handler | `biz/handler/v1/*.go` | 读取上下文信息、传递给 Service |
| Service | `pkg/service/*.go` | 设置数据库超时、派生异步任务 |
| Repository | `biz/dal/db/*.go` | GORM WithContext 控制超时 |

## 参考资料

- [Go 官方文档 - context](https://pkg.go.dev/context)
- [Go Blog - Context](https://go.dev/blog/context)
