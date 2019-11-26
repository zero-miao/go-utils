package worker

import (
	"github.com/zero-miao/go-utils/logger"
	"sync"
	"testing"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-07-08 11:22
// @File   : example_test.go
// @Project: utils/cron
// ==========================

func TestTimeCondition(t *testing.T) {
	logger.YamlInit([]byte(`logging:
  worker:
    handler:
      - typ: file
        filename: "/dev/stdout"
        format: "console"
        level: "debug"
    caller: true
`))
	Register("normal_short1", &NormalShortWorker{}, []ConditionInterface{&MaxCountCondition{4}, &TimeCondition{time.Second * 5, time.Second * 2}})
	//Register("normal_short2", &NormalShortWorker{}, []ConditionInterface{&MaxCountCondition{10}, &ConcurrencyCondition{1}})
	//Register("normal_short3", &NormalShortWorker{}, []ConditionInterface{&MaxCountCondition{10}, &ConcurrencyCondition{1}})

	sig := make(chan ControlSig, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Manage(sig) // manage 会自己退出. (只要里面没有 pool 在运行)
	}()
	wg.Wait()
}

func TestNormalShortWorker_Run(t *testing.T) {
	logger.YamlInit([]byte(`logging:
  worker:
    handler:
      - typ: file
        filename: "/dev/stdout"
        format: "console"
        level: "debug"
    caller: true
`))
	Register("normal_short1", &NormalShortWorker{}, []ConditionInterface{&MaxCountCondition{10}, &ConcurrencyCondition{1}})
	Register("normal_short2", &NormalShortWorker{}, []ConditionInterface{&MaxCountCondition{10}, &ConcurrencyCondition{1}})
	Register("normal_short3", &NormalShortWorker{}, []ConditionInterface{&MaxCountCondition{10}, &ConcurrencyCondition{1}})

	sig := make(chan ControlSig, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Manage(sig) // manage 会自己退出. (只要里面没有 pool 在运行)
	}()
	wg.Wait()
}

func TestNormalLongWorker_Run(t *testing.T) {
	logger.YamlInit([]byte(`logging:
  worker:
    handler:
      - typ: file
        filename: "/dev/stdout"
        format: "console"
        level: "debug"
    caller: true
`))
	Register("normal_long1", &NormalLongWorker{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})
	Register("normal_long2", &NormalLongWorker{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})

	sig := make(chan ControlSig, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Manage(sig) // manage 会自己退出. (只要里面没有 pool 在运行)
	}()
	wg.Wait()
}

func TestNormalErrorWorker_Run(t *testing.T) {
	logger.YamlInit([]byte(`logging:
  worker:
    handler:
      - typ: file
        filename: "/dev/stdout"
        format: "console"
        level: "debug"
    caller: true
`))
	Register("normal_error", &NormalErrorWorker{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})
	Register("normal_error1", &NormalErrorWorker1{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})
	Register("normal_error2", &NormalErrorWorker2{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})

	sig := make(chan ControlSig, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Manage(sig) // manage 会自己退出. (只要里面没有 pool 在运行)
	}()
	wg.Wait()
}

func TestPanicErrorWorker_Run(t *testing.T) {
	logger.YamlInit([]byte(`logging:
  worker:
    handler:
      - typ: file
        filename: "/dev/stdout"
        format: "console"
        level: "debug"
    caller: true
`))
	Register("panic_error", &PanicErrorWorker{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})
	Register("panic_error", &PanicErrorWorker{}, []ConditionInterface{&MaxCountCondition{5}, &ConcurrencyCondition{2}})

	sig := make(chan ControlSig, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Manage(sig) // manage 会自己退出. (只要里面没有 pool 在运行)
	}()
	wg.Wait()
}
