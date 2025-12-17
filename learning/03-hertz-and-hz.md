# Hertz 框架与 hz 脚手架

> 日期：2025-11-25

## 1. Hertz 简介

Hertz 是字节跳动 CloudWeGo 团队开源的高性能 Go HTTP 框架，专为微服务场景设计。

**核心特点**：
- 高性能：基于 netpoll 网络库
- 易扩展：模块化设计，支持中间件
- 代码生成：配合 hz 工具从 IDL 生成代码

## 2. hz 脚手架工具

hz 是 Hertz 的代码生成工具，支持从 Protobuf 或 Thrift IDL 生成项目骨架。

### 2.1 安装

```bash
go install github.com/cloudwego/hertz/cmd/hz@latest
```

### 2.2 常用命令

```bash
# 初始化新项目
hz new --module <module_name> --idl <proto_file> --proto_path=.

# 更新已有项目（新增 proto 文件时）
hz update --idl <proto_file> --proto_path=.
```

### 2.3 生成的目录结构

```
project/
├── main.go                 # 入口
├── router.go               # 自定义路由
├── router_gen.go           # 生成的路由注册
├── biz/
│   ├── handler/            # HTTP 处理器（业务逻辑写这里）
│   ├── model/              # 生成的 Protobuf Go 结构体
│   └── router/             # 生成的路由定义
```

## 3. api.proto 详解

### 3.1 为什么需要 api.proto？

标准 Protobuf 只能描述数据结构和 RPC 接口，无法表达 HTTP 路由信息。`api.proto` 通过 Protobuf 的 `extend` 机制扩展了选项，让我们能在 proto 文件中声明 HTTP 相关信息。

### 3.2 核心注解

#### HTTP 方法注解（用于 rpc 方法）

```protobuf
service UserService {
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (.api.get) = "/api/v1/user/login";    // GET 请求
    option (.api.post) = "/api/v1/user/login";   // POST 请求
    option (.api.put) = "/api/v1/user/:id";      // PUT 请求
    option (.api.delete) = "/api/v1/user/:id";   // DELETE 请求
  }
}
```

#### 参数绑定注解（用于 message 字段）

```protobuf
message LoginRequest {
  // 从请求体 JSON 获取
  string username = 1 [(.api.body) = "username"];

  // 从 URL query 参数获取 (?user_id=xxx)
  string user_id = 2 [(.api.query) = "user_id"];

  // 从请求头获取
  string token = 3 [(.api.header) = "Authorization"];

  // 从表单获取（multipart/form-data）
  string file = 4 [(.api.form) = "file"];

  // 从 URL 路径参数获取 (/user/:id)
  string id = 5 [(.api.path) = "id"];
}
```

#### 参数校验注解

```protobuf
message RegisterRequest {
  string username = 1 [
    (.api.body) = "username",
    (.api.vd) = "len($) > 0 && len($) < 50"  // 验证规则
  ];
  string password = 2 [
    (.api.body) = "password",
    (.api.vd) = "len($) >= 6"
  ];
}
```

### 3.3 注意：点号前缀

在有 `package` 声明的 proto 文件中，注解需要使用 `(.api.xxx)` 格式（带点号前缀），否则会被解析到当前包下导致找不到定义。

```protobuf
// 错误：会被解析为 fanone.api.body
string username = 1 [(api.body) = "username"];

// 正确：使用全局作用域
string username = 1 [(.api.body) = "username"];
```

**参考文件**：`video-platform/api.proto:9-31`

## 4. 生成的 Handler 结构

hz 生成的 handler 函数骨架：

```go
// biz/handler/v1/user_service.go:15-27
func Register(ctx context.Context, c *app.RequestContext) {
    var err error
    var req v1.RegisterRequest
    err = c.BindAndValidate(&req)  // 自动绑定参数并校验
    if err != nil {
        c.String(consts.StatusBadRequest, err.Error())
        return
    }

    resp := new(v1.RegisterResponse)
    // TODO: 在这里实现业务逻辑

    c.JSON(consts.StatusOK, resp)
}
```

