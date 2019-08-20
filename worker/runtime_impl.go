package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zero-miao/go-utils/logger"
	"strings"
	"sync"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 09:30
// @File   : pool_impl.go
// @Project: utils/cron
// ==========================

func Register(name string, worker WorkerInterface, conditions []ConditionInterface) {
	if running {
		panic("can't register after running")
	}
	pool := RuntimePool{
		Name:      name,
		worker:    worker,
		condition: conditions,
	}
	pool.Init()
	poolHub[name] = &pool
}

type poolStatus int

const (
	PoolInit = iota
	PoolReset
	PoolRunning
	PoolClosing
)

type RuntimePool struct {
	Name   string
	worker WorkerInterface

	// 静态区域
	Static map[string]interface{}

	// 按顺序依次判断条件.
	condition []ConditionInterface
	// runtime 运行时环境
	runtimeLock  sync.Mutex
	runtimeParam map[string]interface{}

	// 必定存在的运行环境.
	coreLock    sync.RWMutex
	coreRuntime map[string]interface{}
	// 任务池状态
	Status poolStatus
}

func (p *RuntimePool) SetRuntime(pairs ...RuntimePair) {
	p.runtimeLock.Lock()
	defer p.runtimeLock.Unlock()
	for _, item := range pairs {
		origin, ok := p.runtimeParam[item.Key]
		if ok {
			if item.SetNX {
				continue
			}
			if item.CAS == nil || item.CAS == origin {
				p.runtimeParam[item.Key] = item.Value
			}
		} else {
			if item.SetXX {
				continue
			}
			p.runtimeParam[item.Key] = item.Value
		}
	}
}

func (p *RuntimePool) GetRuntime(keys ...string) map[string]interface{} {
	p.runtimeLock.Lock()
	defer p.runtimeLock.Unlock()
	res := map[string]interface{}{}
	for _, item := range keys {
		temp, ok := p.runtimeParam[item]
		if ok {
			res[item] = temp
		}
	}
	return res
}

// core_check:
// core_check_count:
// core_run:
// core_run_count:
// core_finish:
// core_finish_count:
func (p *RuntimePool) GetCoreRuntime() map[string]interface{} {
	p.coreLock.RLock()
	defer p.coreLock.RUnlock()
	return p.coreRuntime
}

func (p *RuntimePool) IncrBy(key string, score int, init int, min, max int) int {
	p.runtimeLock.Lock()
	defer p.runtimeLock.Unlock()
	origin, ok := p.runtimeParam[key]
	var originInt int
	if !ok {
		originInt = init
	} else {
		temp, ok := origin.(int)
		if ok {
			originInt = temp
		} else {
			originInt = init
		}
	}
	result := originInt + score
	if result > max {
		result = max
	}
	if result < min {
		result = min
	}
	p.runtimeParam[key] = result
	return result
}

func (p *RuntimePool) Reset() {
	p.runtimeLock.Lock()
	p.runtimeParam = map[string]interface{}{}
	p.Status = PoolReset
	p.runtimeLock.Unlock()

	p.coreLock.Lock()
	p.coreRuntime = map[string]interface{}{}
	p.coreLock.Unlock()
}

func (p *RuntimePool) CoreCheck() {
	p.coreLock.Lock()
	defer p.coreLock.Unlock()
	p.coreRuntime["core_check"] = time.Now()
	count := 0
	if temp, ok := p.coreRuntime["core_check_count"]; ok && temp != nil {
		count = temp.(int)
	}
	p.coreRuntime["core_check_count"] = count + 1
}

func (p *RuntimePool) CoreToRun() {
	p.coreLock.Lock()
	defer p.coreLock.Unlock()
	p.coreRuntime["core_run"] = time.Now()
	count := 0
	if temp, ok := p.coreRuntime["core_run_count"]; ok && temp != nil {
		count = temp.(int)
	}
	p.coreRuntime["core_run_count"] = count + 1
}

