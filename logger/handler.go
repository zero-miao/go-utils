package logger

import (
	"errors"
	"fmt"
	"github.com/zero-miao/go-utils/email"
	"go.uber.org/zap/zapcore"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-05-22 16:36
// @File   : handler.go
// @Project: aotu/logger
// ==========================

type Handler interface {
	BuildWriter() (io.Writer, error)
	GetLevel() zapcore.Level
	GetFormat() zapcore.Encoder
}

type FileHandler struct {
	Filename string
	Level    zapcore.Level
	Format   zapcore.Encoder
}

func (h *FileHandler) BuildWriter() (io.Writer, error) {
	switch h.Filename {
	case "/dev/stdout":
		return os.Stdout, nil
	case "/dev/stderr":
		return os.Stderr, nil
	case os.DevNull:
		return ioutil.Discard, nil
	case "":
		return nil, errors.New("empty filename")
	default:
		return os.OpenFile(h.Filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	}
}

func (h *FileHandler) GetLevel() zapcore.Level {
	return h.Level
}

func (h *FileHandler) GetFormat() zapcore.Encoder {
	return h.Format
}

type EmailHandler struct {
	Level   zapcore.Level
	Format  zapcore.Encoder
	Subject string
}

func (h *EmailHandler) Write(p []byte) (int, error) {
	go func(body string) {
		subject := "邮件报警默认标题"
		if h.Subject != "" {
			subject = h.Subject
		}
		err := email.SendServerMail(subject, body, "text/plain")
		if err != nil {
			SugarL("app").Warnw("日志邮件发送模块异常", "thing", "error", "err", err, "mail_body", string(p))
		}
	}(string(p))
	return 0, nil
}

func (h *EmailHandler) BuildWriter() (io.Writer, error) {
	return h, nil
}

func (h *EmailHandler) GetLevel() zapcore.Level {
	return h.Level
}

func (h *EmailHandler) GetFormat() zapcore.Encoder {
	return h.Format
}

type RotateHandler struct {
	Filename string
	Layout   string
	Duration time.Duration
	Replica  int
	Level    zapcore.Level
	Format   zapcore.Encoder
}

func (h *RotateHandler) BuildWriter() (io.Writer, error) {
	if h.Filename == "" {
		return nil, errors.New("invalid filename: " + h.Filename)
	}
	fmt.Println("NOTICE: 查看是否有同一个文件被初始化两次", h.Filename, h.Duration)
	return RotateWriter(h.Filename, h.Layout, h.Duration, h.Replica)
}

func (h *RotateHandler) GetLevel() zapcore.Level {
	return h.Level
}

func (h *RotateHandler) GetFormat() zapcore.Encoder {
	return h.Format
}
