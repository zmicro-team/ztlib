package driver

import (
	"context"
	"fmt"

	dySmsApi "github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/zmicro-team/ztlib/limiter/limit_verified"
	"github.com/zmicro-team/ztlib/tidings"
)

type SendMessageRequest struct {
	Mobile        string // require, 手机号
	SignName      string // required, 签名
	TemplateCode  string // required, 模板代号, 如 SMS_xxxxxx
	TemplateParam string // optional, 模板的参数, 如 {"code": "123456"}
}

type SmsAliYun struct {
	signName string
	*dySmsApi.Client
}

func NewSmsAliYun(signName string, c *dySmsApi.Client) *SmsAliYun {
	return &SmsAliYun{
		signName: signName,
		Client:   c,
	}
}

func (*SmsAliYun) Name() string { return "AliYunSms" }

func (sf *SmsAliYun) SendCode(c limit_verified.CodeParam) error {
	_ = c.Kind // tips: 应根据(Kind)业务发送不同的模板代码, 这里不区分业务
	return sf.SendMessage(SendMessageRequest{
		Mobile:        c.Target,
		SignName:      sf.signName,
		TemplateCode:  c.TemplateCode,
		TemplateParam: fmt.Sprintf(`{"code":"%s"}`, c.Code),
	})
}

func (sf *SmsAliYun) Send(ctx context.Context, c *tidings.NoticeParam) error {
	return sf.SendMessage(SendMessageRequest{
		Mobile:        c.Target,
		SignName:      sf.signName,
		TemplateCode:  c.Template,
		TemplateParam: c.TemplateParam,
	})
}

func (sf *SmsAliYun) SendMessage(req SendMessageRequest) error {
	request := dySmsApi.CreateSendSmsRequest()
	request.PhoneNumbers = req.Mobile         // 目标手机号
	request.SignName = req.SignName           // 短信签名名称
	request.TemplateCode = req.TemplateCode   // 短信模板id
	request.TemplateParam = req.TemplateParam // 短信模板变量对应的实际值，JSON格式

	response, err := sf.Client.SendSms(request)
	if err != nil {
		return err
	}
	// 默认流控：使用同一个签名，对同一个手机号码发送短信验证码，
	// 支持1条/分钟，5条/小时 ，累计10条/天
	if response.Code == "isv.BUSINESS_LIMIT_CONTROL" {
		return limit_verified.ErrMaxSendPerDay
	}
	if response.Code != "OK" {
		return tidings.ErrSendFailed
	}
	return nil
}

type SmsDummy struct{}

func (*SmsDummy) Name() string { return "DummySms " }

func (*SmsDummy) SendCode(limit_verified.CodeParam) error { return nil }

func (*SmsDummy) Send(ctx context.Context, c *tidings.NoticeParam) error { return nil }
