#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
ENV_EXAMPLE="$ROOT_DIR/.env.example"
ENV_FILE="$ROOT_DIR/.env"
STORAGE_DIRS=("$ROOT_DIR/storage/avatars" "$ROOT_DIR/storage/videos")

echo "[init] 项目目录: $ROOT_DIR"

if [[ ! -f "$ENV_EXAMPLE" ]]; then
  echo "[init] 缺少 .env.example，无法初始化" >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  cp "$ENV_EXAMPLE" "$ENV_FILE"
  echo "[init] 已创建 .env，请按实际环境检查数据库和 Redis 配置"
else
  echo "[init] 检测到现有 .env，跳过覆盖"
fi

for dir in "${STORAGE_DIRS[@]}"; do
  mkdir -p "$dir"
  echo "[init] 已确保目录存在: $dir"
done

echo "[init] 下载 Go 依赖"
go mod tidy
go mod download

echo "[init] 初始化完成"
echo "[init] 下一步："
echo "  1. 编辑 $ENV_FILE"
echo "  2. 执行 make run 或 go run ."
echo "  3. 如需端到端测试，先启动服务，再到 ../test 执行 go run ."
