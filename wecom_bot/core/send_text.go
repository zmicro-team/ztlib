package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
	"github.com/zmicro-team/ztlib/wecom_bot/consts"
	"go.opentelemetry.io/otel"
)

func (rt *Client) Send(ctx context.Context, req any, response any) error {
	return rt.Send(ctx, req, response)
}

func Send[Req any, Rsp any](rt *Client, ctx context.Context, req Req, response Rsp) error {
	path := consts.SEND_MODULE
	if fromTraceId(ctx) == "" {
		tracer := otel.GetTracerProvider().Tracer(cast.ToString(rt.config.Key))
		newCtx, span := tracer.Start(ctx, consts.SEND_MODULE)
		ctx = newCtx
		defer span.End()
	}
	r := rt.http.R()
	r.SetHeader("Content-Type", "application/json")
	commonQuery := url.Values{
		"key": {rt.config.Key},
	}
	r = r.SetQueryParamsFromValues(commonQuery)
	r = r.SetBody(req)
	isEnableLog := rt.config.IsEnableLog
	rsp, err := r.Post(consts.SEND_MODULE)
	if err != nil {
		return err
	}
	if rsp.StatusCode() != http.StatusOK {
		return fmt.Errorf("http status code: %d", rsp.RawResponse.StatusCode)
	}
	respBody := rsp.Body()
	if isEnableLog {
		rt.log.InfoHttp(ctx, path, http.MethodPost, rsp.Header(), respBody)
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	return err
}
