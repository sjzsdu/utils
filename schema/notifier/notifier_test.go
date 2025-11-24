package notifier

import (
	"os"
	"testing"
)

const testConfig = `
dingtalk:
  enabled: false
  webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=test"
  secret: "test_secret"
  message_type: "text"

email:
  enabled: false
  smtp_host: "smtp.example.com"
  smtp_port: 587
  username: "test@example.com"
  password: "test_password"
  from: "test@example.com"
  to:
    - "user@example.com"
  message_type: "text"

feishu:
  enabled: false
  webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/test"
  secret: "test_secret"
  message_type: "text"
`

func TestLoadAndCreateNotifierManager(t *testing.T) {
	// 创建临时配置文件
	filePath := "/tmp/test_config.yaml"
	err := os.WriteFile(filePath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}
	defer os.Remove(filePath)

	// 测试从文件加载并创建NotifierManager
	manager, err := LoadAndCreateNotifierManager(filePath)
	if err != nil {
		t.Fatalf("创建NotifierManager失败: %v", err)
	}

	// 验证获取已启用的渠道
	channels := manager.GetEnabledChannels()
	// 由于配置中所有通知器都设置为enabled: false，所以应该返回空切片
	if len(channels) != 0 {
		t.Errorf("期望获取到0个已启用渠道，实际获取到: %d", len(channels))
	}
}

func TestManagerSchema_LoadFromBytes(t *testing.T) {
	schema := NewManagerSchema()
	// 测试从字节数组加载配置
	err := schema.LoadFromBytes([]byte(testConfig))
	if err != nil {
		t.Fatalf("从字节数组加载配置失败: %v", err)
	}

	// 测试创建NotifierManager
	manager, err := schema.CreateNotifierManager()
	if err != nil {
		t.Fatalf("创建NotifierManager失败: %v", err)
	}

	// 验证获取已启用的渠道
	channels := manager.GetEnabledChannels()
	if len(channels) != 0 {
		t.Errorf("期望获取到0个已启用渠道，实际获取到: %d", len(channels))
	}
}

func TestManagerSchema_WithEnabledNotifier(t *testing.T) {
	// 测试配置中包含启用的通知器
	enabledConfig := `
dingtalk:
  enabled: true
  webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=test"
  secret: "test_secret"
  message_type: "text"
`

	schema := NewManagerSchema()
	err := schema.LoadFromBytes([]byte(enabledConfig))
	if err != nil {
		t.Fatalf("从字节数组加载配置失败: %v", err)
	}

	manager, err := schema.CreateNotifierManager()
	if err != nil {
		t.Fatalf("创建NotifierManager失败: %v", err)
	}

	// 验证获取已启用的渠道
	channels := manager.GetEnabledChannels()
	if len(channels) != 1 {
		t.Errorf("期望获取到1个已启用渠道，实际获取到: %d", len(channels))
	} else if channels[0] != "dingtalk" {
		t.Errorf("期望获取到dingtalk渠道，实际获取到: %s", channels[0])
	}
}
