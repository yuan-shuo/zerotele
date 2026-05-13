# 现有任务
1. 实现 zerotele check zerotele.yaml 指令，用于校验 yaml 文件是否符合要求

# 已完成
1. 合并 gologfields + gometrics 项目，构建 github.com/yuan-shuo/zerotele 项目
2. 要求：仅需要一份 yaml 文件，即可自动生成日志字段代码和类型安全指标代码

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