func (p *RuntimePool) CoreResult() {
	p.coreLock.Lock()
	defer p.coreLock.Unlock()
	p.coreRuntime["core_finish"] = time.Now()
	count := 0
	if temp, ok := p.coreRuntime["core_finish_count"]; ok && temp != nil {
		count = temp.(int)
	}
	p.coreRuntime["core_finish_count"] = count + 1
}

// RuntimePool.Run 死循环, 负责启动任务.
// ctx: 结束时, 将不再启动任务. 已有任务如何应对取决于任务本身.
// 条件判断本身是同步顺序执行. 会阻塞函数进行. Condition 的判断逻辑应该尽可能简单
func (p *RuntimePool) Run(ctx context.Context) {
	p.Status = PoolRunning
	rest := true
	for rest {
		select {
		case <-ctx.Done():
			// 不在运行新的 worker
			rest = false
			logger.SugarL("worker").Debugw("pool never run", "pool", p.String(), "reason", "context done")
		case <-time.Tick(time.Second):
			// 执行条件检查
			run := true
			//logger.SugarL("worker").Debugw("runtime for "+p.Name, "core", p.coreRuntime)
			p.CoreCheck()
			for _, cond := range p.condition {
				status := cond.Check(p)
				if status == CheckAbort {
					run = false
					logger.SugarL("worker").Debugw("pool condition check abort", "pool", p.String(), "cond", cond)
					break
				} else if status == CheckQuit {
					run = false
					rest = false
					logger.SugarL("worker").Debugw("pool condition check quit", "pool", p.String(), "cond", cond)
				}
			}
			if run {
				p.CoreToRun()
				for _, cond := range p.condition {
					cond.ToRun(p)
				}
				p.coreLock.RLock()
				counter := p.coreRuntime["core_run_count"].(int)
				p.coreLock.RUnlock()
				go func(counter int) {
					// counter 为当前协程 id.
					defer func() {
						if err := recover(); err != nil {
							logger.SugarL("worker").Errorw("pool worker panic", "pool", p.Detail())
							rest = false
						}
					}()
					name := fmt.Sprintf("%s_%d", p.Name, counter)
					result := p.worker.Run(ctx, name)
					p.CoreResult()
					for _, cond := range p.condition {
						cond.MarkResult(p, name, result)
					}
				}(counter)
			} else {
				for _, cond := range p.condition {
					cond.ToSkip(p)
				}
			}
		}
	}
}

func (p *RuntimePool) Close() {
	p.Status = PoolClosing
}

func (p *RuntimePool) Init() {
	p.Static = map[string]interface{}{}
	p.Status = PoolInit
	p.runtimeParam = map[string]interface{}{}
	p.coreRuntime = map[string]interface{}{}
	for _, cond := range p.condition {
		cond.Init(p.Static)
	}
	//logger.SugarL("worker").Debugw("static for "+p.Name, "static", p.Static)
}

func (p *RuntimePool) StringStatus() string {
	switch p.Status {
	case PoolInit:
		return "init"
	case PoolReset:
		return "reset"
	case PoolRunning:
		return "running"
	case PoolClosing:
		return "closing"
	default:
		return "unknown"
	}
}

func (p *RuntimePool) String() string {
	return fmt.Sprintf("%s:%s", p.Name, p.StringStatus())
}

func (p *RuntimePool) Detail() string {
	template := `Pool: %s Status: %s

condition: 
	%v

Core: 
	%s

runtime: 
	%s
`
	return fmt.Sprintf(template, p.Name, p.StringStatus(), printCondition(p.condition), printMap(p.coreRuntime), printMap(p.runtimeParam))
}

func printCondition(data []ConditionInterface) []string {
	res := make([]string, len(data))
	for index, item := range data {
		b, _ := json.Marshal(item)
		res[index] = fmt.Sprintf("%v", string(b))
	}
	return res
}

func printMap(data map[string]interface{}) string {
	res := make([]string, 0, len(data))
	for key, value := range data {
		res = append(res, fmt.Sprintf("%s:%v", key, value))
	}
	return strings.Join(res, "\n\t")
}
