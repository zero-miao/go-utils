package worker

import (
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 10:10
// @File   : condition_template.go
// @Project: utils/cron
// ==========================

// 时间点控制, 比如每天 1 点执行: TimeInterval=24h, Offset=1h
type TimeCondition struct {
	// 每整 TimeInterval 执行一次.
	TimeInterval time.Duration
	// TimeInterval 内的一个时间偏移.
	Offset time.Duration
}

func (c *TimeCondition) Init(static map[string]interface{}) {
	if c.TimeInterval < c.Offset {
		panic("TimeCondition.TimeInterval > TimeCondition.Offset")
	}
	static["base_time"] = c.TimeInterval.String()
	static["base_offset"] = c.Offset.String()
}

// 需要记录下, 上次运行时间
func (c *TimeCondition) Check(store RuntimeInterface) CheckStatus {
	rt := store.GetCoreRuntime()
	timeToCheck := rt["core_check"].(time.Time)
	env := store.GetRuntime("next_to_run")

	if nextRun, ok := env["next_to_run"]; ok && nextRun != nil {
		nextRunTime := nextRun.(time.Time)
		du := timeToCheck.Sub(nextRunTime)
		if du < 0 {
			// 时间未到 不能运行
			return CheckAbort
		} else {
			store.SetRuntime(RuntimePair{
				Key:   "next_to_run",
				Value: timeToCheck.Truncate(c.TimeInterval).Add(c.TimeInterval + c.Offset),
				SetXX: true,
			})
			if du > time.Second {
				// 过期了, 距离目标运行时间, 大于 1s, 不能执行
				return CheckAbort
			}
			return CheckContinue
		}
	} else {
		store.SetRuntime(RuntimePair{
			Key:   "next_to_run",
			Value: timeToCheck.Truncate(c.TimeInterval).Add(c.TimeInterval + c.Offset),
			SetNX: true,
		})
		return CheckAbort
	}
}

func (c *TimeCondition) ToRun(store RuntimeInterface) {}

func (c *TimeCondition) ToSkip(store RuntimeInterface) {}

func (c *TimeCondition) MarkResult(store RuntimeInterface, name string, result WorkerRunStatus) {}

// 频率控制
type FrequencyCondition struct {
	Interval time.Duration
}

func (c *FrequencyCondition) Init(static map[string]interface{}) {
	static["frequency"] = c.Interval.String()
}

// 需要记录下, 上次运行时间
func (c *FrequencyCondition) Check(store RuntimeInterface) CheckStatus {
	rt := store.GetCoreRuntime()
	timeToCheck := rt["core_check"].(time.Time)
	last, ok := rt["core_run"]
	if !ok {
		// 之前没有运行过
		return CheckContinue
	}
	if last.(time.Time).Add(c.Interval).After(timeToCheck) {
		return CheckAbort
	}
	return CheckContinue
}

func (c *FrequencyCondition) ToRun(store RuntimeInterface) {}

func (c *FrequencyCondition) ToSkip(store RuntimeInterface) {}

func (c *FrequencyCondition) MarkResult(store RuntimeInterface, name string, result WorkerRunStatus) {}

// 总执行次数控制
type MaxCountCondition struct {
	// 总运行次数
	MaxCount int
}

func (c *MaxCountCondition) Init(static map[string]interface{}) {
	static["max_count"] = c.MaxCount
}

// 需要记录下, 上次运行时间
func (c *MaxCountCondition) Check(store RuntimeInterface) CheckStatus {
	rt := store.GetCoreRuntime()
	runCount, ok := rt["core_run_count"]
	if !ok {
		return CheckContinue
	}
	if runCount.(int) >= c.MaxCount {
		return CheckQuit
	}
	return CheckContinue
}

func (c *MaxCountCondition) ToRun(store RuntimeInterface) {}

func (c *MaxCountCondition) ToSkip(store RuntimeInterface) {}

func (c *MaxCountCondition) MarkResult(store RuntimeInterface, name string, result WorkerRunStatus) {}

// 最大并发数控制
type ConcurrencyCondition struct {
	MaxConcurrency int
}

func (c *ConcurrencyCondition) Init(static map[string]interface{}) {
	static["max_concurrency"] = c.MaxConcurrency
}

// 需要记录下, 上次运行时间
func (c *ConcurrencyCondition) Check(store RuntimeInterface) CheckStatus {
	rt := store.GetCoreRuntime()
	runCount, ok := rt["core_run_count"]
	if !ok {
		return CheckContinue
	}
	run := runCount.(int)
	finishCount, ok := rt["core_finish_count"]
	finish := 0
	if ok {
		finish = finishCount.(int)
	}
	if run-finish >= c.MaxConcurrency {
		return CheckAbort
	}
	return CheckContinue
}

func (c *ConcurrencyCondition) ToRun(store RuntimeInterface) {
	//store.IncrBy("concurrency_count", 1, 0, 0, c.MaxConcurrency+1)
}

func (c *ConcurrencyCondition) ToSkip(store RuntimeInterface) {}

func (c *ConcurrencyCondition) MarkResult(store RuntimeInterface, name string, result WorkerRunStatus) {
	//store.IncrBy("concurrency_count", -1, 0, 0, c.MaxConcurrency+1)
}
