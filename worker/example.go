package worker

import (
	"context"
	"fmt"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 11:07
// @File   : example.go
// @Project: utils/cron
// ==========================

//  需要实现 WorkerInterface

type NormalShortWorker struct {
}

func (w *NormalShortWorker) Run(ctx context.Context, id string) WorkerRunStatus {
	fmt.Println(id, "run===")
	return RunOK
}

type NormalLongWorker struct {
}

func (w *NormalLongWorker) Run(ctx context.Context, id string) WorkerRunStatus {
	time.Sleep(time.Second * 10)
	fmt.Println(id, "run===")
	return RunOK
}

type NormalErrorWorker struct {
}

func (w *NormalErrorWorker) Run(ctx context.Context, id string) WorkerRunStatus {
	fmt.Println(id, "run===")
	return RunFailed
}

type NormalErrorWorker1 struct {
}

func (w *NormalErrorWorker1) Run(ctx context.Context, id string) WorkerRunStatus {
	fmt.Println(id, "run===")
	return RunExit
}

type NormalErrorWorker2 struct {
}

func (w *NormalErrorWorker2) Run(ctx context.Context, id string) WorkerRunStatus {
	fmt.Println(id, "run===")
	return RunRetry
}

type PanicErrorWorker struct {
}

func (w *PanicErrorWorker) Run(ctx context.Context, id string) WorkerRunStatus {
	panic(id + " panic")
	return RunOK
}
