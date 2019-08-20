package logger

import (
	"go.uber.org/zap"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-05-22 15:00
// @File   : logger.go
// @Project: aotu/logger
// ==========================

var loggerMap = map[string]*zap.Logger{}

// 考虑并发读写 map 的问题
func L(module string) *zap.Logger {
	logger, ok := loggerMap[module]
	if ok {
		return logger
	}
	config, ok := Logging[module]
	if ok {
		logger := config.Build()
		loggerMap[module] = logger
		return logger
	}
	panic("invalid module: " + module)
}

var loggerSugarMap = map[string]*zap.SugaredLogger{}

// 考虑并发读写 map 的问题
func SugarL(module string) *zap.SugaredLogger {
	logger, ok := loggerSugarMap[module]
	if ok {
		return logger
	}
	llogger := L(module)
	logger = llogger.Sugar()
	loggerSugarMap[module] = logger
	return logger
}
