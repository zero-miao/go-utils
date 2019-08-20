package email

import (
	"crypto/tls"
	"errors"
	"gopkg.in/gomail.v2"
	"net"
	"os"
	"strings"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-03-27 10:30
// @File   : email.go
// @Project: go-gin/email
// ==========================

func getMailDialer() *gomail.Dialer {
	if GeneralConfig == nil {
		panic("`email` not init yet")
	}
	server := gomail.NewDialer(GeneralConfig.MailServer, GeneralConfig.MailPort, GeneralConfig.MailUser, GeneralConfig.MailPassword)
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return server
}

// SendEmail 核心邮件函数, 封装了对发送者, 接收者的信息的封装.
// tips: 登录邮件服务器的邮箱和From中的邮箱必须一致. 因此只能自定义名称
// 以服务器配置的smtp信息发送邮件.
// args:
// 		fromUserName 发送方用户名,
// 		to 收件人列表,
// 		cc 抄送人列表,
// 		bcc 秘密抄送列表,
// 		subject 主题,
// 		body, bodyType 邮件主体,
// 		attaches 邮件附件, attach和embed区别在于 disposition=attachment和inline. 显示上没有区别.
// return: 发送邮件时产生的错误.
func SendEmail(fromUserName string, to []*Manager, cc []*Manager, bcc []*Manager, subject, body, bodyType string, attaches ...string) error {
	if GeneralConfig == nil {
		panic("`email` not init yet")
	}
	msg := gomail.NewMessage()
	if fromUserName == "" {
		fromUserName = GeneralConfig.MailUser
	}
	msg.SetHeader("From", msg.FormatAddress(GeneralConfig.MailUser, fromUserName))

	_to := make([]string, 0)
	if to == nil {
		return errors.New("no recipient supported")
	}
	for _, item := range to {
		_to = append(_to, msg.FormatAddress(item.Email, item.Name))
	}
	msg.SetHeader("To", _to...)

	if GeneralConfig.AppendIpInSubject {
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			res := make([]string, 0, len(addrs))
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
					res = append(res, ipnet.String())
				}
			}
			appendSub := strings.Join(res, ";")
			subject += "-" + appendSub
		}
	}

	msg.SetHeader("Subject", subject)

	if cc != nil {
		_cc := make([]string, 0)
		for _, item := range cc {
			_cc = append(_cc, msg.FormatAddress(item.Email, item.Name))
		}
		msg.SetHeader("Cc", _cc...)
	}
	if bcc != nil {
		_bcc := make([]string, 0)
		for _, item := range bcc {
			_bcc = append(_bcc, msg.FormatAddress(item.Email, item.Name))
		}
		msg.SetHeader("Bcc", _bcc...)
	}
	msg.SetBody(bodyType, body)

	// https://tools.ietf.org/html/rfc4021#section-2.2
	for _, attach := range attaches {
		if _, err := os.Stat(attach); err != nil {
			return err
		}
		msg.Attach(attach)
	}

	dialer := getMailDialer()
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

// SendServerMail 利用系統邮件账户发送邮件; 发送给系统管理员.
func SendServerMail(subject, body, bodyType string) error {
	if GeneralConfig == nil {
		panic("`email` not init yet")
	}
	if GeneralConfig.AppendIpInSubject {
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			res := make([]string, 0, len(addrs))
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
					res = append(res, ipnet.String())
				}
			}
			appendSub := strings.Join(res, ";")
			subject += "-" + appendSub
		}
	}
	return SendEmail(GeneralConfig.DefaultEmailFrom, GeneralConfig.Admins, nil, nil, GeneralConfig.MailSubjectPrefix+subject, body, bodyType)
}
