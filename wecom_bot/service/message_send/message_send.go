package message_send

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/zmicro-team/ztlib/wecom_bot/consts"
	services "github.com/zmicro-team/ztlib/wecom_bot/service"

	"github.com/zmicro-team/ztlib/wecom_bot/core"
)

type MessageSend services.Service

// SendText 发送文本消息
func (rt *MessageSend) SendText(ctx context.Context, req *TextRequest) (*TextResponse, error) {
	params := struct {
		MsgType string       `json:"msgtype"`
		Text    *TextRequest `json:"text"`
	}{
		MsgType: "text",
		Text:    req,
	}
	resp := &TextResponse{}
	err := core.Send(rt.Client, ctx, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type SendMarkdownRequest struct {
	Content string `json:"content"`
	// MentionedList       []string `json:"mentioned_list,omitempty"`
	// MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type SendMarkdownResponse struct {
	consts.CommonResp `json:",inline"`
}

func (rt *MessageSend) SendMarkdown(ctx context.Context, req *SendMarkdownRequest) (*SendMarkdownResponse, error) {
	params := struct {
		MsgType  string               `json:"msgtype"`
		Markdown *SendMarkdownRequest `json:"markdown"`
	}{
		MsgType:  "markdown",
		Markdown: req,
	}
	if len(params.Markdown.Content) > 4096 {
		return nil, errors.New("消息内容超过限制")
	}
	resp := &SendMarkdownResponse{}
	err := core.Send(rt.Client, ctx, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type sendImageRequest struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

type SendImageResponse struct {
	consts.CommonResp `json:",inline"`
}

func (rt *MessageSend) SendImage(ctx context.Context, r []byte) (*SendImageResponse, error) {
	params := struct {
		MsgType string            `json:"msgtype"`
		Image   *sendImageRequest `json:"image"`
	}{
		MsgType: "image",
		Image:   &sendImageRequest{},
	}
	params.Image.Base64 = base64.StdEncoding.EncodeToString(r)
	hash := md5.New()
	hash.Write(r)
	params.Image.Md5 = hex.EncodeToString(hash.Sum(nil))
	resp := &SendImageResponse{}
	err := core.Send(rt.Client, ctx, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type NewsArticles struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	PicUrl      string `json:"picurl"`
}

type sendNewsRequest struct {
	News []*NewsArticles `json:"articles"` // 图文消息，一个图文消息支持1到8条图文
}

type SendNewsResponse struct {
	consts.CommonResp `json:",inline"`
}

// news 发送图文
func (rt *MessageSend) SendNews(ctx context.Context, req []*NewsArticles) (*SendNewsResponse, error) {
	if req == nil || len(req) == 0 {
		return nil, errors.New("图文消息不能为空")
	}
	if len(req) > 8 {
		return nil, errors.New("图文消息不能超过8条")
	}
	params := struct {
		MsgType string           `json:"msgtype"`
		News    *sendNewsRequest `json:"news"`
	}{
		MsgType: "news",
		News: &sendNewsRequest{
			News: req,
		},
	}
	// params
	resp := &SendNewsResponse{}
	err := core.Send(rt.Client, ctx, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (rt *MessageSend) SendTemplateCard(ctx context.Context, msg *templateCard) (*consts.CommonResp, error) {
	req := newTemplateCardMsg(msg)
	resp := &consts.CommonResp{}
	err := core.Send(rt.Client, ctx, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
