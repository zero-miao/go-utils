package email

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

// ==========================
// @Author : zero-miao
// @Date   : 2019-03-27 10:30
// @File   : configs.go
// @Project: go-gin/email
// ==========================

type Manager struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type Config struct {
	MailServer        string     `yaml:"mail_server"`         // smtp server host
	MailPort          int        `yaml:"mail_port"`           // smtp server port
	MailUser          string     `yaml:"mail_user"`           // smtp server user
	MailPassword      string     `yaml:"mail_password"`       // smtp server password
	MailSubjectPrefix string     `yaml:"mail_subject_prefix"` // 系统邮件会加到前缀到邮件标题前
	DefaultEmailFrom  string     `yaml:"mail_default_from"`   // 系统邮件的发送人.
	Admins            []*Manager `yaml:"mail_admins"`         // admin
	AppendIpInSubject bool       `yaml:"append_ip_subject"`   // 是否添加 ip, 以便多容器时, 查找 bug.
}

func (c *Config) String() string {
	admins := make([]string, 0)
	for _, item := range c.Admins {
		admins = append(admins, item.Name)
	}
	return fmt.Sprintf("%s[%s:%d](%s)", c.MailUser, c.MailServer, c.MailPort, strings.Join(admins, ";"))
}

var GeneralConfig *Config

func YamlInit(data []byte) {
	var temp Config
	err := yaml.Unmarshal(data, &temp)
	if err != nil {
		panic(err)
	}
	GeneralConfig = &temp
}
