# FanOne 项目待办清单

最后更新：2026-03-31

本文档按“课程最高要求 + 简历可讲述性”整理，目标不是只把接口写完，而是把项目补到能答辩、能展示、能写进简历。

## 一、P0：本周必须完成

### 1. 社交模块功能闭环

- [x] 补齐社交模块 e2e 测试
  - 文件：[test/relation.go](/home/particle/2025-2/west2onlie_GoWeb/work4/test/relation.go)
  - 现状：仍是 TODO，占位未实现
  - 需要覆盖：
    - 关注
    - 重复关注幂等
    - 取关
    - 关注列表
    - 粉丝列表
    - 好友列表
    - 不能关注自己
    - 未登录访问好友列表的行为

- [x] 在测试入口串联社交流程
  - 文件：[test/main.go](/home/particle/2025-2/west2onlie_GoWeb/work4/test/main.go)
  - 目标：把“注册 -> 登录 -> 关注 -> 粉丝 -> 好友 -> 取关”加入完整流程

### 2. 修复关注关系表未迁移问题

- [x] 将 `Follow` 模型加入自动迁移
  - 文件：[video-platform/biz/dal/store.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/dal/store.go)
  - 关联模型：[video-platform/biz/dal/model/follow.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/dal/model/follow.go)
  - 风险：新数据库环境下社交接口会直接失败

### 3. 修复好友列表鉴权缺口

- [x] 给好友列表路由挂载认证中间件
  - 文件：[video-platform/biz/router/v1/middleware.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/router/v1/middleware.go)
  - 问题：`ListFriends` 在 handler 里直接读取 JWT 中的 `user_id`，但路由没有鉴权
  - 关联 handler：[video-platform/biz/handler/v1/relation_handler.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/handler/v1/relation_handler.go)

### 4. 补齐 Docker 交付

- [ ] 新增 `Dockerfile`
  - 目标：支持 `docker build` 和单容器运行
  - 需要说明：
    - `DB_DSN`
    - `REDIS_ADDR`
    - `REDIS_PASSWORD`
    - `REDIS_DB`
    - `JWT_SECRET`

- [ ] 补充运行说明
  - 建议写入 README
  - 至少包含：
    - 本地启动
    - 环境变量
    - Docker 构建
    - Docker 运行

### 5. 补齐项目结构图

- [ ] 新增 README 或 `docs/architecture.md`
  - 课程要求：需要有目录树，答辩时方便讲解
  - 至少包含：
    - 目录结构图
    - 四层职责：router / handler / service / dal
    - 存储目录说明
    - Swagger 路径

## 二、P1：补到“能答辩、能自圆其说”

### 1. 修复 Redis 降级逻辑

- [ ] 让 Redis 连接失败时可降级，而不是直接退出
  - 文件：[video-platform/biz/dal/redis.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/dal/redis.go)
  - 当前问题：
    - 未配置 `REDIS_ADDR` 会 `log.Fatal`
    - Ping 失败也会 `log.Fatal`
  - 目标：
    - Redis 不可用时服务仍可启动
    - 热榜接口回退到 DB 计算
    - 日志中明确标记 Redis 降级

- [ ] 修正 `HasRedis()` 逻辑
  - 文件：[video-platform/biz/dal/store.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/dal/store.go)
  - 当前问题：固定返回 `true`

### 2. 完整验证 17 个接口的协议一致性

- [ ] 对照官方文档逐项核验请求参数、响应结构和错误码
  - 参考：[work4-api.md](/home/particle/2025-2/west2onlie_GoWeb/work4/work4-api.md)
  - 重点核验：
    - 分页字段默认值与上限
    - 评论删除权限
    - 好友列表是否必须登录
    - 搜索条件是否为 AND
    - 热榜是否经过 Redis

### 3. 补日志中间件

- [ ] 增加请求日志
  - 建议位置：[video-platform/biz/router/v1/middleware.go](/home/particle/2025-2/west2onlie_GoWeb/work4/video-platform/biz/router/v1/middleware.go)
  - 目标字段：
    - request_id
    - method
    - path
    - user_id
    - status_code
    - cost_ms

- [ ] 统一 handler 层错误日志格式
  - 要求：遵循 AGENTS.md 中的 `[模块名][操作名] 错误描述 关键参数: %v`
  - 核查对象：
    - 用户模块
    - 视频模块
    - 互动模块
    - 社交模块

### 4. 完整回归测试

- [ ] 每次修改后执行完整 e2e
  - 流程：
    - 启动服务
    - 跑 `test/`
    - 校验所有用例通过

- [ ] 增加异常场景测试
  - 未登录访问受保护接口
  - 参数缺失
  - 非法 ID
  - 重复点赞
  - 重复关注
  - 删除他人评论

## 三、P2：加分项

