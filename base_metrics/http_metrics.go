package basemetrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// 所有指标应包含的基础标签
// constLabels := prometheus.Labels{
//     "app":        "order-service",
//     "version":    "v1.2.3",
//     "cluster":    "prod-us-west-1",
//     "namespace":  "production",
// }

// 应用指标 app metrics

/*
1. http_requests_total HTTP 请求总数
2. http_error_rate HTTP 错误率
3. http_request_duration_seconds_bucket HTTP 请求响应时间的直方图指标
*/

type AppMetrics struct {
	meter               *sdkmetric.MeterProvider
	httpRequestsTotal   metric.Int64Counter
	httpErrorRate       metric.Float64Gauge
	httpRequestDuration metric.Float64Histogram
	AppName             string
	Version             string
	Cluster             string
	Namespace           string
}

func NewAppMetrics(meterProvider *sdkmetric.MeterProvider, appName, version, cluster, namespace string) (*AppMetrics, error) {
	meter := meterProvider.Meter(appName)

	// HTTP 请求总数
	httpRequestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// HTTP 错误率
	httpErrorRate, err := meter.Float64Gauge(
		"http_error_rate",
		metric.WithDescription("HTTP error rate (4xx and 5xx responses)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// HTTP 请求响应时间直方图
	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &AppMetrics{
		meter:               meterProvider,
		httpRequestsTotal:   httpRequestsTotal,
		httpErrorRate:       httpErrorRate,
		httpRequestDuration: httpRequestDuration,
		AppName:             appName,
		Version:             version,
		Cluster:             cluster,
		Namespace:           namespace,
	}, nil
}

// RecordHttpRequest 记录HTTP请求指标
func (am *AppMetrics) RecordHttpRequest(ctx context.Context, method, path string, code int, duration time.Duration, isClientError bool) {
	// 基础属性
	commonAttrs := []attribute.KeyValue{
		attribute.String("app", am.AppName),
		attribute.String("version", am.Version),
		attribute.String("cluster", am.Cluster),
		attribute.String("namespace", am.Namespace),
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.Int("status_code", code),
	}

	// 记录请求总数
	am.httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(commonAttrs...))

	// 记录响应时间
	am.httpRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(commonAttrs...))
}

// RecordHttpError 计算并记录HTTP错误率 一般不用统计，直接使用平台聚合后计算,例如prometheus rate(http_requests_total{status_code=~"4..|5.."}[5m]) / rate(http_requests_total[5m]) * 100
func (am *AppMetrics) RecordHttpError(ctx context.Context, totalRequests, errorRequests int64) {
	errorRate := float64(0)
	if totalRequests > 0 {
		errorRate = float64(errorRequests) / float64(totalRequests)
	}

	attrs := []attribute.KeyValue{
		attribute.String("app", am.AppName),
		attribute.String("version", am.Version),
		attribute.String("cluster", am.Cluster),
		attribute.String("namespace", am.Namespace),
	}

	am.httpErrorRate.Record(ctx, errorRate, metric.WithAttributes(attrs...))
}

// GetCommonAttributes 获取通用属性
func (am *AppMetrics) GetCommonAttributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("app", am.AppName),
		attribute.String("version", am.Version),
		attribute.String("cluster", am.Cluster),
		attribute.String("namespace", am.Namespace),
	}
}
