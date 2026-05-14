# zerotele

[![CI](https://github.com/yuan-shuo/zerotele/workflows/ci/badge.svg)](https://github.com/yuan-shuo/zerotele/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuan-shuo/zerotele)](https://goreportcard.com/report/github.com/yuan-shuo/zerotele)
[![codecov](https://codecov.io/gh/yuan-shuo/zerotele/branch/main/graph/badge.svg)](https://codecov.io/gh/yuan-shuo/zerotele)
[![Release](https://img.shields.io/github/release/yuan-shuo/zerotele.svg)](https://github.com/yuan-shuo/zerotele/releases/latest)
[![Go Version](https://img.shields.io/badge/go%20version-%3E=1.25-61CFDD.svg)](https://golang.org/)

用于生成 go-zero 观测代码的工具：用于补足 go-zero 未封装或缺少类型安全的部分观测功能：日志、指标

## 功能特性

- **统一配置**：通过一份 YAML 配置文件同时管理日志字段和指标定义
- **类型安全**：为日志字段和指标标签生成专门的类型，编译期检查避免错误
- **IDE 友好**：通过 `W` 前缀的构造器函数，IDE 自动补全一目了然
- **链式调用**：支持流畅的链式 API，代码更简洁优雅
- **自动排序**：带 `s` 后缀的日志方法按 YAML 字段定义顺序自动排序，无需手动控制字段顺序
- **自动脱敏**：支持敏感数据自动脱敏，符合 `go-zero` 的 `Sensitive` 接口
- **基数控制**：指标标签支持枚举值限制，防止标签基数爆炸
- **自动补全 mask 函数**：自动生成未实现的脱敏函数存根，避免手动查看类型

## 安装

使用 `go install` 安装，或者通过 [release](https://github.com/yuan-shuo/zerotele/releases/latest) 获取二进制文件自行存放到环境变量目录下

```bash
go install github.com/yuan-shuo/zerotele@latest
```

或者从源码构建：

```bash
git clone https://github.com/yuan-shuo/zerotele.git
cd zerotele
go build -o zerotele .
```

## 指令集

### zerotele check
用于校验 YAML 配置文件是否符合规范，一次性输出所有校验错误及错误总数
```bash
# 校验 zerotele.yaml 文件
zerotele check zerotele.yaml
```

### zerotele lf
用于省略 gozero 日志统一字段及重复的脱敏创建操作
```bash
# 在 ./internal/logger 目录下生成日志字段代码
zerotele lf zerotele.yaml -d ./internal/logger
# 在 ./logger 目录下生成日志字段代码 (更全面)
# 并自动补全 ./internal/logger/mask.go 的脱敏函数存根
# 若不存在 mask.go 会自行在 -d 指定的目录创建
zerotele lf zerotele.yaml -d ./internal/logger -m mask.go
```


### zerotele met
用于封装 gozero 的指标功能，使其满足类型安全和基数控制
```bash
# 在 ./internal/metrics 目录下生成类型安全指标器代码
zerotele met zerotele.yaml -d ./internal/metrics
```

## 使用方法

### 创建 YAML 配置文件

参考 `zerotele.example.yaml` 创建你的配置文件：

```yaml
# 观测服务名
service: user

# 日志字段清单
logfields:
  - name: user_id
    type: int64
    mask: true
    comment: 用户ID

  - name: user_name
    type: string
    comment: 用户名

# 指标清单
metrics:
  # Counter 类型示例
  - name: registrations_total
    help: Total number of user registrations
    type: counter
    labels:
      - name: source
        vals: [app, web]
      - name: error_type
        vals: ["*"]
    methods: [inc, add]

  # Gauge 类型示例
  - name: active_connections
    help: Current number of active connections
    type: gauge
    labels:
      - name: pool
        vals: [default, cache, db]
    methods: [set, inc, dec]

  # Histogram 类型示例
  - name: request_duration_ms
    help: HTTP request duration in milliseconds
    type: histogram
    labels:
      - name: method
        vals: [GET, POST, PUT, DELETE]
    buckets: [5, 10, 25, 50, 100, 250, 500, 1000]
    methods: [observe]
```

### 在项目中使用生成的代码

通过指令集中的代码获取生成的go代码文件

#### 使用日志字段

**传统方式（仍然支持）**

```go
package main

import (
    "context"
    "yourproject/internal/logger"
    "github.com/zeromicro/go-zero/core/logx"
)

func main() {
    ctx := context.Background()
    log := logx.WithContext(ctx)

    // 结构化日志，自动脱敏
    log.Infow("用户登录",
        logger.WUserId(12345678),           // 输出: "user_id": "****5678" (脱敏后)
        logger.WUserName("张三"),            // 输出: "user_name": "张三"
    )
}
```

**链式调用方式（推荐）**

```go
package main

import (
    "context"
    "yourproject/internal/logger"
)

func main() {
    ctx := context.Background()

    // 方式1：自动排序（带 s 后缀的方法）
    // 字段按 YAML 中定义的顺序输出，无需关心代码中的调用顺序
    logger.L(ctx, "用户登录").
        WUserName("张三").      // 在 YAML 中后定义
        WUserId(12345678).      // 在 YAML 中先定义
        Infos()                 // 输出顺序: user_id -> user_name

    // 方式2：保持代码顺序（不带 s 后缀的方法）
    // 字段按代码中调用的顺序输出
    logger.L(ctx, "用户登录").
        WUserName("张三").      // 先输出
        WUserId(12345678).      // 后输出
        Info()                  // 输出顺序: user_name -> user_id
}
```

**链式调用方法说明**

| 方法 | 是否排序 | 说明 |
|------|----------|------|
| `L(ctx, content)` | - | 创建日志构建器 |
| `WFieldName(value)` | - | 添加字段（以 `W` 开头，IDE 自动提示） |
| `Debugs()` | 排序 | Debug 级别，按 YAML 字段顺序输出 |
| `Infos()` | 排序 | Info 级别，按 YAML 字段顺序输出 |
| `Errors()` | 排序 | Error 级别，按 YAML 字段顺序输出 |
| `Slows()` | 排序 | Slow 级别，按 YAML 字段顺序输出 |
| `Debug()` | 不排序 | Debug 级别，按代码调用顺序输出 |
| `Info()` | 不排序 | Info 级别，按代码调用顺序输出 |
| `Error()` | 不排序 | Error 级别，按代码调用顺序输出 |
| `Slow()` | 不排序 | Slow 级别，按代码调用顺序输出 |

#### 使用指标

```go
package main

import (
    "yourproject/internal/metrics"
)

func main() {
    // 创建指标实例
    m := metrics.NewMetrics()

    // 使用 Counter
    m.RegistrationsTotal.Inc(logger.SourceApp, "err_1001")
    m.RegistrationsTotal.Add(5, logger.SourceWeb, "")

    // 使用 Gauge
    m.ActiveConnections.Set(100, logger.PoolDefault)
    m.ActiveConnections.Inc(logger.PoolCache)
    m.ActiveConnections.Dec(logger.PoolDb)

    // 使用 Histogram
    m.RequestDurationMs.Observe(150, logger.MethodGET)
}
```

### 实现脱敏函数 ( 针对 mask: true )

如果使用 `-m` 参数运行工具，会自动生成 `mask.go` 文件：

```go
package logger

// maskUserId 对用户ID进行脱敏
// 请在此实现具体的脱敏逻辑
func maskUserId(userId int64) any {
    // TODO: 实现用户ID脱敏逻辑
    return userId
}
```

你只需要填充具体的脱敏逻辑即可。当你添加新的 mask 字段后，再次使用 `-m` 运行工具，会自动追加新的函数存根，不会覆盖已有的实现。

**示例：实现脱敏逻辑**

```go
package logger

import "strconv"

// maskUserId 对用户ID进行脱敏
// 脱敏规则：显示前2位和后2位，中间用 **** 替换
func maskUserId(userId int64) any {
    s := strconv.FormatInt(userId, 10)
    if len(s) <= 4 {
        return "****"
    }
    return s[:2] + "****" + s[len(s)-2:]
}
```

## 配置文件说明

### 根级字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `service` | string | 是 | 观测服务名，作为指标的 service 标签值 |
| `logfields` | array | 否 | 日志字段列表 |
| `metrics` | array | 否 | 指标列表 |

### 日志字段配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 字段名（snake_case），如 `user_id` |
| `type` | string | 是 | Go 类型，如 `string`, `int64`, `bool` 等 |
| `mask` | bool | 否 | 是否需要脱敏，默认 `false` |
| `comment` | string | 否 | 字段注释说明 |

### 指标配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 指标名称（snake_case） |
| `help` | string | 是 | 指标帮助信息 |
| `type` | string | 是 | 指标类型：`counter`, `gauge`, `histogram` |
| `labels` | array | 否 | 标签列表 |
| `methods` | array | 是 | 支持的方法列表 |
| `buckets` | array | 否 | 直方图桶边界（仅 histogram） |

### 标签配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 标签名（snake_case） |
| `vals` | array | 是 | 标签可选值列表，使用 `["*"]` 表示任意值 |

### 支持的方法

| 指标类型 | 方法 | 说明 |
|----------|------|------|
| Counter | `inc` | 递增计数器 |
| Counter | `add` | 增加指定值 |
| Gauge | `set` | 设置值 |
| Gauge | `inc` | 递增 |
| Gauge | `dec` | 递减 |
| Histogram | `observe` | 记录观察值 |

## 命名规范

### snake_case 命名规范（OpenTelemetry）

需要满足 snake_case 规范的内容：
- `service`
- `logfields[].name`
- `metrics[].name`
- `metrics[].labels[].name`

命名规则：
- 必须以小写字母开头
- 只能包含小写字母、数字和下划线
- 不能连续出现多个下划线
- 不能以下划线开头或结尾
- 有效示例：`user_id`, `http_request_duration`, `trace_id`
- 无效示例：`UserID`, `user-id`, `user__id`, `_user_id`, `user_id_`, `123_user`

### 命名转换规则

| YAML 配置 | Go 类型名 | 说明 |
|-----------|-----------|------|
| `name: user_id` | `UserId` | 字段类型名 |
| `name: user_id` | `WUserId()` | 构造器函数名 |
| `name: user_id` | `user_id` | JSON 字段名 |
| `label: source` | `Source` | 枚举类型名 |
| `label: source` | `SourceApp` | 枚举值名（PascalCase） |

