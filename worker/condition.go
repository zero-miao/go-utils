package worker

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 09:43
// @File   : condition.go
// @Project: utils/cron
// ==========================

type CheckStatus int

const (
	// 当前条件判定通过
	CheckContinue = iota
	// 当前条件判定不通过
	CheckAbort
	// 判定永久退出
	CheckQuit
)

type ConditionInterface interface {
	// 初始化. 传入共享区域, 可以向其中写入共享数据.
	Init(map[string]interface{})
	// 每一个控制器, 依据当前状态, 给出是否执行的判断
	Check(RuntimeInterface) CheckStatus
	// 当决定要执行前, 调用该函数
	ToRun(RuntimeInterface)
	// 当决定本轮不执行时, 调用该函数
	ToSkip(RuntimeInterface)
	// 当执行结束时, 调用该函数
	MarkResult(RuntimeInterface, string, WorkerRunStatus)
}
