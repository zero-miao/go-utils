package grpc_error

import "google.golang.org/grpc/codes"

// ==========================
// @Author : zero-miao
// @Date   : 2019-06-28 11:11
// @File   : redis.go
// @Project: utils/grpc_error
// ==========================

const (
	ERedisServer = "E101"
	ERedisClient = "E102"
)

func RedisErrorInit() {
	RegisterError(ErrorType{Code: ERedisClient, Name: "Redis请求错误", GRPCCode: codes.Internal})
	RegisterError(ErrorType{Code: ERedisServer, Name: "Redis服务器错误", GRPCCode: codes.Internal, Email: true})
}

func RedisServerError(info string, err error, args ...interface{}) *AppError {
	return New(ERedisServer, info, err, args)
}

func RedisClientError(info string, err error, args ...interface{}) *AppError {
	return New(ERedisClient, info, err, args)
}
