package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"net"
	"strings"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-05-22 18:36
// @File   : configs.go
// @Project: aotu/logger
// ==========================

func YamlInit(data []byte) {
	temp := Config{}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		panic(err)
	}
	temp.Transform()
	for key := range Logging {
		SugarL(key)
	}
}

var LogLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

var consoleEncoderConfig = zapcore.EncoderConfig{
	TimeKey:    "ts_",
	EncodeTime: zapcore.ISO8601TimeEncoder,

	LevelKey:    "level_",
	EncodeLevel: zapcore.LowercaseLevelEncoder,

	CallerKey:    "caller_",
	EncodeCaller: zapcore.ShortCallerEncoder,

	MessageKey:     "msg_",
	NameKey:        "logger",
	StacktraceKey:  "trace_",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeDuration: zapcore.NanosDurationEncoder,
}

var ConsoleFormatter = zapcore.NewConsoleEncoder(consoleEncoderConfig)
var JsonFormatter = zapcore.NewJSONEncoder(consoleEncoderConfig)

var Logging map[string]LoggingConfig

type StringField struct {
	Key      string `yaml:"key"`
	Value    string `yaml:"value"`
	ValueGen string `yaml:"dynamic_value"`
}

func (f *StringField) GetValue() string {
	if f.Value != "" {
		return f.Value
	}
	switch f.ValueGen {
	case "ipv4":
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return ""
		}
		res := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
				res = append(res, ipnet.String())
			}
		}
		f.Value = strings.Join(res, ";")
		return f.Value
	default:
		f.Value = f.ValueGen
		return f.Value
	}
}

type ConfigHandler struct {
	Handler []struct {
		Typ      string        `yaml:"typ"`
		Filename string        `yaml:"filename"`
		Level    string        `yaml:"level"`
		Format   string        `yaml:"format"`
		Duration string        `yaml:"duration"`
		Replica  int           `yaml:"replica"`
		StrField []StringField `yaml:"str_field"`
	} `yaml:"handler"`
	Caller   bool          `yaml:"caller"`
	StrField []StringField `yaml:"str_field"`
}

type Config struct {
	Logging map[string]ConfigHandler `yaml:"logging"`
}

func (c *Config) Transform() {
	tempConfig := map[string]LoggingConfig{}
	for key, handlers := range c.Logging {
		opts := make([]zap.Option, 0)
		if handlers.Caller {
			opts = append(opts, zap.AddCaller())
		}
		fields := make([]zap.Field, 0)
		if len(handlers.StrField) > 0 {
			for _, item := range handlers.StrField {
				value := item.GetValue()
				if value != "" {
					fields = append(fields, zap.Field{
						Key:    item.Key,
						Type:   zapcore.StringType,
						String: item.Value,
					})
				}
			}
		}
		opts = append(opts, zap.Fields(fields...))
		tempHandlers := make([]Handler, 0)
		for _, handler := range handlers.Handler {
			format := ConsoleFormatter
			if handler.Format == "json" {
				format = JsonFormatter
			}
			level, ok := LogLevelMap[handler.Level]
			if !ok {
				panic("level invalid: " + handler.Level)
			}
			var temp Handler
			switch handler.Typ {
			case "file":
				temp = &FileHandler{
					Filename: handler.Filename,
					Level:    level,
					Format:   format,
				}
			case "rotate_file":
				du, err := time.ParseDuration(handler.Duration)
				if err != nil {
					panic("invalid duration:" + handler.Duration)
				}
				var layout string
				if du >= time.Hour*24 {
					layout = "2006-01-02"
				} else if du >= time.Hour {
					layout = "2006-01-02_15"
				} else if du >= time.Minute {
					layout = "2006-01-02_15_04"
				} else {
					layout = "2006-01-02_15_04_05"
				}
				temp = &RotateHandler{
					Filename: handler.Filename,
					Layout:   layout,
					Duration: du,
					Replica:  handler.Replica,
					Format:   format,
					Level:    level,
				}
			case "email":
				subjectL := make([]string, 0)
				for _, item := range handler.StrField {
					value := item.GetValue()
					if value != "" {
						subjectL = append(subjectL, item.Key+"="+item.GetValue())
					}
				}
				temp = &EmailHandler{
					Level:   level,
					Format:  format,
					Subject: strings.Join(subjectL, ";"),
				}
			}
			tempHandlers = append(tempHandlers, temp)
		}
		tempConfig[key] = LoggingConfig{Handlers: tempHandlers, Opts: opts}
	}
	Logging = tempConfig
}

