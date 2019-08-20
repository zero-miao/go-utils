package worker

import (
	"context"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-07 15:48
// @File   : pool.go
// @Project: utils/cron
// ==========================

// worker 运行的状态
type WorkerRunStatus int

const (
	// 运行成功
	RunOK = iota
	// 运行失败
	RunFailed
	// 运行失败, 立即进行重试判定
	RunRetry
	// 运行失败, 退出.
	RunExit
	// 运行超时
	RunTimeout
)

type WorkerInterface interface {
	// 传入运行上下文, 以及本次运行的唯一 id. id=name + counter
	Run(ctx context.Context, id string) WorkerRunStatus
}
