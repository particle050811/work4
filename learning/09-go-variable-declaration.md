# Go 语言变量声明与赋值详解

**日期**: 2025-12-16
**主题**: Go 短变量声明（`:=`）与赋值（`=`）的区别与最佳实践

---

## § 1. 问题背景

在实现互动模块点赞功能时，发现以下代码：

```go
// video-platform/biz/handler/v1/interaction_service.go:85-88
videoID, err := util.ParseUint(req.VideoId)  // err 首次声明
if err != nil {
    // 错误处理
}

err = store.WithTx(func(txStore *dal.Store) error {  // 为什么这里用 = 而不是 :=？
    // 事务逻辑
})
```

**核心疑问**：
- 为什么前面用 `:=` 声明 `err`，后面却用 `=` 赋值？
- 什么时候必须用 `=`，什么时候可以用 `:=`？

---

## § 2. Go 短变量声明的三种情况

### 2.1 情况一：所有变量都是新的

```go
a, err := foo()  // err 第一次声明
```

- **行为**：声明并初始化 `a` 和 `err`
- **等价于**：
  ```go
  var a TypeOfA
  var err error
  a, err = foo()
  ```

### 2.2 情况二：至少有一个新变量

```go
a, err := foo()   // err 第一次声明
b, err := bar()   // b 是新变量，err 已存在
```

- **行为**：声明 `b`，**重新赋值** `err`（相当于 `err = ...`）
- **等价于**：
  ```go
  var a TypeOfA
  var err error
  a, err = foo()

  var b TypeOfB
  b, err = bar()  // err 这里是赋值，不是重新声明
  ```

### 2.3 情况三：所有变量都已存在（编译错误）

```go
a, err := foo()
a, err := bar()  // ❌ 编译错误：no new variables on left side of :=
```

- **正确写法**：
  ```go
  a, err := foo()
  a, err = bar()   // ✅ 必须用 =
  ```

---

## § 3. 为什么 `err = store.WithTx(...)` 必须用 `=`？

### 代码分析

```go
// video-platform/biz/handler/v1/interaction_service.go:70-88
videoID, err := util.ParseUint(req.VideoId)  // ① err 首次声明
if err != nil {
    c.JSON(consts.StatusBadRequest, &v1.VideoLikeActionResponse{
        Base: response.ParamError("video_id 格式错误"),
    })
    return
}

// ... 其他逻辑 ...

// ② 只有 err 一个变量需要赋值，必须用 =
err = store.WithTx(func(txStore *dal.Store) error {
    // 事务逻辑
    return nil
})

// ③ 后续检查事务执行结果
if err != nil {
    log.Printf("[互动模块][点赞操作] 事务执行失败: %v", err)
    // ...
}
```

### 为什么不能用 `:=`？

```go
// ❌ 错误写法
err := store.WithTx(func(txStore *dal.Store) error {
    return nil
})
// 编译错误：no new variables on left side of :=
```

**原因**：
1. `err` 已经在第 85 行通过 `videoID, err := util.ParseUint(...)` 声明过
2. 第 88 行只有 `err` 一个变量，且已存在
3. 根据规则 2.3，必须使用 `=` 进行赋值

---

## § 4. `:=` 中已存在变量的行为等价于 `=`

### 关键结论

**当 `:=` 左侧至少有一个新变量时，已存在的变量会被重新赋值（等价于 `=`）**

```go
a, err := foo()  // err 第一次声明
b, err := bar()  // err 等价于用 = 赋值 ✅
```

### 编译器实际处理

```go
// 源代码
a, err := foo()
b, err := bar()

// 编译器处理类似于
var a TypeOfA
var err error
a, err = foo()

var b TypeOfB
b, err = bar()  // err 这里是赋值操作（=）
```

---

## § 5. 为什么这样设计？Go 的语法糖

这是 Go 为了**方便错误处理**而设计的语法糖：

### 不允许重用 err 的情况（繁琐）

```go
result1, err1 := operation1()
if err1 != nil {
    return err1
}

result2, err2 := operation2()
if err2 != nil {
    return err2
}

result3, err3 := operation3()
if err3 != nil {
    return err3
}
```

### 允许重用 err 的情况（优雅）

```go
result1, err := operation1()
if err != nil {
    return err
}

result2, err := operation2()  // err 被重新赋值
if err != nil {
    return err
}

result3, err := operation3()  // err 被重新赋值
if err != nil {
    return err
}
```

---

## § 6. 完整规则总结表

| 情况 | 语法 | 说明 | 示例 |
|------|------|------|------|
| 所有变量都是新的 | `a, err := foo()` | 正常声明 | `videoID, err := util.ParseUint(...)` |
| 至少有一个新变量 | `b, err := bar()` | 新变量声明，旧变量赋值（err 等价于 `=`） | `result, err := db.Query(...)` |
| 所有变量都已存在（多变量） | `a, err = baz()` | 必须用 `=` | `videoID, err = getFromCache()` |
| 所有变量都已存在（单变量） | `err = qux()` | 只能用 `=` | `err = store.WithTx(...)` ⭐ |

---

## § 7. 实际项目中的最佳实践

### 7.1 错误处理链

```go
// ✅ 推荐写法：优雅地复用 err
func (s *InteractionServiceImpl) VideoLikeAction(ctx context.Context, req *v1.VideoLikeActionRequest) (*v1.VideoLikeActionResponse, error) {
    // 第一次声明 err
    videoID, err := util.ParseUint(req.VideoId)
    if err != nil {
        return nil, err
    }

    // 复用 err（至少有一个新变量 video）
    video, err := db.GetVideoByID(store, videoID)
    if err != nil {
        return nil, err
    }

    // 只有 err 一个变量，必须用 =
    err = store.WithTx(func(txStore *dal.Store) error {
        return performLikeAction(txStore, videoID)
    })
    if err != nil {
        return nil, err
    }

    return &v1.VideoLikeActionResponse{
        Base: response.Success(),
    }, nil
}
```