type LoggingConfig struct {
	Handlers []Handler
	Opts     []zap.Option
}

func (cfg *LoggingConfig) Build() *zap.Logger {
	var Cores []zapcore.Core
	for _, handler := range cfg.Handlers {
		writer, err := handler.BuildWriter()
		if err != nil {
			panic(fmt.Sprintf("%v, %v", err, handler))
		}
		syncer := zapcore.AddSync(writer)
		tempCore := zapcore.NewCore(handler.GetFormat(), syncer, handler.GetLevel())
		Cores = append(Cores, tempCore)
	}
	core := zapcore.NewTee(Cores...)
	return zap.New(core, cfg.Opts...)
}

//
//const (
//	EnvUseDefault           = "USE_DEFAULT"
//	EnvLogDir               = "LOG_DIR"
//	EnvLogLevel             = "LOG_LEVEL"
//	EnvWorkerLogLevel       = "LOG_WORKER_LEVEL"
//	EnvLoggerRotateDuration = "LOGGER_ROTATE_DURATION"
//	EnvLoggerRotateReplica  = "LOGGER_ROTATE_REPLICA"
//	EnvEmailLevel           = "EMAIL_LOG_LEVEL"
//	EnvWithCaller           = "LOG_WITH_CALLER"
//)
//
//var (
//	AccessFile string
//	ErrorFile  string
//	WorkerFile string
//
//	// LOGGER_DURATION 日志的切割间隔.
//	Duration time.Duration
//	// LOGGER_LAYOUT logger需要记录的时间格式.
//	Layout string
//	// LoggerReplica logger 切割需要保留的文件份数
//	Replica int
//
//	// LOG_LEVEL 日志等级. 其中worker的日志等级可以单独设置的.
//	LogLevel       string
//	WorkerLogLevel string
//	EmailLevel     string
//
//	LogWithCaller bool // 是否使用 opts 来构造 logger.
//	Logging       map[string]LoggingConfig
//)
//
//var DEFAULT = os.Getenv(EnvUseDefault) != "false"
//
//func getenv(key, d string) string {
//	res := os.Getenv(key)
//	if DEFAULT {
//		if res == "" {
//			res = d
//		}
//	}
//	return res
//}
//
//func invalidEnv(key string, value ...interface{}) string {
//	return fmt.Sprintf("invalid: %s=%v", key, os.Getenv(key))
//}
//
//func init() {
//	du, err := time.ParseDuration(getenv(EnvLoggerRotateDuration, "24h"))
//	if err != nil {
//		panic(invalidEnv(EnvLoggerRotateDuration))
//	}
//	Duration = du
//
//	if Duration >= time.Hour*24 {
//		Layout = "2006-01-02"
//	} else if Duration >= time.Hour {
//		Layout = "2006-01-02_15"
//	} else if Duration >= time.Minute {
//		Layout = "2006-01-02_15_04"
//	} else {
//		Layout = "2006-01-02_15_04_05"
//	}
//	temp, err := strconv.Atoi(getenv(EnvLoggerRotateReplica, "3"))
//	if err != nil {
//		panic(invalidEnv(EnvLoggerRotateReplica))
//	}
//	Replica = temp
//	LogLevel = getenv(EnvLogLevel, "debug")
//	WorkerLogLevel = getenv(EnvWorkerLogLevel, "info")
//	EmailLevel = getenv(EnvEmailLevel, "error")
//
//	LogDir := getenv(EnvLogDir, "/tmp/")
//	AccessFile = LogDir + "access.log"
//	ErrorFile = LogDir + "error.log"
//	WorkerFile = LogDir + "worker.log"
//
//	LogWithCaller = getenv(EnvWithCaller, "true") == "true"
//	Logging = map[string]LoggingConfig{
//		"app": {
//			Handlers: []Handler{
//				&FileHandler{
//					Filename: "/dev/stdout",
//					Level:    LogLevelMap[LogLevel],
//					Format:   ConsoleFormatter,
//				},
//				&RotateHandler{
//					Filename: AccessFile,
//					Layout:   Layout,
//					Duration: Duration,
//					Replica:  Replica,
//					Level:    LogLevelMap[LogLevel],
//					Format:   ConsoleFormatter,
//				},
//				// error 额外写到 error 文件里.
//				&RotateHandler{
//					Filename: ErrorFile,
//					Layout:   Layout,
//					Duration: Duration,
//					Replica:  Replica,
//					Level:    zapcore.ErrorLevel,
//					Format:   ConsoleFormatter,
//				},
//				&EmailHandler{
//					Level:  LogLevelMap[EmailLevel],
//					Format: JsonFormatter,
//				},
//			},
//			Opts: []zap.Option{
//				zap.AddCaller(), // logger 的行
//				//zap.AddStacktrace(zap.ErrorLevel), // 打印调用栈
//				//zap.Fields(zap.Any("test", "test")), // 初始字段
//			},
//		},
//		"worker": {
//			Handlers: []Handler{
//				&FileHandler{
//					Filename: "/dev/stdout",
//					Level:    LogLevelMap[WorkerLogLevel],
//					Format:   ConsoleFormatter,
//				},
//				&RotateHandler{
//					Filename: WorkerFile,
//					Layout:   Layout,
//					Duration: Duration,
//					Replica:  Replica,
//					Level:    LogLevelMap[WorkerLogLevel],
//					Format:   ConsoleFormatter,
//				},
//				&EmailHandler{
//					Level:  LogLevelMap[EmailLevel],
//					Format: JsonFormatter,
//				},
//			},
//			Opts: []zap.Option{
//				zap.AddCaller(), // logger 的行
//				//zap.AddStacktrace(zap.DebugLevel), // 打印调用栈
//				//zap.Fields(zap.Any("test", "test")), // 初始字段
//			},
//		},
//	}
//}
//

