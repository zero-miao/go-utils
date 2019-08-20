package grpc_error

import (
	"fmt"
	"github.com/zero-miao/go-utils/email"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime/debug"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-05-29 17:26
// @File   : errors.go
// @Project: gos/utils
// ==========================
const EUnknown = "E000"

func init() {
	codeTypeMap = map[string]ErrorType{}
	RegisterError(ErrorType{Code: EUnknown, Name: "未知错误", GRPCCode: codes.Unknown, Email: true})
}

func New(code, detail string, err error, info interface{}) *AppError {
	c, ok := codeTypeMap[code]
	if !ok {
		c = codeTypeMap[EUnknown]
	}
	e := &AppError{Code: c, Desc: detail, Err: err, Info: info}
	if c.Email {
		go email.SendServerMail("app error email", fmt.Sprintf("%s\n\ntraceback: \n%s", e.Detail(), string(debug.Stack())), "text/plain")
	}
	return e
}

type ErrorType struct {
	Code     string
	Name     string
	GRPCCode codes.Code
	Email    bool
}

func (t *ErrorType) String() string {
	return fmt.Sprintf("%s-%s", t.Code, t.Name)
}

var codeTypeMap map[string]ErrorType

func RegisterError(item ErrorType) {
	if errType, ok := codeTypeMap[item.Code]; ok {
		panic("自定义异常编码 " + item.Code + " 重复: " + errType.String())
	}
	codeTypeMap[item.Code] = item
}

type AppError struct {
	Code ErrorType // 错误名称代码

	Desc string      // 错误描述, 用户填写
	Err  error       // 可能由其他 error 引起.
	Info interface{} // 详细信息, 可以依据错误代码来识别具体格式.
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("[%s]%s %v", e.Code.String(), e.Desc, e.Err)
	}
	return fmt.Sprintf("[%s]%s %v", e.Code.String(), e.Desc, e.Info)
}

func (e *AppError) Detail() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("code: %s \n\t desc: %s \n\t err: %s \n\t info: %v", e.Code.String(), e.Desc, e.Err, e.Info)
}

func (e *AppError) GRPCStatus() *status.Status {
	if e == nil {
		return nil
	}
	return status.Newf(e.Code.GRPCCode, e.Desc)
}
