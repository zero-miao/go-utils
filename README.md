# golang common utils
开发 golang 程序的通用工具

## Email
发送邮件. 

通过统一的邮件服务器, 发送邮件(普通邮件|系统邮件), 对自带的 email 库进行了封装. 

## Logger
记录日志

封装 zap 包. 并实现了三个 handler: file, rotate_file, email

其中 email handler 基于 email 模块. 

rotate 实现了定时切割的功能. 

## grpc_error
自定义的 grpc 框架下的 error 结构: AppError 

用户 可以使用默认的 ErrorType 

也可以自定义 ErrorType, 然后通过 RegisterError 来注册. 

注意 ErrorType 中的 Code 不能重复, 会在注册时进行检查. 

## worker
异步定时任务. 具体参见 `worker/readme.md`