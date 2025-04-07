package core

import (
	"context"
	"net/http"

	"github.com/zmicro-team/zmicro/core/log"

	"go.opentelemetry.io/otel/trace"
)

type Log interface {
	InfoHttp(ctx context.Context, url string, method string, header http.Header, body []byte)
}

type defaultLog struct {
	AppName string
}

func (dl *defaultLog) InfoHttp(ctx context.Context, url string, method string, header http.Header, body []byte) {
	log.Info(dl.AppName,
		log.String("traceId", fromTraceId(ctx)),
		log.String("url", url),
		log.String("method", method),
		log.Any("header", header),
		log.String("body", string(body)),
	)
}

func fromTraceId(ctx context.Context) string {
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		return sc.TraceID().String()
	}
	return ""
}
