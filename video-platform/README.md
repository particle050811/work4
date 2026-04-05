# FanOne 视频平台

## 项目结构与职责

### 目录结构图

```text
video-platform/
├── .env.example               # 环境变量示例
├── main.go                    # 服务入口
├── router.go                  # 自定义路由扩展
├── router_gen.go              # Hertz 生成路由入口
├── api.proto                  # HTTP 注解定义
├── api/video/v1/              # Protobuf 接口定义
│   ├── common.proto
│   ├── user.proto
│   ├── video.proto
│   ├── interaction.proto
│   └── relation.proto
├── biz/
│   ├── router/                # 路由注册与中间件挂载
│   ├── handler/               # HTTP 入参解析、鉴权上下文读取、响应封装
│   ├── service/               # 业务编排、权限校验、分页与事务逻辑
│   ├── dal/                   # 数据访问层，统一管理 DB / Redis / DAO / Model
│   └── model/api/             # Protobuf 生成的 Go 代码
├── pkg/
│   ├── auth/                  # JWT 双 Token、密码哈希
│   ├── middleware/            # 通用中间件
│   ├── response/              # 统一响应工具
│   ├── storage/               # 静态资源访问辅助
│   └── util/                  # 通用工具函数
├── storage/
│   ├── avatars/               # 用户头像上传目录
│   └── videos/                # 视频文件上传目录
├── swagger/
│   ├── user/
│   ├── video/
│   ├── interaction/
│   └── relation/
├── script/                    # 启动/初始化脚本
├── build.sh                   # 本地构建脚本
├── Dockerfile
└── README.md
```

### 分层职责

- `router`：注册各模块路由，挂载认证和其他中间件，控制接口暴露边界。
- `handler`：接收 HTTP 请求，绑定参数，读取上下文中的用户身份，调用 service，并返回统一响应。
- `service`：承接核心业务逻辑，包括权限校验、事务控制、分页处理、幂等判断和缓存更新。
- `dal`：负责数据库与 Redis 访问，包括模型定义、DAO 查询、缓存读写和底层存储初始化。

### 存储目录说明

- `storage/avatars/`：保存用户头像文件。
- `storage/videos/`：保存投稿视频文件。
- Docker 启动时建议挂载整个 `storage/` 目录，避免容器重建后上传文件丢失。

### Swagger 路径

- 用户模块：`http://localhost:8888/swagger/user/index.html`
- 视频模块：`http://localhost:8888/swagger/video/index.html`
- 互动模块：`http://localhost:8888/swagger/interaction/index.html`
- 社交模块：`http://localhost:8888/swagger/relation/index.html`

## 本地启动

1. 复制环境变量文件：

```bash
cp .env.example .env
```

2. 按实际环境修改 `.env` 中的以下配置：

- `DB_DSN`：MySQL 连接串，例如 `root:123456@tcp(127.0.0.1:3306)/fanone?charset=utf8mb4&parseTime=True&loc=Local`
- `REDIS_ADDR`：Redis 地址，例如 `127.0.0.1:6379`
- `REDIS_PASSWORD`：Redis 密码，没有可留空
- `REDIS_DB`：Redis DB 编号，默认 `0`
- `JWT_SECRET`：JWT 密钥
- `SERVER_PORT`：服务监听端口，默认 `8888`

3. 启动服务：

```bash
go run .
```

默认访问地址：

- 服务：`http://localhost:8888`
- Swagger（用户模块）：`http://localhost:8888/swagger/user/index.html`
- Swagger（视频模块）：`http://localhost:8888/swagger/video/index.html`
- Swagger（互动模块）：`http://localhost:8888/swagger/interaction/index.html`
- Swagger（社交模块）：`http://localhost:8888/swagger/relation/index.html`

## 初始化与常用命令

### 一键初始化

仓库已提供初始化脚本 [script/init.sh](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/script/init.sh)，用于完成以下动作：

- 首次复制 `.env.example` 为 `.env`
- 创建 `storage/avatars`、`storage/videos` 目录
- 执行 `go mod tidy` 和 `go mod download`

执行方式：

```bash
bash script/init.sh
```

### Make 命令说明

仓库根目录提供 [Makefile](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/Makefile)，便于统一本地操作：

```bash
make help
```

常用命令如下：

