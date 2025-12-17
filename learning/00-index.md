# FanOne 视频平台学习笔记索引

## 笔记列表

| 日期 | 主题 | 文件 |
|------|------|------|
| 2025-01-25 | API 设计规范 | [01-api-design.md](./01-api-design.md) |
| 2025-11-25 | Protobuf 设计与字段规范 | [02-protobuf-design.md](./02-protobuf-design.md) |
| 2025-11-25 | Hertz 框架与 hz 脚手架 | [03-hertz-and-hz.md](./03-hertz-and-hz.md) |
| 2025-12-07 | Go Context 用法详解 | [04-go-context.md](./04-go-context.md) |
| 2025-12-15 | 互动模块实现详解 | [05-interaction-module.md](./05-interaction-module.md) |
| 2025-12-16 | 认证与双 Token 机制 | [06-auth-and-jwt.md](./06-auth-and-jwt.md) |
| 2025-12-16 | 事务机制与类型系统 | [07-transaction-and-type-system.md](./07-transaction-and-type-system.md) |
| 2025-12-16 | ID 类型设计：uint vs string | [08-id-type-design.md](./08-id-type-design.md) |
| 2025-12-16 | Go 变量声明与赋值详解 | [09-go-variable-declaration.md](./09-go-variable-declaration.md) |

### 09-go-variable-declaration.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | 问题背景（err 为何用 = 而不是 :=） |
| §2 | Go 短变量声明的三种情况（全新变量、混合变量、全已存在） |
| §3 | 为什么 `err = store.WithTx(...)` 必须用 `=` |
| §4 | `:=` 中已存在变量的行为等价于 `=` ⭐ 核心结论 |
| §5 | Go 语法糖设计原因（优雅的错误处理） |
| §6 | 完整规则总结表 |
| §7 | 实际项目最佳实践（错误处理链、避免遮蔽） |
| §8 | 常见错误与调试（编译错误、变量遮蔽、IDE 工具） |
| §9 | 作用域可视化（图解遮蔽机制） |
| §10 | 关键代码位置 |
| §11 | 推荐阅读 |
| §12 | 总结 |

### 08-id-type-design.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | 核心原因（分层架构职责划分） |
| §2 | JavaScript 精度丢失问题（Number.MAX_SAFE_INTEGER、实际案例） |
| §3 | Protobuf 数值类型的跨语言兼容性（类型对照表、实验对比） |
| §4 | 数据库层使用 uint 的优势（存储效率、索引性能、自增支持） |
| §5 | 业界最佳实践（Twitter Snowflake、Discord、Google APIs） |
| §6 | 当前项目实现（Protobuf 定义、数据库模型、Handler 层转换） |
| §7 | 分层职责划分（API → Handler → Service → 数据库） |
| §8 | 未来扩展能力（UUID/Snowflake 无缝切换） |
| §9 | 常见错误与调试（直接用 uint64、忘记校验、SQL 注入） |
| §10 | 性能对比实测（INT vs BIGINT vs UUID） |
| §11 | 最佳实践总结 |
| §12 | 关键代码位置 |
| §13 | 推荐阅读 |

### 07-transaction-and-type-system.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | GORM 事务机制（WithTx 设计、执行流程、普通/事务 Store 对比、Redis 一致性） |
| §2 | Go 类型系统与 Proto 枚举（类型不匹配问题、为何需要 int32() 转换） |
| §3 | 魔法数字重构（识别坏味道、重构步骤、优势对比） |
| §4 | 关键代码位置 |
| §5 | 最佳实践（事务使用原则、避免的操作、缓存更新策略） |
| §6 | 进阶话题（分布式事务、Redis 事务、类型别名 vs 定义） |
| §7 | 常见错误与调试 |
| §8 | 推荐阅读 |

### 06-auth-and-jwt.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | 问题背景（为何不用 req.UserId） |
| §2 | 两种获取用户 ID 的方式对比（来源、可信度、使用场景） |
| §3 | JWT 中间件实现详解（完整流程、核心代码） |
| §4 | JWT Claims 结构（自定义字段、生成时写入） |
| §5 | 双 Token 机制（Access/Refresh、有效期、刷新流程） |
| §6 | Hertz 上下文存取机制（Set/Get、与标准 Context 区别） |
| §7 | 关键代码位置 |
| §8 | 最佳实践 |
| §9 | 推荐阅读 |

### 05-interaction-module.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | 模块概览（5 个接口、认证要求） |
| §2 | 点赞模块设计（数据模型、事务处理、增量更新、幂等性） |
| §3 | 评论模块设计（发布、删除事务） |
| §4 | 权限校验对比（点赞 vs 删除评论） |
| §5 | Redis 热榜同步 |
| §6 | 数据流总结 |
| §7 | 关键代码位置 |
| §8 | 推荐阅读 |

### 04-go-context.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | Context 概述（ctx 与 c 的区别） |
| §2 | 核心功能（传值、超时、取消、截止时间） |
| §3 | WithValue 链式原理（链表结构、查找机制、遮蔽） |
| §4 | 实际场景：HTTP 请求链路追踪 |
| §5 | 常用方法速查表 |
| §6 | 最佳实践 |
| §7 | 在 Hertz 项目中的应用位置 |

### 02-protobuf-design.md 更新记录

| 日期 | 新增内容 |
|------|----------|
| 2025-11-25 | #7 双Token、#8 身份从Token获取、#9 Followings/Followers区别、#10 点赞命名、#11 评论模块归属、#12 删评论参数精简 |

### 03-hertz-and-hz.md 内容概要

| 章节 | 内容 |
|------|------|
| §1 | Hertz 简介 |
| §2 | hz 脚手架工具（安装、命令、目录结构） |
| §3 | api.proto 详解（HTTP 注解、参数绑定、校验） |
| §4 | 生成的 Handler 结构 |
| §5 | 多模块路由冲突解决方案 |
| §6 | **响应格式规范**（c.String vs c.JSON、统一 JSON 响应、错误码规范） ⭐ 新增 2025-12-16 |
| §7 | 当前项目 API 路由表（19 个接口） |
| §8 | **中间件机制深度解析**（c.Next()、洋葱模型、c.Set/Get、执行顺序） ⭐ 新增 2025-12-16 |

## 待整理主题

- [x] 认证 & 双 Token (`06-auth-and-jwt.md`) ✅ 2025-12-16
- [x] 事务机制 & 类型系统 (`07-transaction-and-type-system.md`) ✅ 2025-12-16
- [x] ID 类型设计 (`08-id-type-design.md`) ✅ 2025-12-16
- [x] Go 变量声明 (`09-go-variable-declaration.md`) ✅ 2025-12-16
- [ ] Redis / 缓存应用 (`10-redis-cache.md`)
- [ ] 社交模块 (`11-relation-module.md`)
