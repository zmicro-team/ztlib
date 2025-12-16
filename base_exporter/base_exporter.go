package base_exporter

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type AuthMode string

const (
	BasicAuth  AuthMode = "basic_auth"
	BearerAuth AuthMode = "bearer_auth"
)

type Config struct {
	AuthMode AuthMode // auth mode (basic_auth)
	UserName string
	Password string
	// Endpoint         string
	Environment      string
	ServiceName      string
	Version          string
	TraceHttpConfig  *OtlpTraceHttpConfig
	MetricHttpConfig *OtlpMetricHttpConfig
}

type OtlpTraceHttpConfig struct {
	Endpoint           string
	Insecure           bool
	Timeout            time.Duration
	BatchTimeout       time.Duration
	MaxExportBatchSize int
	RatioBased         float64
	RetryConfig        *OtlpTraceHttpRetryConfig
}

type OtlpTraceHttpRetryConfig struct {
	Enabled         bool
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
}

type BaseExporter struct {
	Config             *Config
	TracerExporter     *TracerExporter
	PrometheusExporter *PrometheusExporter
}

type TracerExporter struct {
	TracerProvider    oteltrace.TracerProvider
	TextMapPropagator propagation.TextMapPropagator
}

type PrometheusExporter struct {
	Exporter *otlpmetrichttp.Exporter
	Provider *sdkmetric.MeterProvider
}

// buildAuthHeaders 独立出验证方式，根据认证模式构建HTTP头
func buildAuthHeaders(c *Config) map[string]string {
	headers := make(map[string]string)

	switch c.AuthMode {
	case BasicAuth:
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(c.UserName+":"+c.Password))
	case BearerAuth:
		headers["Authorization"] = "Bearer " + c.Password
	default:
		panic("unsupported auth mode")
	}

	headers["X-Application-Name"] = c.ServiceName
	headers["X-Environment"] = c.Environment

	return headers
}

func NewBaseExporter(c *Config) *BaseExporter {
	return &BaseExporter{
		Config:             c,
		TracerExporter:     NewTracerExporter(c),
		PrometheusExporter: NewPrometheus(c),
	}
}

func (be *BaseExporter) GetName() string {
	return fmt.Sprintf("%s-%s", be.Config.ServiceName, be.Config.Environment)
}

func NewTracerExporter(c *Config) *TracerExporter {
	var ctx = context.Background()

	headers := buildAuthHeaders(c)

	var opts = []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(c.TraceHttpConfig.Endpoint),
		otlptracehttp.WithHeaders(headers),
	}

	if c.TraceHttpConfig.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if c.TraceHttpConfig.Timeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(c.TraceHttpConfig.Timeout))
	} else {
		opts = append(opts, otlptracehttp.WithTimeout(time.Second*5))
	}

	if c.TraceHttpConfig.RatioBased <= 0 {
		c.TraceHttpConfig.RatioBased = 1.0
	}

	if c.TraceHttpConfig.RetryConfig != nil && c.TraceHttpConfig.RetryConfig.Enabled {
		opts = append(opts,
			otlptracehttp.WithRetry(
				otlptracehttp.RetryConfig{
					Enabled:         c.TraceHttpConfig.RetryConfig.Enabled,
					InitialInterval: c.TraceHttpConfig.RetryConfig.InitialInterval,
					MaxInterval:     c.TraceHttpConfig.RetryConfig.MaxInterval,
					MaxElapsedTime:  c.TraceHttpConfig.RetryConfig.MaxElapsedTime,
				},
			),
		)
	}

	var exporter, err = otlptracehttp.New(ctx, opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to create trace exporter: %v", err))
	}

	// 创建资源
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceNameKey.String(fmt.Sprintf("%s-%s", c.ServiceName, c.Environment)),
			semconv.ServiceVersionKey.String(c.Version),
			semconv.DeploymentEnvironmentKey.String(c.Environment),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create resource: %v", err))
	}

	// 创建 TracerProvider
	batchOpts := []sdktrace.BatchSpanProcessorOption{}
	if c.TraceHttpConfig.BatchTimeout > 0 {
		batchOpts = append(batchOpts, sdktrace.WithBatchTimeout(c.TraceHttpConfig.BatchTimeout))
	} else {
		batchOpts = append(batchOpts, sdktrace.WithBatchTimeout(5*time.Second))
	}

	if c.TraceHttpConfig.MaxExportBatchSize > 0 {
		batchOpts = append(batchOpts, sdktrace.WithMaxExportBatchSize(c.TraceHttpConfig.MaxExportBatchSize))
	} else {
		batchOpts = append(batchOpts, sdktrace.WithMaxExportBatchSize(512))
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, batchOpts...),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.TraceHttpConfig.RatioBased)),
	)

	return &TracerExporter{
		TracerProvider:    tracerProvider,
		TextMapPropagator: propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	}
}

func FromTraceId(ctx context.Context) string {
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		return sc.TraceID().String()
	}
	return ""
}