### 7.2 需要保留中间错误时

```go
// 场景：需要区分不同的错误来源
parseErr := validateInput(req)
if parseErr != nil {
    log.Printf("参数校验失败: %v", parseErr)
    return response.ParamError()
}

dbErr := db.SaveData(data)
if dbErr != nil {
    log.Printf("数据库操作失败: %v", dbErr)
    return response.InternalError()
}
```

### 7.3 避免遮蔽（Shadowing）

```go
// ❌ 错误示例：err 被遮蔽
func foo() error {
    data, err := fetchData()
    if err != nil {
        return err
    }

    if needValidation {
        // 这里创建了新的局部 err，遮蔽了外层的 err
        result, err := validate(data)
        if err != nil {
            return err
        }
        processResult(result)
    }

    // 这里的 err 仍然是 fetchData 的错误，不是 validate 的
    return err  // 可能返回旧错误！
}

// ✅ 正确写法：明确作用域
func foo() error {
    data, err := fetchData()
    if err != nil {
        return err
    }

    if needValidation {
        // 复用外层的 err
        var result ValidationResult
        result, err = validate(data)  // 使用 = 而不是 :=
        if err != nil {
            return err
        }
        processResult(result)
    }

    return nil
}
```

---

## § 8. 常见错误与调试

### 8.1 编译错误：`no new variables on left side of :=`

```go
// 错误代码
videoID, err := util.ParseUint(req.VideoId)
videoID, err := util.ParseUint(req.UserId)  // ❌ 编译错误

// 修复方式
videoID, err := util.ParseUint(req.VideoId)
videoID, err = util.ParseUint(req.UserId)   // ✅ 改用 =
```

### 8.2 变量遮蔽（Shadowing）

```go
// 错误代码
func process() error {
    data, err := loadData()
    if err != nil {
        return err
    }

    if condition {
        // 这里的 err 是新变量，遮蔽了外层的 err
        result, err := transform(data)
        log.Printf("transform error: %v", err)  // 这个 err 是局部的
    }

    // 这里的 err 还是 loadData 的错误！
    return err  // ⚠️ 可能返回错误的 err
}

// 使用 go vet 检测
$ go vet ./...
# 会警告：declaration of "err" shadows declaration at line X
```

**修复方式**：
```go
func process() error {
    data, err := loadData()
    if err != nil {
        return err
    }

    if condition {
        var result TransformResult
        result, err = transform(data)  // 复用外层 err
        if err != nil {
            return err
        }
        log.Printf("result: %v", result)
    }

    return nil
}
```

### 8.3 IDE 检查工具

- **GoLand / VSCode Go**：会高亮遮蔽的变量
- **go vet**：检测遮蔽问题
  ```bash
  go vet ./...
  ```
- **golangci-lint**：更严格的检查
  ```bash
  golangci-lint run --enable=shadow
  ```

---

## § 9. 进阶：作用域可视化

### 代码示例

```go
func example() {
    // 作用域 A
    a, err := foo()  // err₁ 在作用域 A

    if condition {
        // 作用域 B（嵌套在 A 内）
        b, err := bar()  // err₂ 在作用域 B，遮蔽 err₁
        fmt.Println(err)  // 这里访问的是 err₂
    }

    fmt.Println(err)  // 这里访问的是 err₁（err₂ 已超出作用域）
}
```

### 作用域图解

```
┌─────────────────────────────────────┐
│ 函数作用域                           │
│                                     │
│  a, err₁ := foo()                  │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ if 块作用域                    │ │
│  │                               │ │
│  │  b, err₂ := bar()  // 遮蔽 err₁│ │
│  │  fmt.Println(err)  // 输出 err₂│ │
│  │                               │ │
│  └───────────────────────────────┘ │
│                                     │
│  fmt.Println(err)  // 输出 err₁     │
│                                     │
└─────────────────────────────────────┘
```

---

## § 10. 关键代码位置

| 文件 | 行号 | 说明 |
|------|------|------|
| [interaction_service.go](../video-platform/biz/handler/v1/interaction_service.go#L85) | 85 | `videoID, err := util.ParseUint(...)` - err 首次声明 |
| [interaction_service.go](../video-platform/biz/handler/v1/interaction_service.go#L88) | 88 | `err = store.WithTx(...)` - 单变量赋值必须用 = |
| [user_service.go](../video-platform/biz/handler/v1/user_service.go) | 多处 | 错误处理链中 err 的复用示例 |

---

## § 11. 推荐阅读

- [Go 官方文档 - Short variable declarations](https://go.dev/ref/spec#Short_variable_declarations)
- [Effective Go - Redeclaration and reassignment](https://go.dev/doc/effective_go#redeclaration)
- [Go by Example - Variables](https://gobyexample.com/variables)
- [Common Go Mistakes - Variable Shadowing](https://100go.co/#shadowing-variables-12)

---

## § 12. 总结

1. **`:=` 与 `=` 的选择规则**：
   - 所有变量都是新的 → 用 `:=`
   - 至少有一个新变量 → 可以用 `:=`（已存在变量等价于 `=`）
   - 所有变量都已存在 → 必须用 `=`

2. **实际项目中的建议**：
   - 优先复用 `err` 变量（Go 惯用法）
   - 注意变量遮蔽问题（使用 `go vet` 检测）
   - 单变量赋值时明确使用 `=`

3. **关键记忆点**：
   ```go
   a, err := foo()   // err 第一次声明
   b, err := bar()   // err 等价于 = 赋值 ✅
   err = baz()       // 只有 err 时必须用 = ✅
   ```
