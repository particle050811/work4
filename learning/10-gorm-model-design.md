# GORM 模型设计详解

> 日期：2025-12-23
> 主题：GORM 模型定义、TableName 接口、联合唯一索引、JSON 标签

---

## 1. 问题背景

在设计社交模块的 `Follow` 模型时，需要理解以下概念：
- `TableName()` 方法的作用
- 联合唯一索引的使用
- JSON 标签中 `omitempty` 的含义

---

## 2. TableName() 接口

### 2.1 作用

`TableName()` 是 GORM 的约定接口，用于指定模型对应的数据库表名：

```go
// video-platform/biz/dal/model/video.go:25-27
func (Video) TableName() string {
    return "videos"
}
```

### 2.2 使用场景

| 场景 | 示例代码 | 说明 |
|------|----------|------|
| 自动迁移 | `db.AutoMigrate(&model.Video{})` | GORM 读取 TableName() 确定表名 |
| 查询操作 | `db.Model(&model.Video{}).Where(...)` | 知道要操作哪张表 |
| 手写 SQL | `tx.Where("videos.title LIKE ?", kw)` | 需与 TableName() 一致 |

### 2.3 实际代码位置

```go
// video-platform/biz/dal/store.go:79 - 自动迁移
&model.Video{},

// video-platform/biz/dal/db/video_dao.go:42 - 查询
store.DB().Model(&model.Video{})

// video-platform/biz/dal/db/video_dao.go:47 - JOIN 语句
tx = tx.Where("(videos.title LIKE ? OR videos.description LIKE ?)", kw, kw)
tx = tx.Joins("JOIN users ON users.id = videos.user_id")
```

### 2.4 为什么要显式定义

如果不定义，GORM 会自动推导（`Video` → `videos`），但显式定义有以下优势：

| 优势 | 说明 |
|------|------|
| 明确性 | 代码更清晰，不依赖 GORM 命名规则 |
| 灵活性 | 可指定非常规表名（如 `t_video`、`tbl_videos`） |
| 一致性 | 确保 JOIN 等手写 SQL 与 GORM 操作一致 |

---

## 3. 联合唯一索引

### 3.1 语法

```go
type Follow struct {
    FollowerID  uint `gorm:"uniqueIndex:idx_follower_following;index;not null"`
    FollowingID uint `gorm:"uniqueIndex:idx_follower_following;index;not null"`
}
```

**关键点**：两个字段使用**相同的索引名** `idx_follower_following`，GORM 会将它们合并为一个联合索引。

### 3.2 生成的 SQL

```sql
CREATE UNIQUE INDEX idx_follower_following ON follows(follower_id, following_id);
```

### 3.3 作用

防止同一用户重复关注同一个人：

```
✅ (follower=1, following=2) -- 允许
✅ (follower=1, following=3) -- 允许（同一关注者，不同被关注者）
✅ (follower=2, following=2) -- 允许（不同关注者）
❌ (follower=1, following=2) -- 拒绝，已存在
```

### 3.4 对比：独立唯一索引 vs 联合唯一索引

```go
// ❌ 错误：独立唯一索引
FollowerID  uint `gorm:"uniqueIndex"` // 每个 follower_id 只能出现一次
FollowingID uint `gorm:"uniqueIndex"` // 每个 following_id 只能出现一次
// 结果：一个用户只能关注一个人，一个人只能被一个人关注

// ✅ 正确：联合唯一索引
FollowerID  uint `gorm:"uniqueIndex:idx_follower_following"`
FollowingID uint `gorm:"uniqueIndex:idx_follower_following"`
// 结果：(follower_id, following_id) 组合唯一
```

### 3.5 类似设计：VideoLike

```go
// video-platform/biz/dal/model/like.go:10-16
type VideoLike struct {
    UserID  uint `gorm:"uniqueIndex:idx_user_video;not null"`
    VideoID uint `gorm:"uniqueIndex:idx_user_video;index;not null"`
}
```

同样使用联合唯一索引，确保同一用户不能重复点赞同一视频。

---

## 4. JSON 标签详解

### 4.1 基本语法

```go
DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
```

| 部分 | 含义 |
|------|------|
| `deleted_at` | JSON 序列化时的字段名 |
| `omitempty` | 选项：值为空时省略该字段 |

### 4.2 omitempty 效果

```go
type User struct {
    Name      string     `json:"name"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

```json
// DeletedAt 有值时
{"name": "张三", "deleted_at": "2025-01-01T00:00:00Z"}

// DeletedAt 为 nil 时（omitempty 生效）
{"name": "张三"}
```

### 4.3 常用 JSON 标签选项

| 选项 | 作用 | 示例 |
|------|------|------|
| `omitempty` | 空值时省略 | `json:"name,omitempty"` |
| `-` | 完全忽略 | `json:"-"` |
| `string` | 数值转字符串 | `json:"id,string"` |

---

## 5. GORM 标签速查

### 5.1 常用标签

| 标签 | 作用 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `not null` | 非空约束 | `gorm:"not null"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `uniqueIndex:name` | 命名联合唯一索引 | `gorm:"uniqueIndex:idx_a_b"` |
| `index` | 普通索引 | `gorm:"index"` |
| `default` | 默认值 | `gorm:"default:0"` |
| `size` | 字段长度 | `gorm:"size:255"` |
| `type` | 指定类型 | `gorm:"type:varchar(100)"` |

### 5.2 多标签组合

使用分号 `;` 分隔多个标签：

```go
FollowerID uint `gorm:"uniqueIndex:idx_follower_following;index;not null"`
//                     │                                    │     │
//                     联合唯一索引                          单独索引  非空
```

---

## 6. Follow 模型完整设计

```go
// video-platform/biz/dal/model/follow.go
package model