## 5. 多模块路由冲突问题

### 问题描述

当项目有多个 proto 文件（user.proto、video.proto 等），分别执行 `hz update` 会在同一个包下生成多个同名的 `Register` 函数，导致编译错误。

### 解决方案

手动将各模块的 `Register` 函数重命名：

```go
// biz/router/v1/user.go
func RegisterUser(r *server.Hertz) { ... }

// biz/router/v1/video.go
func RegisterVideo(r *server.Hertz) { ... }
```

然后在 `biz/router/register.go` 中统一调用：

```go
func GeneratedRegister(r *server.Hertz) {
    v1.RegisterUser(r)
    v1.RegisterVideo(r)
    v1.RegisterInteraction(r)
    v1.RegisterRelation(r)
}
```

**参考文件**：`video-platform/biz/router/register.go:11-17`

## 6. 响应格式规范：c.String vs c.JSON

### 6.1 问题背景

hz 脚手架默认生成的 handler 使用 `c.String()` 和 `c.JSON()` 混合返回响应，但这**不符合 FanOne API 规范**。

### 6.2 两种响应方式对比

#### 方式 A：hz 默认生成（❌ 不推荐）

```go
// biz/handler/v1/relation_service.go:20-21
if err != nil {
    c.String(consts.StatusBadRequest, err.Error())  // 返回纯文本
    return
}

resp := new(v1.RelationActionResponse)
c.JSON(consts.StatusOK, resp)  // 返回空 JSON 对象
```

**HTTP 响应示例：**
```
# 错误时
HTTP/1.1 400 Bad Request
Content-Type: text/plain

参数错误
```

```json
// 成功时
{}
```

