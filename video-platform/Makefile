APP_NAME := fanone-video

.PHONY: help init tidy build run test e2e docker-build

help:
	@echo "可用命令："
	@echo "  make init         初始化本地开发环境（生成 .env、创建存储目录、下载依赖）"
	@echo "  make tidy         整理并下载 Go 依赖"
	@echo "  make build        构建本地二进制"
	@echo "  make run          启动服务"
	@echo "  make test         运行 video-platform 单元测试"
	@echo "  make e2e          运行 test/ 目录端到端测试"
	@echo "  make docker-build 构建 Docker 镜像"

init:
	@bash script/init.sh

tidy:
	@go mod tidy
	@go mod download

build:
	@go build -o $(APP_NAME) .

run:
	@go run .

test:
	@go test ./...

e2e:
	@cd ../test && go run .

docker-build:
	@docker build -t $(APP_NAME):latest .
