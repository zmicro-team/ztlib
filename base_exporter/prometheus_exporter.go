package base_exporter

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)


type OtlpMetricHttpConfig struct {
	Endpoint string
	Insecure bool
	Timeout  time.Duration
	Interval time.Duration
}


func NewPrometheus(c *Config) *PrometheusExporter {
	var ctx = context.Background()

	headers := buildAuthHeaders(c)

	// 或者使用 OTLP HTTP exporter
	var opts = []otlpmetrichttp.Option{
		otlpmetrichttp.WithHeaders(headers),
		otlpmetrichttp.WithEndpointURL(c.MetricHttpConfig.Endpoint),
	}

	if c.MetricHttpConfig.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if c.MetricHttpConfig.Timeout > 0 {
		opts = append(opts, otlpmetrichttp.WithTimeout(c.MetricHttpConfig.Timeout))
	} else {
		opts = append(opts, otlpmetrichttp.WithTimeout(time.Second*5))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to create metric exporter: %v", err))
	}

	readerOption := []sdkmetric.PeriodicReaderOption{}
	if c.MetricHttpConfig.Interval > 0 {
		readerOption = append(readerOption, sdkmetric.WithInterval(c.MetricHttpConfig.Interval))
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