**问题：**
1. `c.String()` 返回**纯文本**（Content-Type: text/plain），前端无法解析为 JSON
2. 缺少统一的 `base` 响应结构（status_code、status_msg）
3. 不符合 [FanOne 官方 API 规范](https://doc.west2.online/)

#### 方式 B：统一 JSON 响应（✅ 推荐）

```go
// biz/handler/v1/interaction_service.go:74-77
if video == nil {
    c.JSON(consts.StatusNotFound, &v1.VideoLikeActionResponse{
        Base: response.NotFound("视频不存在"),
    })
    return
}
```

**HTTP 响应示例：**
```json
{
  "base": {
    "status_code": 40004,
    "status_msg": "视频不存在"
  }
}
```

**优点：**
1. 始终返回 JSON 格式
2. 包含统一的 `base` 结构
3. 符合官方 API 规范，前端可统一处理

### 6.3 响应内容示例对比

| 场景 | c.String() | c.JSON() + base |
|------|-----------|-----------------|
| 参数错误 | `"参数错误"` (纯文本) | `{"base": {"status_code": 40000, "status_msg": "参数错误"}}` |
| 资源不存在 | `"视频不存在"` | `{"base": {"status_code": 40004, "status_msg": "视频不存在"}}` |
| 成功无数据 | `{}` | `{"base": {"status_code": 0, "status_msg": "success"}}` |
| 成功有数据 | 无法表达 | `{"base": {...}, "data": {...}}` |

### 6.4 为什么必须用统一 JSON 格式？

根据官方规范要求：

1. **所有响应必须包含 `base` 字段**
   ```protobuf
   // api/video/v1/common.proto:5-8
   message BaseResponse {
     int32 status_code = 1;  // 0 表示成功，其他为错误码
     string status_msg = 2;   // 状态描述
   }
   ```

2. **错误码规范**（参考 `pkg/response/response.go:9-14`）
   ```go
   const (
       CodeSuccess      = 0      // 成功
       CodeParamError   = 40000  // 参数错误
       CodeUnauthorized = 40100  // 未授权
       CodeForbidden    = 40300  // 无权限
       CodeNotFound     = 40004  // 资源不存在
       CodeInternal     = 50000  // 内部错误
   )
   ```

3. **前端依赖统一格式**
   ```javascript
   // 前端代码
   if (response.base.status_code === 0) {
     // 成功处理
   } else {
     // 错误提示：response.base.status_msg
   }
   ```

### 6.5 正确的响应模式

#### 错误响应
```go
c.JSON(consts.StatusBadRequest, &v1.XxxResponse{
    Base: response.ParamError("具体错误信息"),
})
```

#### 成功响应（无额外数据）
```go
c.JSON(consts.StatusOK, &v1.XxxResponse{
    Base: response.Success("操作成功"),
})
```

#### 成功响应（有数据）
```go
c.JSON(consts.StatusOK, &v1.ListFollowingsResponse{
    Base: response.Success(),
    Data: &v1.UserListWithTotal{
        Items: items,
        Total: total,
    },
})
```

### 6.6 关键代码位置

- **响应工具函数**：`pkg/response/response.go:16-52`
- **正确示例**：`biz/handler/v1/interaction_service.go:74-77`（点赞接口）
- **需要修改的文件**：`biz/handler/v1/relation_service.go`（社交模块）

### 6.7 修改建议

将所有 hz 生成的 handler 中的响应代码改为统一格式：

```diff
- c.String(consts.StatusBadRequest, err.Error())
+ c.JSON(consts.StatusBadRequest, &v1.XxxResponse{
+     Base: response.ParamError(err.Error()),
+ })

- resp := new(v1.XxxResponse)
- c.JSON(consts.StatusOK, resp)
+ c.JSON(consts.StatusOK, &v1.XxxResponse{
+     Base: response.Success("操作成功"),
+ })
```

## 7. 当前项目 API 路由表

| 模块 | 路由 | 方法 | 处理器 |
|------|------|------|--------|
| 用户 | `/api/v1/user/register` | POST | Register |
| 用户 | `/api/v1/user/login` | POST | Login |
| 用户 | `/api/v1/user/refresh` | POST | RefreshToken |
| 用户 | `/api/v1/user/info` | GET | GetUserInfo |
| 用户 | `/api/v1/user/avatar` | POST | UploadAvatar |
| 视频 | `/api/v1/video/publish` | POST | PublishVideo |
| 视频 | `/api/v1/video/list` | GET | ListPublishedVideos |
| 视频 | `/api/v1/video/search` | GET | SearchVideos |
| 视频 | `/api/v1/video/comments` | GET | ListVideoComments |
| 视频 | `/api/v1/video/hot` | GET | GetHotVideos |
| 互动 | `/api/v1/interaction/like` | POST | VideoLikeAction |
| 互动 | `/api/v1/interaction/like/list` | GET | ListLikedVideos |
| 互动 | `/api/v1/interaction/comment` | POST | PublishComment |
| 互动 | `/api/v1/interaction/comment/list` | GET | ListUserComments |
| 互动 | `/api/v1/interaction/comment/delete` | POST | DeleteComment |
| 社交 | `/api/v1/relation/action` | POST | RelationAction |
| 社交 | `/api/v1/relation/following/list` | GET | ListFollowings |
| 社交 | `/api/v1/relation/follower/list` | GET | ListFollowers |
| 社交 | `/api/v1/relation/friend/list` | GET | ListFriends |

## 8. Hertz 中间件机制深度解析

> 新增日期：2025-12-16

### 8.1 核心问题：为什么需要 `c.Next(ctx)`？

**问题**：函数结束后不应该自动执行下一个函数吗？为什么还需要显式调用 `c.Next(ctx)`？

**答案**：这是洋葱模型设计的核心，`c.Next()` 提供了**精确控制权**，让中间件能够：
1. 在下一个函数**执行前**做预处理
2. 在下一个函数**执行后**做后处理
3. **条件性地决定**是否继续执行

### 8.2 如果不写 `c.Next(ctx)` 会怎样？

请求链会**中断**，后续所有处理函数（包括实际的业务 handler）都不会执行。

#### 示例 A：正常流程 ✅

```go
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "未登录"})
            return  // 不调用 c.Next()，请求到此为止
        }

        claims := parseToken(token)
        c.Set("user_id", claims.UserID)

        c.Next(ctx)  // ✅ 继续执行 RelationAction
    }
}

func RelationAction(ctx context.Context, c *app.RequestContext) {
    // ✅ 会被执行
    userID := c.GetInt64("user_id")
    // 业务逻辑...
    c.JSON(200, resp)
}
```

#### 示例 B：忘记 c.Next() ❌

```go
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        token := c.GetHeader("Authorization")
        claims := parseToken(token)
        c.Set("user_id", claims.UserID)

        // ❌ 忘记写 c.Next(ctx)
    }
}

func RelationAction(ctx context.Context, c *app.RequestContext) {
    // ❌ 永远不会执行！客户端收到空响应或超时
}
```

### 8.3 为什么需要显式调用而不是自动执行？

#### 原因 1：支持前置/后置处理（洋葱模型）

```go
func LoggerMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        start := time.Now()

        c.Next(ctx)  // ← 在这里执行后续逻辑

        // 等后续全部执行完，再记录耗时
        duration := time.Since(start)
        log.Printf("请求耗时: %v", duration)
    }
}
```

如果是自动执行，就无法实现"先放行，执行完再做统计"的模式。

#### 原因 2：支持条件性中断

```go
func CacheMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        cached := getCache(key)
        if cached != nil {
            c.JSON(200, cached)
            return  // 命中缓存，不需要执行后续业务逻辑
        }

        c.Next(ctx)  // 未命中才执行

        setCache(key, response)  // 执行完后保存缓存
    }
}
```

#### 原因 3：支持多中间件串联

```go
请求流程：
[中间件A 开始]
    ↓
  c.Next() → [中间件B 开始]
                ↓
              c.Next() → [Handler 执行]
                ↓
              [中间件B 继续]  // 可在这里做清理
    ↓
[中间件A 继续]  // 可在这里做清理
```

### 8.4 洋葱模型可视化

```
┌─────────────────────────────┐
│  LoggerMiddleware (开始计时) │
│  ┌─────────────────────────┐│
│  │  AuthMiddleware (验证)  ││
│  │  ┌─────────────────────┐││
│  │  │  RateLimitMiddleware│││
│  │  │  ┌─────────────────┐│││
│  │  │  │  Handler (业务) ││││
│  │  │  └─────────────────┘│││
│  │  │  RateLimit 后处理   │││
│  │  └─────────────────────┘││
│  │  Auth 后处理             ││
│  └─────────────────────────┘│
│  Logger 记录耗时             │
└─────────────────────────────┘
```

### 8.5 `c.Set()` 与 `c.Get()` - 上下文传递机制

#### 为什么需要 `c.Set()` 保存中间变量？

`c` (RequestContext) 不仅是返回结果的容器，更是**请求生命周期的数据中心**。

```go
// 在中间件中保存
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        claims := parseToken(...)
        c.Set("user_id", claims.UserID)     // 保存到上下文
        c.Set("username", claims.Username)
        c.Next(ctx)
    }
}

// 在 handler 中取出
func RelationAction(ctx context.Context, c *app.RequestContext) {
    userID := c.GetInt64("user_id")      // 不需要再解析 token！
    username := c.GetString("username")
    // 使用用户信息...
    c.JSON(200, resp)  // 这里才是返回结果
}
```

**作用**：避免在每个 handler 中重复解析 token，中间件解析一次，后续都能用。

#### `c` 的三种用途

| 功能 | 方法 | 说明 |
|------|------|------|
| **请求数据获取** | `c.GetHeader()`, `c.Query()`, `c.Param()` | 获取请求参数 |
| **上下文传递** | `c.Set()`, `c.Get()` | 中间件之间传递数据 |
| **响应输出** | `c.JSON()`, `c.String()` | 返回响应给客户端 |

### 8.6 中间件注册的两种方式

#### 方式 1：单个中间件函数

```go
// pkg/middleware/auth.go:15
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 验证逻辑...
    }
}
```

这是**中间件工厂函数**，返回一个处理函数，方便复用。

#### 方式 2：路由级中间件配置

```go
// biz/router/v1/middleware.go:130-133
func _relationactionMw() []app.HandlerFunc {
    return []app.HandlerFunc{middleware.AuthMiddleware()}
}
```

这是**路由级配置**，返回一个中间件数组，支持多个中间件串联：

```go
func _complexMw() []app.HandlerFunc {
    return []app.HandlerFunc{
        middleware.LoggerMiddleware(),
        middleware.RateLimitMiddleware(),
        middleware.AuthMiddleware(),  // 按顺序执行
    }
}
```

### 8.7 路由注册的实际流程

```go
// biz/router/v1/relation.go:26
_relation.POST("/action", append(_relationactionMw(), v1.RelationAction)...)
```

这行代码做了什么：
1. `_relationactionMw()` 返回 `[]app.HandlerFunc{AuthMiddleware()}`
2. `append(..., v1.RelationAction)` 把业务函数追加到中间件数组后
3. 最终形成：`[AuthMiddleware函数, RelationAction函数]`
4. 请求按顺序执行这两个函数

### 8.8 执行顺序总结

```go
客户端请求 POST /api/v1/relation/action
  ↓
AuthMiddleware 开始
  ├─ 解析 token
  ├─ 如果失败：c.JSON(401, ...) + return (不调用 c.Next)
  ├─ 如果成功：c.Set("user_id", ...) + c.Next(ctx)
  │    ↓
  │  RelationAction 开始
  │    ├─ userID := c.GetInt64("user_id")
  │    ├─ 执行关注/取关业务逻辑
  │    └─ c.JSON(200, resp)
  │    ↓
  │  RelationAction 结束
  │    ↓
  └─ c.Next(ctx) 返回
AuthMiddleware 结束
  ↓
响应返回客户端
```

### 8.9 对比其他框架

| 框架 | 是否需要显式调用 | 调用方式 |
|------|----------------|----------|
| Gin (Go) | ✅ | `c.Next()` |
| Hertz (Go) | ✅ | `c.Next(ctx)` |
| Express.js (Node.js) | ✅ | `next()` |
| Django (Python) | ❌ | process_request/response 自动分离 |
| Spring Boot (Java) | ✅ | `chain.doFilter(request, response)` |

### 8.10 关键代码位置

- **认证中间件实现**：`pkg/middleware/auth.go:15-60`
- **路由级中间件配置**：`biz/router/v1/middleware.go:130-133`
- **路由注册逻辑**：`biz/router/v1/relation.go:26`
- **中间件使用示例**：`biz/router/v1/middleware.go:30-33`（上传头像）、`76-78`（投稿）

### 8.11 最佳实践

1. **必须写 `c.Next(ctx)`**：除非你要主动中断请求
2. **用 `c.Set()` 传递数据**：避免在每个 handler 中重复解析
3. **利用洋葱模型**：在 `c.Next()` 前做预处理，后做清理
4. **明确中断时机**：验证失败时 `c.JSON() + return`，不调用 `c.Next()`
5. **中间件可组合**：返回 `[]app.HandlerFunc` 支持多个中间件串联

### 8.12 常见错误

❌ **忘记 c.Next()**
```go
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        c.Set("user_id", 123)
        // ❌ 忘了 c.Next(ctx)，handler 不会执行
    }
}
```

❌ **验证失败后仍调用 c.Next()**
```go
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        if !valid {
            c.JSON(401, "未授权")
            c.Next(ctx)  // ❌ 不应该继续
        }
    }
}
```

✅ **正确写法**
```go
func AuthMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        if !valid {
            c.JSON(401, "未授权")
            c.Abort()  // 或者直接 return
            return
        }
        c.Set("user_id", 123)
        c.Next(ctx)  // 验证通过才继续
    }
}
```

## 9. 推荐阅读

- [Hertz 官方文档](https://www.cloudwego.io/zh/docs/hertz/)
- [hz 工具使用指南](https://www.cloudwego.io/zh/docs/hertz/tutorials/toolkit/toolkit/)
- [Hertz + Protobuf 示例](https://github.com/cloudwego/hertz-examples)
- [Hertz 中间件开发指南](https://www.cloudwego.io/zh/docs/hertz/tutorials/basic-feature/middleware/)
