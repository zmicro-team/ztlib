package main

import (
	"context"
	"fmt"
	"time"

	basemetrics "github.com/zmicro-team/ztlib/base_metrics"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var _ = sdktrace.AlwaysSample

func main() {
	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("example-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		panic(err)
	}

	// 创建指标提供者
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
	)

	// 创建应用指标实例
	appMetrics, err := basemetrics.NewAppMetrics(
		meterProvider,
		"example-service",
		"1.0.0",
		"prod-us-west-1",
		"production",
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 模拟HTTP请求指标记录
	fmt.Println("Recording HTTP request metrics...")

	// 记录成功请求
	appMetrics.RecordHttpRequest(ctx, "GET", "/api/users", 200, 150*time.Millisecond, false)
	appMetrics.RecordHttpRequest(ctx, "POST", "/api/users", 201, 300*time.Millisecond, false)
	appMetrics.RecordHttpRequest(ctx, "GET", "/api/users/123", 200, 100*time.Millisecond, false)

	// 记录错误请求
	appMetrics.RecordHttpRequest(ctx, "GET", "/api/notfound", 404, 50*time.Millisecond, true)
	appMetrics.RecordHttpRequest(ctx, "POST", "/api/users", 500, 200*time.Millisecond, true)

	// 计算错误率 (2个错误请求 / 5个总请求 = 40%)
	appMetrics.RecordHttpError(ctx, 5, 2)

	fmt.Println("Metrics recorded successfully!")
	fmt.Println("App Name:", appMetrics.AppName)
	fmt.Println("Version:", appMetrics.Version)
	fmt.Println("Cluster:", appMetrics.Cluster)
	fmt.Println("Namespace:", appMetrics.Namespace)
}
