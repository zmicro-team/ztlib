package consts

import "fmt"

// https://qyapi.weixin.qq.com
const (
	BASE_URL    = "https://qyapi.weixin.qq.com"
	SEND_MODULE = "cgi-bin/webhook/send"
)

type Media struct {
	// 文件id, 通过文件上传接口获取
	ID string `json:"media_id"`
}

type ICommonResp interface {
	IsSuccess() bool
	Error() error
}

// "body": "{\"errcode\":0,\"errmsg\":\"ok\"}"
type CommonResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (c CommonResp) IsSuccess() bool {
	return c.Errcode == 0
}
func (c CommonResp) Error() error {
	return fmt.Errorf("errcode: %d, errmsg: %s", c.Errcode, c.Errmsg)
}
