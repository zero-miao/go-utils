package worker

import (
	"reflect"
	"testing"
)

func TestRuntimePool_SetRuntime(t *testing.T) {
	p := RuntimePool{runtimeParam: map[string]interface{}{}}
	pair := RuntimePair{
		Key:   "test1",
		Value: "test1",
	}
	p.SetRuntime(pair)
	data, ok := p.GetRuntime("test1")["test1"]
	if !ok || data != "test1" {
		t.Error("test1")
	}
	// test1=test1,
	pair = RuntimePair{
		Key:   "test2",
		Value: "test2",
		SetNX: true, // 不存在时才能设置
	}
	p.SetRuntime(pair)
	m := p.GetRuntime("test1", "test2")
	if m["test1"] != "test1" || m["test2"] != "test2" {
		t.Error("test2")
	}

	// test1=test1, test2=test2
	pair = RuntimePair{
		Key:   "test3",
		Value: "test3",
		SetXX: true, // 存在时才能设置
	}
	p.SetRuntime(pair)
	data, ok = p.GetRuntime("test3")["test3"]
	if ok {
		t.Error("test3")
	}

	// test1=test1, test2=test2
	pair = RuntimePair{
		Key:   "test2",
		Value: "test3",
		SetXX: true, // 存在时才能设置
	}
	p.SetRuntime(pair)
	data, ok = p.GetRuntime("test2")["test2"]
	if !ok || data != "test3" {
		t.Error("test2=test3")
	}

	// test1=test1, test2=test3
	pair = RuntimePair{
		Key:   "test2",
		Value: "test4",
		SetXX: true,    // 存在时才能设置
		CAS:   "test2", // 存在, 且原值等于 test2 才能设置
	}
	p.SetRuntime(pair)
	data, ok = p.GetRuntime("test2")["test2"]
	if !ok || data != "test3" {
		t.Error("test2!=test3")
	}

	// test1=test1, test2=test3
	pair = RuntimePair{
		Key:   "test2",
		Value: "test4",
		SetXX: true,    // 存在时才能设置
		CAS:   "test3", // 存在, 且原值等于 test3 才能设置
	}
	p.SetRuntime(pair)
	data, ok = p.GetRuntime("test2")["test2"]
	if !ok || data != "test4" {
		t.Error("test2!=test4")
	}
}

func assert(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}

func TestRuntimePool_IncrBy(t *testing.T) {
	p := RuntimePool{runtimeParam: map[string]interface{}{}}
	assert(t, p.IncrBy("test1", 1, 0, 0, 100), 1)
	assert(t, p.IncrBy("test1", 1, 0, 0, 100), 2)
	assert(t, p.IncrBy("test1", 10, 0, 0, 3), 3)
	assert(t, p.IncrBy("test1", -10, 0, 0, 3), 0)

	p.SetRuntime(RuntimePair{Key: "test2", Value: "test2", SetNX: true})
	assert(t, p.IncrBy("test2", -10, -1, 0, 3), 0)
	assert(t, p.IncrBy("test2", 10, -1, 0, 3), 3)
	assert(t, p.IncrBy("test2", 10, -1, 0, 30), 13)
}
