package notifier

import (
	"github.com/sjzsdu/utils/notifier/dingtalk"
	"github.com/sjzsdu/utils/notifier/email"
	"github.com/sjzsdu/utils/notifier/feishu"
	"github.com/sjzsdu/utils/notifier/ntfy"
	"github.com/sjzsdu/utils/notifier/sms"
	"github.com/sjzsdu/utils/notifier/webhook"
	"github.com/sjzsdu/utils/notifier/wecom"
)

// Config 配置文件结构体
type Config struct {
	Dingtalk *dingtalk.DingtalkNotifierConfig `yaml:"dingtalk" json:"dingtalk"`
	Email    *email.EmailNotifierConfig       `yaml:"email" json:"email"`
	Feishu   *feishu.FeishuNotifierConfig     `yaml:"feishu" json:"feishu"`
	NTFY     *ntfy.NtfyNotifierConfig         `yaml:"ntfy" json:"ntfy"`
	SMS      *sms.SMSNotifierConfig           `yaml:"sms" json:"sms"`
	Webhook  *webhook.WebhookNotifierConfig   `yaml:"webhook" json:"webhook"`
	Wecom    *wecom.WecomNotifierConfig       `yaml:"wecom" json:"wecom"`
}
