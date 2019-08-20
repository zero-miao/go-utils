package logger

import (
	"encoding/json"
	"github.com/zero-miao/go-utils/email"
	"strconv"
	"testing"
	"time"
)

func TestSugarL(t *testing.T) {
	var shortLog = "test"
	var objectLog = map[string]map[string][]interface{}{}
	for i := 0; i < 100; i++ {
		temp := "test" + strconv.Itoa(i)
		objectLog[temp] = map[string][]interface{}{
			temp: {temp, temp},
		}
	}
	//TestEnv()
	email.YamlInit([]byte(`mail_server: "smtp.exmail.qq.com"
mail_port: 465
mail_user: "zabbix@gaeamobile.com"
mail_password: "xxxxxxxx"
mail_subject_prefix: "[app-local]"
mail_default_from: "zabbix"
mail_admins:
  - name: "ao.mei"
    email: "ao.mei@gaea.com"
append_ip_subject: false`))
	YamlInit([]byte(`logging:
  test:
    handler:
      - typ: file
        filename: "/dev/stdout"
        format: "console"
        level: "debug"
      - typ: email
        level: "error"
        format: "json"
    caller: true
    str_field:
      - key: test
        value: the_test
      - key: test_ip
        dynamic_value: ipv4`))
	logger := SugarL("test")
	//logger.Debug(shortLog)
	logger.Error(shortLog)
	time.Sleep(time.Second * 2)
	//logger.Debug(objectLog)
	//logger.Error(objectLog)
}

// 100000             13720 ns/op             384 B/op          7 allocs/op
func BenchmarkSugarLShort(b *testing.B) {
	TestEnv()
	logger := SugarL("test")
	for i := 0; i < b.N; i++ {
		logger.Errorw("test", "i", i, "b.n", b.N)
	}
}

// 5000            234139 ns/op           40044 B/op        717 allocs/op
func BenchmarkSugarLObject(b *testing.B) {
	TestEnv()
	var objectLog = map[string]map[string][]interface{}{}
	for i := 0; i < 100; i++ {
		temp := "test" + strconv.Itoa(i)
		objectLog[temp] = map[string][]interface{}{
			temp: {temp, temp},
		}
	}
	logger := SugarL("test")
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(objectLog)
		logger.Errorw("test", "obj", string(data), "i", i, "b.n", b.N)
	}
}

func TestDeque_Add(t *testing.T) {
	var d Deque
	cap_ := 10
	d.Reset(cap_)
	for i := 0; i < 21; i++ {
		temp := d.Add(strconv.Itoa(i))
		if i < cap_ {
			if temp != "" {
				t.Errorf("error return1: %v, want %v", temp, "")
			}
		} else {
			if temp != strconv.Itoa(i-cap_) {
				t.Errorf("error return2: %v, want %v", temp, i-cap_)
			}
		}
	}
}
