package basemetrics

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// ErrMetrics 用于记录应用错误相关的指标
// 错误指标主要包含以下维度：
// - 错误类型（error_type）：标识错误的分类（如http_error, db_error, validation_error等）
// - 错误来源（error_source）：错误发生的组件或位置
// - 错误严重程度（severity）：错误级别（如critical, high, medium, low）
// 基础标签继承自结构体字段：AppName, Version, Cluster, Namespace
type ErrMetrics struct {
	meter                    *sdkmetric.MeterProvider
	errorCounterTotal        metric.Int64Counter
	errorRatePerMinute       metric.Float64Gauge
	errorRecoveryTimeSeconds metric.Float64Histogram
	consecutiveErrors        metric.Int64Gauge

	AppName   string
	Version   string
	Cluster   string
	Namespace string
}

// NewErrMetrics 创建新的错误指标实例
func NewErrMetrics(meterProvider *sdkmetric.MeterProvider, appName, version, cluster, namespace string) (*ErrMetrics, error) {
	meter := meterProvider.Meter(appName)

	// 创建错误计数器
	errorCounterTotal, err := meter.Int64Counter(
		"error_counter_total",
		metric.WithDescription("应用运行期间发生的错误总数"),
	)
	if err != nil {
		return nil, err
	}

	// 创建每分钟错误率测量值
	errorRatePerMinute, err := meter.Float64Gauge(
		"error_rate_per_minute",
		metric.WithDescription("每分钟发生的错误数量"),
	)
	if err != nil {
		return nil, err
	}

	// 创建错误恢复时间直方图
	errorRecoveryTimeSeconds, err := meter.Float64Histogram(
		"error_recovery_time_seconds",
		metric.WithDescription("从错误发生到恢复的时间（秒）"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 5, 10, 30, 60),
	)
	if err != nil {
		return nil, err
	}

	// 创建连续错误计数测量值
	consecutiveErrors, err := meter.Int64Gauge(
		"consecutive_errors",
		metric.WithDescription("连续发生的同类型错误数量"),
	)
	if err != nil {
		return nil, err
	}

	return &ErrMetrics{
		meter:                    meterProvider,
		errorCounterTotal:        errorCounterTotal,
		errorRatePerMinute:       errorRatePerMinute,
		errorRecoveryTimeSeconds: errorRecoveryTimeSeconds,
		consecutiveErrors:        consecutiveErrors,
		AppName:                  appName,
		Version:                  version,
		Cluster:                  cluster,
		Namespace:                namespace,
	}, nil
}

// GetCommonAttributes 获取通用属性
func (em *ErrMetrics) GetCommonAttributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("app_name", em.AppName),
		attribute.String("version", em.Version),
		attribute.String("cluster", em.Cluster),
		attribute.String("namespace", em.Namespace),
	}
}

// RecordError 记录错误发生
func (em *ErrMetrics) RecordError(ctx context.Context, errorType, errorSource, severity string, count int64) {
	attrs := em.GetCommonAttributes()
	attrs = append(attrs,
		attribute.String("error_type", errorType),
		attribute.String("error_source", errorSource),
		attribute.String("severity", severity),
	)

	em.errorCounterTotal.Add(ctx, count, metric.WithAttributes(attrs...))
}

// RecordErrorRate 记录每分钟错误率
func (em *ErrMetrics) RecordErrorRate(ctx context.Context, errorType, errorSource string, rate float64) {
	attrs := em.GetCommonAttributes()
	attrs = append(attrs,
		attribute.String("error_type", errorType),
		attribute.String("error_source", errorSource),
	)

	em.errorRatePerMinute.Record(ctx, rate, metric.WithAttributes(attrs...))
}

// RecordErrorRecoveryTime 记录错误恢复时间
func (em *ErrMetrics) RecordErrorRecoveryTime(ctx context.Context, errorType, recoveryStrategy string, recoveryTimeSeconds float64) {
	attrs := em.GetCommonAttributes()
	attrs = append(attrs,
		attribute.String("error_type", errorType),
		attribute.String("recovery_strategy", recoveryStrategy),
	)

	em.errorRecoveryTimeSeconds.Record(ctx, recoveryTimeSeconds, metric.WithAttributes(attrs...))
}

// RecordConsecutiveErrors 记录连续错误数量
func (em *ErrMetrics) RecordConsecutiveErrors(ctx context.Context, errorType, errorSource string, count int64) {
	attrs := em.GetCommonAttributes()
	attrs = append(attrs,
		attribute.String("error_type", errorType),
		attribute.String("error_source", errorSource),
	)

	em.consecutiveErrors.Record(ctx, count, metric.WithAttributes(attrs...))
}
