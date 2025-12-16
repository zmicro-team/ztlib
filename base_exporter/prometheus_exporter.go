package base_exporter

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type OtlpMetricHttpConfig struct {
	AuthMode    AuthMode // auth mode (basic_auth)
	UserName    string
	Password    string
	Endpoint    string
	Environment string
	ServiceName string
	Version     string
	Insecure    bool
	Timeout     time.Duration
	Interval    time.Duration
}

func NewPrometheus(c *OtlpMetricHttpConfig) *PrometheusExporter {
	var ctx = context.Background()

	headers := buildAuthHeaders(c.AuthMode, c.UserName, c.Password)

	// 或者使用 OTLP HTTP exporter
	var opts = []otlpmetrichttp.Option{
		otlpmetrichttp.WithHeaders(headers),
		otlpmetrichttp.WithEndpointURL(c.Endpoint),
	}

	if c.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if c.Timeout > 0 {
		opts = append(opts, otlpmetrichttp.WithTimeout(c.Timeout))
	} else {
		opts = append(opts, otlpmetrichttp.WithTimeout(time.Second*5))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to create metric exporter: %v", err))
	}

	readerOption := []sdkmetric.PeriodicReaderOption{}
	if c.Interval > 0 {
		readerOption = append(readerOption, sdkmetric.WithInterval(c.Interval))
	} else {
		readerOption = append(readerOption, sdkmetric.WithInterval(5*time.Second))
	}

	// 创建 Meter Provider
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				readerOption...),
		),
	)

	return &PrometheusExporter{
		Exporter: exporter,
		Provider: provider,
	}
}
