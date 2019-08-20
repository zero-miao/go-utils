package grpc_error

import "google.golang.org/grpc/codes"

// ==========================
// @Author : zero-miao
// @Date   : 2019-06-28 10:47
// @File   : define.go
// @Project: utils/grpc_error
// ==========================

const (
	// E0: 协议错误, 底层硬件错误: 网络错误, 内存错误, 硬盘错误,
	EGRPC    = "E001"
	EMemory  = "E002"
	EDisk    = "E003"
	ENetwork = "E004" // 一般反应为 协议错误.

	// E10: 三方依赖错误, 中间件错误: redis, nginx, mysql,
	EMiddleware = "E100"

	// E11: 三方库错误
	EPkg = "E110"

	// E4: 业务错误, 用户原因
	ERequest = "E400"

	// E5: 代码错误
	EInternal = "E500"
)

var ErrorTypeList []ErrorType

func init() {
	ErrorTypeList = []ErrorType{
		{Code: EGRPC, Name: "GRPC异常", GRPCCode: codes.Internal, Email: true},
		{Code: EMemory, Name: "内存操作异常", GRPCCode: codes.Internal, Email: true},
		{Code: EDisk, Name: "文件系统异常", GRPCCode: codes.Internal, Email: true},
		{Code: ENetwork, Name: "网络异常", GRPCCode: codes.Internal, Email: true},
		{Code: EMiddleware, Name: "中间件异常", GRPCCode: codes.Internal, Email: true},
		{Code: EPkg, Name: "库异常", GRPCCode: codes.Internal},
		{Code: ERequest, Name: "请求异常", GRPCCode: codes.InvalidArgument},
		{Code: EInternal, Name: "内部错误", GRPCCode: codes.Internal},

		//{Code: ERedisServer, Name: "Redis服务器错误", GRPCCode: codes.Internal},

		//{Code: EDetectUnknown, Name: "三方机检服务异常", GRPCCode: codes.Internal, Email: true},
		//{Code: EDetect, Name: "三方机检服务异常", GRPCCode: codes.Internal},
		//
		//{Code: EJsonInternal, Name: "JSON 解析异常", GRPCCode: codes.Internal, Email: true},
		//{Code: EJson, Name: "JSON 解析异常", GRPCCode: codes.InvalidArgument},
		//{Code: EBase64Internal, Name: "Base64 解析异常", GRPCCode: codes.Internal, Email: true},
		//{Code: EBase64, Name: "Base64 解析异常", GRPCCode: codes.InvalidArgument},

		//{Code: EForbidden, Name: "请求禁止", GRPCCode: codes.PermissionDenied},
		//{Code: EAuth, Name: "认证失败", GRPCCode: codes.Unauthenticated},
	}

	for _, item := range ErrorTypeList {
		RegisterError(item)
	}
}

func MemoryError(info string, err error, args interface{}) *AppError {
	return New(EMemory, info, err, args)
}

func DiskError(info string, err error, args interface{}) *AppError {
	return New(EDisk, info, err, args)
}

func NetworkError(info string, err error, args interface{}) *AppError {
	return New(ENetwork, info, err, args)
}

func GRPCError(info string, err error, args interface{}) *AppError {
	return New(EGRPC, info, err, args)
}

func MiddlerwareError(middleware string, err error, args interface{}) *AppError {
	return New(EMiddleware, middleware, err, args)
}

func PkgError(pkgMethod string, err error, args interface{}) *AppError {
	return New(EPkg, pkgMethod, err, args)
}

func RequestError(info string, err error, args interface{}) *AppError {
	return New(ERequest, info, err, args)
}

func BadRequest(msg string) *AppError {
	return RequestError(msg, nil, nil)
}

func SomethingInvalid(kvPair ...interface{}) *AppError {
	l := len(kvPair)
	if l%2 != 0 {
		panic("k-v pair not pairs")
	}
	name := "something"
	if l > 0 {
		name = kvPair[0].(string)
	}
	args := map[string]interface{}{}
	for i := 0; i+1 < l; i += 2 {
		args[kvPair[i].(string)] = kvPair[i+1]
	}
	return RequestError(name+" invalid", nil, args)
}

func SomethingRequired(key string) *AppError {
	return RequestError(key+" required", nil, nil)
}

func GRPCRecvError(err error) *AppError {
	return GRPCError("stream.recv", err, nil)
}

func GRPCSendError(err error) *AppError {
	return GRPCError("stream.send", err, nil)
}
