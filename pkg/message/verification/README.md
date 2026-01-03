# 验证码服务工厂模式设计

## 设计理念

使用工厂模式（Factory Pattern）管理验证码发送器，通过配置文件动态选择不同的服务商实现，无需修改业务代码。

## 架构设计

```
CodeSender (接口)
    ├── SMSSender (控制台输出 - 测试用)
    ├── AliyunSMSSender (阿里云短信)
    ├── TencentSMSSender (腾讯云短信)
    ├── EmailCodeSender (SMTP邮件)
    └── SendGridEmailSender (SendGrid邮件)

CodeSenderFactory (工厂)
    └── GetCodeSender(codeType) -> CodeSender
```

## 使用方法

### 1. 业务代码中使用

```go
import "github.com/iceymoss/go-hichat-api/pkg/message/verification"

// 获取短信验证码发送器
codeSender, err := verification.GetCodeSender(verification.CodeTypeSMS)
if err != nil {
    return err
}

// 生成验证码
code := codeSender.GenerateCode(6)

// 发送验证码
err = codeSender.SendCode(phone, code)

// 保存验证码到Redis
key := verification.GetRedisKey(verification.CodeTypeSMS, phone)
err = codeSender.SaveCode(ctx, rdb, key, code)

// 验证验证码
pass, err := codeSender.VerifyCode(ctx, rdb, key, code)
```

### 2. 配置文件设置

在 `config/config-local.yaml` 中配置：

```yaml
verification:
  sms:
    provider: 'console'  # console/aliyun/tencent
    config:
      # 阿里云配置
      # accessKeyId: 'your-access-key-id'
      # accessKeySecret: 'your-access-key-secret'
      # signName: 'HiChat'
      # templateCode: 'SMS_123456789'
      
      # 腾讯云配置
      # secretId: 'your-secret-id'
      # secretKey: 'your-secret-key'
      # sdkAppId: 'your-sdk-app-id'
      # signName: 'HiChat'
      # templateId: '123456'
  
  email:
    provider: 'smtp'  # smtp/sendgrid
    config:
      # SendGrid配置
      # apiKey: 'your-sendgrid-api-key'
      # from: 'noreply@hichat.com'
```

## 支持的提供商

### 短信服务商

1. **console** (默认) - 控制台输出，用于测试
2. **aliyun** - 阿里云短信服务
3. **tencent** - 腾讯云短信服务

### 邮件服务商

1. **smtp** (默认) - SMTP邮件服务
2. **sendgrid** - SendGrid邮件服务

## 扩展新的服务商

### 1. 实现 CodeSender 接口

```go
type CustomSMSSender struct {
    // 自定义配置
}

func (c *CustomSMSSender) SendCode(to, code string) error {
    // 实现发送逻辑
}

func (c *CustomSMSSender) GenerateCode(length int) string {
    // 实现生成逻辑
}

// ... 实现其他接口方法
```

### 2. 在工厂中注册

在 `factory.go` 的 `createSMSSender()` 或 `createEmailSender()` 方法中添加：

```go
case "custom":
    return NewCustomSMSSender(config), nil
```

### 3. 更新配置文件

在配置文件中添加对应的配置项。

## 优势

1. **解耦**: 业务代码不依赖具体实现，只依赖接口
2. **灵活**: 通过配置文件切换服务商，无需修改代码
3. **可扩展**: 新增服务商只需实现接口并在工厂中注册
4. **易测试**: 可以使用 console 模式进行测试

## 注意事项

- 工厂使用懒加载模式，首次调用时初始化
- 配置变更需要重启服务才能生效
- 新增服务商时，确保实现所有接口方法

