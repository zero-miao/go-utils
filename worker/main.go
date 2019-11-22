package worker

import (
	"context"
	"github.com/zero-miao/go-utils/logger"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 09:44
// @File   : main.go
// @Project: utils/cron
// ==========================

var (
	poolHub = map[string]PoolInterface{}
	running = false
)

type ControlSig int

const (
	ControlStop    ControlSig = iota
	ControlRestart ControlSig = iota
)

// Manage 异步任务池 总控方法
// sig 用于控制退出或者重启
func Manage(sig <-chan ControlSig) {
	running = true
	ctx, cancel := context.WithCancel(context.Background())
	runningPool := 0
	for _, pool := range poolHub {
		pool.Init()
		logger.SugarL("worker").Infow("after pool init", "pool", pool.String())
		runningPool++
		go func(pool PoolInterface) {
			defer func() {
				if err := recover(); err != nil {
					logger.SugarL("worker").Errorw("pool panic", "err", err, "pool", pool.Detail())
				}
				runningPool--
			}()
			pool.Run(ctx)
			logger.SugarL("worker").Infow("after pool run", "pool", pool.String())
		}(pool)
	}
	rerun := true
	for rerun {
		select {
		case control := <-sig:
			cancel() // 让协程(各个任务池)停止
			for i := 0; i < 6 && runningPool > 0; i++ {
				time.Sleep(time.Second * 5)
			}
			if runningPool > 0 {
				logger.SugarL("worker").Warnw("there still pool running after 30s", "count", runningPool)
			}
			switch control {
			case ControlStop:
				logger.SugarL("worker").Infow("try to stop pool-hub")
				for _, pool := range poolHub {
					// 停止后, 可以再次调用 Manage 来启动. (并且可以调用 Register 注册新方法)
					pool.Close()
					logger.SugarL("worker").Infow("after pool close", "pool", pool.String())
					pool.Reset()
					logger.SugarL("worker").Infow("after pool reset", "pool", pool.String())
				}
				running = false
				rerun = false
			case ControlRestart:
				// 重启不会调用 Init.
				logger.SugarL("worker").Debugw("try to restart pool-hub")
				ctx, cancel = context.WithCancel(context.Background())
				for _, pool := range poolHub {
					pool.Close()
					logger.SugarL("worker").Infow("after pool close", "pool", pool.String())
					pool.Reset()
					logger.SugarL("worker").Infow("after pool reset", "pool", pool.String())
					pool.Run(ctx)
					logger.SugarL("worker").Infow("after pool rerun", "pool", pool.String())
				}
			}
		default:
			if runningPool > 0 {
				time.Sleep(time.Second * 3)
			} else {
				rerun = false
				running = false
			}
		}
	}
	logger.SugarL("worker").Info("manager exit")
}
