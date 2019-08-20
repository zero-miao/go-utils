package email

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
	fmt.Println(GeneralConfig.String())
	err = SendServerMail("test yaml init", string(data), "text/plain")
	if err != nil {
		panic(err)
	}
}
