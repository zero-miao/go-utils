package worker

import "context"

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 10:22
// @File   : runtime.go
// @Project: utils/cron
// ==========================

type RuntimePair struct {
	Key   string
	Value interface{}
	// 当 key 对应的值为 CAS 时, 才进行更新. 不等于则不更新
	// CAS=nil 时, 不判断, 直接更新
	CAS interface{}
	// 不存在时才能设置
	SetNX bool
	// 存在时才能设置
	SetXX bool
}

type RuntimeInterface interface {
	GetCoreRuntime() map[string]interface{}
	SetRuntime(pairs ...RuntimePair)
	GetRuntime(keys ...string) map[string]interface{}
	// 对 key 增加 score. 要求 key 对应的值必须是 int 类型.
	// score可正可负, init 表示初始值. 如果 key 不存在, 则会使用初始值.
	// 返回计算后的结果.
	IncrBy(key string, score int, init int, min, max int) int
}

type PoolInterface interface {
	// 初始静态资源, 运行中只读
	Init()
	// 运行任务池
	Run(ctx context.Context)
	// 停止运行任务池(不在启动新的)
	Close()
	// 清理相关装填
	Reset()
	// 相关信息
	String() string
	// 详细信息
	Detail() string
}