func TestEnv() {
	Logging = map[string]LoggingConfig{
		"app": {
			Handlers: []Handler{
				&FileHandler{
					Filename: "/dev/null",
					Level:    zapcore.DebugLevel,
					Format:   ConsoleFormatter,
				},
				&RotateHandler{
					Filename: "access.log.test",
					Layout:   "2006-01-02",
					Duration: 24 * time.Hour,
					Replica:  3,
					Level:    zapcore.DebugLevel,
					Format:   ConsoleFormatter,
				},
				&RotateHandler{
					Filename: "error.log.test",
					Layout:   "2006-01-02",
					Duration: 24 * time.Hour,
					Replica:  3,
					Level:    zapcore.DebugLevel,
					Format:   ConsoleFormatter,
				},
				&EmailHandler{
					Level:  zapcore.ErrorLevel,
					Format: JsonFormatter,
				},
			},
			Opts: []zap.Option{
				zap.AddCaller(), // logger 的行
			},
		},
		"test": {
			Handlers: []Handler{
				&RotateHandler{
					Filename: "test.log",
					Layout:   "2006-01-02",
					Duration: 24 * time.Hour,
					Replica:  3,
					Level:    zapcore.DebugLevel,
					Format:   ConsoleFormatter,
				},
			},
			Opts: []zap.Option{
				zap.Fields(zap.Field{Key: "test", String: "test", Type: zapcore.StringType}),
			},
		},
	}
}
