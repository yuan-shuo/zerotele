# 现有任务

# 已完成
1. 合并 gologfields + gometrics 项目，构建 github.com/yuan-shuo/zerotele 项目
2. 要求：仅需要一份 yaml 文件，即可自动生成日志字段代码和类型安全指标代码
3. 日志字段链式调用与自动排序：
   - 实现 `L` 函数创建日志构建器，支持链式调用
   - 字段方法以 `W` 开头（如 `WUserId`, `WEmail`），便于 IDE 提示
   - 带 `s` 后缀的方法（`Debugs`, `Errors`, `Infos`, `Slows`）按 YAML 字段定义顺序自动排序
   - 不带 `s` 后缀的方法（`Debug`, `Error`, `Info`, `Slow`）保持字段添加顺序
   - 使用示例：
   ```go
   // 自动排序（YAML 中先定义的字段先输出）
   logger.L(ctx, "操作失败").WEmail(email).WErrorMsg(err.Error()).Errors()
   // 不排序（保持代码中字段添加顺序）
   logger.L(ctx, "操作失败").WEmail(email).WErrorMsg(err.Error()).Error()
   ```

## 指令集

### zerotele lf
用于省略 gozero 日志统一字段及重复的脱敏创建操作
```go
// 在 ./internal/logger 目录下生成日志字段代码
zerotele lf zerotele.yaml -d ./internal/logger
// 在 ./logger 目录下生成日志字段代码 (更全面)
// 并自动补全 ./internal/logger/mask.go 的脱敏函数存根
// 若不存在 mask.go 会自行在 -d 指定的目录创建
zerotele lf zerotele.yaml -d ./internal/logger -m mask.go
```


### zerotele met
用于封装 gozero 的指标功能，使其满足类型安全和基数控制
```go
// 在 ./internal/metrics 目录下生成类型安全指标器代码
zerotele met zerotele.yaml -d ./internal/metrics
```