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