### 1. 把热榜做得更像真实业务

- [ ] 为 Redis 热榜补齐防击穿策略
  - 随机 TTL
  - 空结果保护
  - rebuild 锁

- [ ] 将热度计算抽成独立函数
  - 当前规则：`like_count * 3 + comment_count * 2 + visit_count`
  - 目标：便于后续扩展播放量、发布时间衰减

### 2. 提升工程化程度

- [ ] 增加 `.env.example` 完整字段注释
- [ ] 补一键初始化脚本
- [ ] 补 `make` 或脚本命令说明
- [ ] 补接口调用示例

### 3. 选做一个高质量 Bonus

- [ ] 分片上传
  - 适合强调文件服务和断点续传设计

- [ ] WebSocket 通知/聊天
  - 适合强调实时通信能力

- [ ] Elasticsearch / OpenSearch
  - 可用于日志检索或视频搜索增强

说明：三选一即可，不建议同时铺太开。

## 四、简历亮点提炼

以下内容建议在项目完成后写入简历，避免先写上去但实际答不出来。

### 1. 可直接写进简历的亮点

- [ ] 基于 Hertz + Protobuf + `hz` 脚手架实现视频平台后端，完成用户、视频、互动、社交四大模块与 17+ 核心 API
- [ ] 基于 JWT 实现双 Token 认证体系，支持 Access Token 鉴权与 Refresh Token 刷新
- [ ] 基于 MySQL + GORM 完成用户、视频、评论、点赞、关注关系建模，支持分页、幂等、软删除与权限控制
- [ ] 基于 Redis 实现热门排行榜缓存，并在点赞、评论写路径上进行热度增量更新
- [ ] 支持头像与视频文件上传，完成静态资源访问链路
- [ ] 提供 Swagger/OpenAPI 文档、e2e 测试和 Docker 化部署能力

### 2. 推荐简历描述

#### 一句话版本

- [ ] 独立完成基于 Hertz/Protobuf/Redis/MySQL 的视频平台后端开发，设计并实现双 Token 认证、热榜缓存、社交关系、文件上传和 17+ 核心 API。

#### 两句话版本

- [ ] 负责视频平台后端架构设计与核心功能开发，基于 Hertz + Protobuf + GORM + Redis 实现用户、视频、互动、社交四大模块，完成注册登录、投稿、搜索、点赞评论、关注好友等 17+ API。
- [ ] 重点解决 JWT 双 Token 鉴权、社交关系建模、幂等写操作、Redis 热榜缓存一致性等问题，并补充 e2e 测试、Swagger 文档和 Docker 化部署。

## 五、答辩与面试要重点讲的难点

### 1. 双 Token 鉴权

- [ ] 为什么要拆成 Access Token 和 Refresh Token
- [ ] 如何控制有效期
- [ ] 为什么刷新接口不能直接复用访问令牌
- [ ] 中间件如何解析并注入用户身份

### 2. 社交关系建模

- [ ] 关注、粉丝、好友三种列表的 SQL 语义不同
- [ ] 好友本质是互相关注，不是单独一张好友表
- [ ] 为什么需要软删除支持“取关后再关注”

### 3. 幂等与事务

- [ ] 重复点赞不能重复加数
- [ ] 重复关注不能重复加数
- [ ] 取消点赞和删除评论时要同步维护统计字段
- [ ] 为什么“关系表 + 计数表”必须放在一个事务里

### 4. 热榜缓存一致性

- [ ] 为什么读缓存、写数据库、增量更新缓存会出现不一致
- [ ] 为什么需要缓存重建机制
- [ ] Redis 不可用时如何降级

### 5. 搜索和分页

- [ ] 搜索为什么必须是 AND 条件
- [ ] 分页为什么要做默认值与上限保护
- [ ] 为什么 `page_num` 必须从 1 开始

## 六、建议执行顺序

1. 先修 `Follow` 迁移和好友列表鉴权，确保社交接口真实可用。
2. 再补 `test/relation.go` 与 `test/main.go`，把社交模块纳入 e2e。
3. 接着补 `Dockerfile`、README、目录结构图，完成课程交付。
4. 然后修 Redis 降级和日志中间件，把项目补到“能讲工程质量”的层级。
5. 最后再选一个 Bonus 做深，不要摊大饼。

## 七、完成定义

满足以下条件时，可认为项目达到了“课程高质量交付 + 简历可写”标准：

- [ ] 17 个核心接口全部实测通过
- [ ] 社交模块 e2e 已补齐
- [ ] Docker 可成功构建并运行
- [ ] README 含项目结构图、启动说明、技术选型
- [ ] Redis 热榜支持缓存与降级
- [ ] 日志、鉴权、分页、权限控制逻辑完整
- [ ] 能用 3 到 5 分钟清楚讲明项目架构、亮点和难点
