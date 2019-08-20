package logger

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestYamlInit(t *testing.T) {
	data, err := ioutil.ReadFile("sample.yaml")
	if err != nil {
		panic(err)
	}
	YamlInit(data)
	fmt.Println(Logging)
}
