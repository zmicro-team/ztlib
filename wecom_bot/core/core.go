package core

import (
	"github.com/go-resty/resty/v2"
	"github.com/zmicro-team/ztlib/wecom_bot/consts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// bot

type Config struct {
	Key         string
	IsEnableLog bool
}

type Client struct {
	http   *resty.Client
	config Config
	log    Log
}

func NewClient(cfg Config) *Client {
	c := &Client{
		http:   resty.New(),
		config: cfg,
		log: &defaultLog{
			AppName: "wecombot",
		},
	}
	c.http = c.http.SetBaseURL(consts.BASE_URL)
	setTracerProvider(cfg.Key)
	return c
}

func setTracerProvider(name string) *trace.TracerProvider {
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	return tp
}