- `make init`：初始化本地开发环境
- `make tidy`：整理并下载依赖
- `make build`：构建 `fanone-video` 二进制
- `make run`：启动服务
- `make test`：运行 `video-platform` 单元测试
- `make e2e`：进入 `../test` 执行端到端测试
- `make docker-build`：构建 Docker 镜像

### 推荐开发流程

```bash
make init
make run
make test
make e2e
```

说明：

- 执行 `make e2e` 前，请先确保服务已经在 `http://localhost:8888` 启动。
- 本项目测试客户端默认禁用代理；手动使用 `curl` 调试本地接口时，也必须带上 `--noproxy localhost`。

## 接口调用示例

以下示例均默认服务运行在 `http://localhost:8888`，并遵循仓库约束，使用 `curl --noproxy localhost` 避免本地代理干扰。

### 1. 用户注册

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/user/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "fanone_demo",
    "password": "123456"
  }'
```

### 2. 用户登录

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "fanone_demo",
    "password": "123456"
  }'
```

登录成功后，从响应体中取出 `access_token` 与 `refresh_token`。

### 3. 刷新令牌

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/user/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "替换为上一步返回的 refresh_token"
  }'
```

### 4. 获取用户信息

```bash
curl --noproxy localhost -s \
  "http://localhost:8888/api/v1/user/info?user_id=1"
```

### 5. 视频投稿

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/video/publish" \
  -H "Authorization: Bearer 替换为 access_token" \
  -F "title=我的第一个视频" \
  -F "description=用于本地联调的投稿样例" \
  -F "file=@./testdata/demo.mp4"
```

### 6. 发布列表

```bash
curl --noproxy localhost -s \
  "http://localhost:8888/api/v1/video/list?user_id=1&page_num=1&page_size=10"
```

### 7. 搜索视频

```bash
curl --noproxy localhost -s \
  "http://localhost:8888/api/v1/video/search?keyword=我的视频&page_num=1&page_size=10&sort_by=latest"
```

### 8. 热门排行榜

```bash
curl --noproxy localhost -s \
  "http://localhost:8888/api/v1/video/hot?page_num=1&page_size=10"
```

### 9. 点赞操作

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/interaction/like" \
  -H "Authorization: Bearer 替换为 access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "1",
    "action_type": 1
  }'
```

说明：

- `action_type=1` 表示点赞
- `action_type=2` 表示取消点赞

### 10. 发表评论

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/interaction/comment" \
  -H "Authorization: Bearer 替换为 access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "1",
    "action_type": 1,
    "content": "这是一条测试评论"
  }'
```

### 11. 关注用户

```bash
curl --noproxy localhost -s \
  -X POST "http://localhost:8888/api/v1/relation/action" \
  -H "Authorization: Bearer 替换为 access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "to_user_id": "2",
    "action_type": 1
  }'
```

说明：

- `action_type=1` 表示关注
- `action_type=2` 表示取关

## Docker 交付

### 1. 构建镜像

在 [video-platform/Dockerfile](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/Dockerfile) 所在目录执行：

```bash
docker build -t fanone-video:latest .
```

### 2. 运行容器

```bash
docker run -d \
  --name fanone-video \
  --restart unless-stopped \
  --network host \
  -e DB_DSN='root:hsr123456@tcp(127.0.0.1:3306)/fanone?charset=utf8mb4&parseTime=True&loc=Local' \
  -e REDIS_ADDR='127.0.0.1:6379' \
  -e REDIS_PASSWORD='' \
  -e REDIS_DB='0' \
  -e JWT_SECRET='fanone-video-platform-secret-key-2024' \
  -e SERVER_PORT='8888' \
  -v /home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/storage:/app/storage \
  particle050811/fanone-video:latest
```

该命令为本地测试示例，使用宿主机网络，适用于本机 MySQL 和 Redis 运行在 `127.0.0.1` 的场景。镜像名使用当前仓库工作流推送到 Docker Hub 的 `particle050811/fanone-video:latest`。

本地测试时挂载 `storage` 目录，便于直接查看上传文件；实际服务器部署如果不需要保留宿主机侧文件，可删除 `-v /home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/storage:/app/storage` 这一行，直接运行容器。

如果使用本地刚构建的镜像，需将命令最后一行的镜像名改为 `fanone-video:latest`。

### 3. 验证服务

启动后可以检查：

```bash
curl --noproxy localhost -s http://localhost:8888/ping
```

如果返回健康检查结果，说明容器已成功启动。
