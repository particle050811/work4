# FanOne 视频平台

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
- Swagger：`http://localhost:8888/swagger/index.html`

## Docker 交付

### 1. 构建镜像

在 [video-platform/Dockerfile](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/Dockerfile) 所在目录执行：

```bash
docker build -t fanone-video:latest .
```

### 2. 运行容器

推荐在交付文档里至少保留一个“可直接复制”的启动示例。对于只提交 `.env.example`、不提交真实 `.env` 的场景，建议直接通过系统环境变量启动：

```bash
docker run --rm \
  --name fanone-video \
  -e DB_DSN='root:123456@tcp(172.17.0.1:3306)/fanone?charset=utf8mb4&parseTime=True&loc=Local' \
  -e REDIS_ADDR='172.17.0.1:6379' \
  -e REDIS_PASSWORD='' \
  -e REDIS_DB='0' \
  -e JWT_SECRET='fanone-video-platform-secret-key-2024' \
  -e SERVER_PORT='8888' \
  -p 8888:8888 \
  -v "$(pwd)/storage:/app/storage" \
  fanone-video:latest
```

上面命令中的 `172.17.0.1` 只是示例，实际应替换为容器可访问的 MySQL / Redis 地址。

如果你本地已经准备好了未提交到仓库的 `.env` 文件，也可以使用 `--env-file` 方式运行。

如果本机 MySQL 和 Redis 跑在宿主机上，Linux 建议直接使用宿主机网络：

```bash
docker run --rm \
  --name fanone-video \
  --network host \
  --env-file .env \
  -v "$(pwd)/storage:/app/storage" \
  fanone-video:latest
```

如果你不使用 `--network host`，则要把 `.env` 里的 `DB_DSN`、`REDIS_ADDR` 改成容器可访问的地址，例如宿主机 IP 或 Docker 网络内服务名。

常见做法：

- MySQL 容器名为 `mysql`
- Redis 容器名为 `redis`
- 则 `DB_DSN` 可写为 `root:123456@tcp(mysql:3306)/fanone?charset=utf8mb4&parseTime=True&loc=Local`
- `REDIS_ADDR` 可写为 `redis:6379`

### 3. 端口映射运行

如果你希望显式映射端口，可这样运行：

```bash
docker run --rm \
  --name fanone-video \
  --env-file .env \
  -e SERVER_PORT=8888 \
  -p 8888:8888 \
  -v "$(pwd)/storage:/app/storage" \
  fanone-video:latest
```

此模式要求容器里的数据库和 Redis 地址可达。

### 4. 验证服务

启动后可以检查：

```bash
curl --noproxy localhost -s http://localhost:8888/ping
```

如果返回健康检查结果，说明容器已成功启动。
