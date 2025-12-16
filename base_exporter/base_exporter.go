package base_exporter

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
)

type AuthMode string

const (
	BasicAuth  AuthMode = "basic_auth"
	BearerAuth AuthMode = "bearer_auth"
)

type Config struct {
	// AuthMode AuthMode // auth mode (basic_auth)
	// UserName string
	// Password string
	Environment string
	ServiceName string
	Version     string

	TraceHttpConfig  *OtlpTraceHttpConfig
	MetricHttpConfig *OtlpMetricHttpConfig
}

type BaseExporter struct {
	Config             *Config
	TracerExporter     *TracerExporter
	PrometheusExporter *PrometheusExporter
}

type PrometheusExporter struct {
	Exporter *otlpmetrichttp.Exporter
	Provider *sdkmetric.MeterProvider
}

// buildAuthHeaders 独立出验证方式，根据认证模式构建HTTP头
func buildAuthHeaders(auth AuthMode, userName, password string) map[string]string {
	headers := make(map[string]string)

	switch auth {
	case BasicAuth:
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(userName+":"+password))
	case BearerAuth:
		headers["Authorization"] = "Bearer " + password
	default:
		panic("unsupported auth mode")
	}
	return headers
}

func NewBaseExporter(c *Config) *BaseExporter {
	c.TraceHttpConfig.Environment = c.Environment
	c.TraceHttpConfig.ServiceName = c.ServiceName
	c.TraceHttpConfig.Version = c.Version
	c.MetricHttpConfig.Environment = c.Environment
	c.MetricHttpConfig.ServiceName = c.ServiceName
	c.MetricHttpConfig.Version = c.Version
	return &BaseExporter{
		Config:             c,
		TracerExporter:     NewTracerExporter(c.TraceHttpConfig),
		PrometheusExporter: NewPrometheus(c.MetricHttpConfig),
	}
}

func (be *BaseExporter) GetName() string {
	return fmt.Sprintf("%s-%s", be.Config.ServiceName, be.Config.Environment)
}

func FromTraceId(ctx context.Context) string {
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		return sc.TraceID().String()
	}
	return ""
}
