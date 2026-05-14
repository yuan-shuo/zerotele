# 现有任务
1. 使用日志生成的代码存在一个问题：字段给入顺序是是手动指定的，举例如下：
```go
if _, err := pipe.Exec(ctx); err != nil {
    logc.Errorw(ctx, "验证码缓存失败（执行 Pipeline 失败）",
        logger.WEmail(email),
        logger.WErrorMsg(err.Error()),
    )
    return ""
}
```
可以看到，字段的顺序是手动指定的，也就是说使用者如果想要按顺序写入日志字段，需要自行记忆顺序或者控制，这实在是不方便
所以我的想法是：实现一个字段的sort函数，基于yaml的字段顺序，自动排序日志字段的写入顺序，比如yaml里的error_msg在顶部那他就是第一个日志字段
实际yaml展示的字段顺序就是日志字段的顺序，例如：
yaml -> {
    name: msg
    name: email
}
log -> {
    msg:"xxx", 
    email:"xxx"
}
但不能仅实现sort方法，因为二次包装会导致调用者侧代码臃肿例如logc.Errorw(ctx, "some content", logger.SortFields(xxx, xxx))
所以需要封装logc的各种方法（针对w*（字段类））：
logc.Debugw
logc.Errorw
logc.Infow
logc.Sloww
如上方法进行封装，使得调用者生成代码后，只需要按如下写法就可以了（logger代表生成代码所在的目录，即包名）：
```go
// 旧写法
if _, err := pipe.Exec(ctx); err != nil {
    logc.Errorw(ctx, "验证码缓存失败（执行 Pipeline 失败）",
        logger.WEmail(email),
        logger.WErrorMsg(err.Error()),
    )
    return ""
}
// 新写法
if _, err := pipe.Exec(ctx); err != nil {
    // 假设这里就是闲的就非得先写email，但是yaml里errormsg先出现，所以触发排序
    // 不关心先调用哪个字段的方法，顺序由yaml字段出现顺序唯一决定
    // L给出构建器，其余的字段方法填充
    // 日志函数末尾有s的才执行排序，无s的不管不会自动排序（注释需要提醒使用s会进行自动排序及其影响）
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Debugs()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Errors()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Infos()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Slows()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Debug()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Error()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Info()
    logger.L(ctx, "用户输入的content信息").WEmail(email).WErrorMsg(err.Error()).Slow()
    return ""
}
// 日志结果：
// {
//     "error_msg": "xxx" // 之所以排在email前面是因为yaml中先声明的errormsg字段
//     "email": "xxx",  
// }
```

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