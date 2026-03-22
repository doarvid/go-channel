# Platform SDK

一个简洁易用的 Go 库，用于快速集成飞书 (Feishu) / Lark 机器人功能。

## 功能特性

- 🚀 简单易用的 API
- 💬 支持文本、图片、文件消息
- 🎨 支持富卡片消息 (Interactive Cards)
- 🔌 WebSocket 长连接模式 (飞书国内版)
- 🌐 Webhook 模式 (Lark 国际版)
- 🛡️ 用户白名单控制

## 快速开始

### 安装

```bash
go get github.com/doarvid/go-channel
```

### 简单示例

```go
package main

import (
    "context"
    "log"

    sdk "github.com/doarvid/go-channel"
)

func main() {
    // 创建飞书机器人
    bot, err := sdk.NewFeishuBot(sdk.FeishuConfig{
        AppID:     "cli_xxx",
        AppSecret: "xxx",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 设置消息处理器
    bot.OnMessage(func(ctx context.Context, msg *sdk.Message) error {
        // 简单回复
        return msg.Reply("你说: " + msg.Content)
    })

    // 启动机器人
    if err := bot.Start(); err != nil {
        log.Fatal(err)
    }
    defer bot.Stop()

    // 等待中断信号...
    select {}
}
```

### 富卡片消息

```go
// 创建富卡片
card := sdk.NewCard().
    Title("🎉 欢迎使用", "blue").
    Markdown("这是一个**富卡片消息**示例!\n\n支持:").
    Divider().
    Markdown("- Markdown 格式\n- 代码块\n- 按钮").
    Buttons(
        sdk.PrimaryBtn("确定", "act:confirm"),
        sdk.DefaultBtn("取消", "act:cancel"),
    )

// 发送卡片
msg.SendCard(card.Build())
```

## 配置选项

### FeishuConfig

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| AppID | string | 是 | 飞书应用 App ID |
| AppSecret | string | 是 | 飞书应用 App Secret |
| IsLark | bool | 否 | 是否为 Lark 国际版 |
| ReactionEmoji | string | 否 | 处理消息时的反应表情，默认 "OnIt" |
| AllowFrom | string | 否 | 允许的用户 ID 列表，逗号分隔，默认 "*" |
| GroupReplyAll | bool | 否 | 是否回复所有群消息，默认 false |
| ShareSessionInChannel | bool | 否 | 是否在频道内共享会话，默认 false |
| ReplyInThread | bool | 否 | 是否在话题中回复，默认 false |
| ThreadIsolation | bool | 否 | 是否按话题隔离会话，默认 false |
| UseWebhook | bool | 否 | 是否使用 Webhook 模式，默认 false |
| Port | string | 否 | Webhook 服务器端口，默认 "8080" |
| CallbackPath | string | 否 | Webhook 回调路径，默认 "/feishu/webhook" |
| EncryptKey | string | 否 | Webhook 加密密钥 |

## Message 方法

| 方法 | 说明 |
|------|------|
| Reply(content string) | 回复消息 |
| Replyf(format, args...) | 格式化回复 |
| Send(content string) | 发送新消息 |
| Sendf(format, args...) | 格式化发送 |
| SendCard(card *Card) | 发送富卡片 |
| ReplyCard(card *Card) | 回复富卡片 |

## 卡片构建器

链式 API 构建富卡片：

```go
card := sdk.NewCard().
    Title("标题", "blue").              // 设置标题和颜色
    Markdown("**粗体** 文本").         // Markdown 内容
    Divider().                         // 分隔线
    Buttons(                          // 按钮组
        sdk.PrimaryBtn("主按钮", "act:primary"),
        sdk.DefaultBtn("次按钮", "act:default"),
    ).
    ListItem("列表项", "按钮", "act:click"). // 带按钮的列表项
    Note("底部注释").                    // 底部注释
    Build()
```

## 示例

更多示例代码请参考 [examples](./examples/) 目录：

- [simple_echo](./examples/simple_echo/) - 简单的回声机器人
- [rich_card](./examples/rich_card/) - 富卡片消息示例

## 项目结构

```
platform-sdk/
├── sdk.go              # 主 SDK API
├── core/               # 核心类型和接口
│   ├── interfaces.go   # Platform 接口定义
│   ├── types.go        # 消息和附件类型
│   ├── card.go         # 富卡片类型
│   ├── dedup.go        # 消息去重
│   └── registry.go     # 平台注册
├── feishu/             # 飞书平台实现
│   ├── feishu.go       # 主实现
│   ├── card.go         # 卡片渲染
│   └── delete_mode_form.go
└── examples/           # 示例代码
    ├── simple_echo/
    └── rich_card/
```

## License

与原项目相同的许可证。