import (
    "time"
    "gorm.io/gorm"
)

// Follow 用户关注关系
type Follow struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    FollowerID  uint           `gorm:"uniqueIndex:idx_follower_following;index;not null" json:"follower_id"`
    FollowingID uint           `gorm:"uniqueIndex:idx_follower_following;index;not null" json:"following_id"`
    CreatedAt   time.Time      `json:"created_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Follow) TableName() string {
    return "follows"
}
```

### 6.1 索引说明

| 索引 | 类型 | 用途 |
|------|------|------|
| `idx_follower_following` | 联合唯一 | 防止重复关注 |
| `follower_id` 单独索引 | 普通 | 查询"我关注了谁" |
| `following_id` 单独索引 | 普通 | 查询"谁关注了我" |
| `deleted_at` 索引 | 普通 | 加速软删除查询 |

### 6.2 查询场景

```go
// 关注列表：我关注了谁
db.Where("follower_id = ?", myID).Find(&follows)

// 粉丝列表：谁关注了我
db.Where("following_id = ?", myID).Find(&follows)

// 好友列表：互相关注（需要双向查询或子查询）
```

---

## 7. 软删除机制

### 7.1 gorm.DeletedAt vs time.Time

`gorm.DeletedAt` 和 `time.Time` 是两种不同的时间类型，各有用途：

#### 类型定义

```go
// time.Time - Go 标准库时间类型
CreatedAt time.Time  // 存储具体时间点，零值是 0001-01-01

// gorm.DeletedAt - GORM 软删除专用类型
// 底层定义：type DeletedAt sql.NullTime
DeletedAt gorm.DeletedAt  // 可空时间，专为软删除设计
```

#### 核心区别

| 特性 | `time.Time` | `gorm.DeletedAt` |
|------|-------------|------------------|
| **可空性** | 不可空（零值是 `0001-01-01`） | 可空（底层是 `sql.NullTime`） |
| **用途** | 记录具体时间（创建、更新） | 软删除标记 |
| **GORM 行为** | 无特殊处理 | 自动过滤已删除记录 |
| **数据库存储** | `datetime NOT NULL` | `datetime NULL` |
| **零值含义** | 表示一个具体时间点 | 表示"未删除" |

#### 为什么 DeletedAt 必须可空

```go
// ❌ 如果用 time.Time
DeletedAt time.Time  // 零值 0001-01-01 仍是一个"有效"时间
// 无法区分"未删除"和"删除于某时间"

// ✅ 使用 gorm.DeletedAt
DeletedAt gorm.DeletedAt  // NULL = 未删除，有值 = 已删除
// 清晰区分两种状态
```

### 7.2 为什么使用软删除

```go
DeletedAt gorm.DeletedAt `gorm:"index"`
```

| 场景 | 硬删除 | 软删除 |
|------|--------|--------|
| 取关后再关注 | 需要重新插入 | 恢复 deleted_at = NULL |
| 数据审计 | 无法追溯 | 保留历史记录 |
| 误操作恢复 | 无法恢复 | 可以恢复 |

### 7.3 GORM 软删除行为

```go
// 删除（设置 deleted_at）
db.Delete(&follow)
// UPDATE follows SET deleted_at = NOW() WHERE id = ?

// 查询（自动过滤已删除）
db.Find(&follows)
// SELECT * FROM follows WHERE deleted_at IS NULL

// 查询包含已删除
db.Unscoped().Find(&follows)
// SELECT * FROM follows

// 恢复软删除记录
db.Unscoped().Model(&follow).Update("deleted_at", nil)
```

---

## 8. 关键代码位置

| 文件 | 行号 | 内容 |
|------|------|------|
| [model/video.go](../video-platform/biz/dal/model/video.go#L25-L27) | 25-27 | Video TableName() |
| [model/like.go](../video-platform/biz/dal/model/like.go#L10-L20) | 10-20 | VideoLike 联合唯一索引 |
| [model/follow.go](../video-platform/biz/dal/model/follow.go) | 全文件 | Follow 模型定义 |
| [store.go](../video-platform/biz/dal/store.go#L79) | 79 | AutoMigrate 注册模型 |
| [video_dao.go](../video-platform/biz/dal/db/video_dao.go#L47) | 47 | SQL 中使用表名 |

---

## 9. 推荐阅读

- [GORM 模型定义](https://gorm.io/zh_CN/docs/models.html)
- [GORM 索引](https://gorm.io/zh_CN/docs/indexes.html)
- [GORM 软删除](https://gorm.io/zh_CN/docs/delete.html#%E8%BD%AF%E5%88%A0%E9%99%A4)
- [Go JSON 标签](https://pkg.go.dev/encoding/json#Marshal)

---

## 10. 总结

| 概念 | 要点 |
|------|------|
| TableName() | GORM 约定接口，指定表名，用于迁移和查询 |
| 联合唯一索引 | 多字段使用相同索引名，确保组合唯一 |
| omitempty | JSON 选项，空值时省略字段 |
| 软删除 | 使用 DeletedAt 字段，便于恢复和审计 |
