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

## 6. 当前项目 API 路由表

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

## 7. 推荐阅读

- [Hertz 官方文档](https://www.cloudwego.io/zh/docs/hertz/)
- [hz 工具使用指南](https://www.cloudwego.io/zh/docs/hertz/tutorials/toolkit/toolkit/)
- [Hertz + Protobuf 示例](https://github.com/cloudwego/hertz-examples)
