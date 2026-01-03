package verification

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// AliyunSMSSender 阿里云短信发送器
type AliyunSMSSender struct {
	accessKeyId     string
	accessKeySecret string
	signName        string
	templateCode    string
}

// NewAliyunSMSSender 创建阿里云短信发送器
func NewAliyunSMSSender(config map[string]string) *AliyunSMSSender {
	return &AliyunSMSSender{
		accessKeyId:     config["accessKeyId"],
		accessKeySecret: config["accessKeySecret"],
		signName:        config["signName"],
		templateCode:    config["templateCode"],
	}
}

// SendCode 发送验证码（通过阿里云短信服务）
func (a *AliyunSMSSender) SendCode(phone, code string) error {
	// TODO: 实现阿里云短信发送逻辑
	// 这里先使用控制台输出作为占位
	logger.Info("通过阿里云发送短信验证码", zap.String("phone", phone), zap.String("code", code))
	fmt.Printf("[Aliyun SMS] 手机号: %s, 验证码: %s\n", phone, code)
	
	// 实际实现示例：
	// client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", a.accessKeyId, a.accessKeySecret)
	// request := requests.NewCommonRequest()
	// request.Method = "POST"
	// request.Domain = "dysmsapi.aliyuncs.com"
	// request.Version = "2017-05-25"
	// request.ApiName = "SendSms"
	// request.QueryParams["PhoneNumbers"] = phone
	// request.QueryParams["SignName"] = a.signName
	// request.QueryParams["TemplateCode"] = a.templateCode
	// request.QueryParams["TemplateParam"] = fmt.Sprintf(`{"code":"%s"}`, code)
	// response, err := client.ProcessCommonRequest(request)
	// return err
	
	return nil
}

// GenerateCode 生成验证码
func (a *AliyunSMSSender) GenerateCode(length int) string {
	return NewSMSSender().GenerateCode(length)
}

// SaveCode 保存验证码到Redis
func (a *AliyunSMSSender) SaveCode(ctx context.Context, rdb *redis.Client, key, code string) error {
	return NewSMSSender().SaveCode(ctx, rdb, key, code)
}

// VerifyCode 验证验证码
func (a *AliyunSMSSender) VerifyCode(ctx context.Context, rdb *redis.Client, key, code string) (bool, error) {
	return NewSMSSender().VerifyCode(ctx, rdb, key, code)
}

// GetCodeType 获取验证码类型
func (a *AliyunSMSSender) GetCodeType() CodeType {
	return CodeTypeSMS
}

