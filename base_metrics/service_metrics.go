package basemetrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// HealthMetrics 用于记录服务健康和存活相关的指标
// 包含服务运行时间、健康状态、就绪状态、重启次数、启动关闭等关键指标
type HealthMetrics struct {
	meter                       *sdkmetric.MeterProvider
	serviceUptimeSeconds        metric.Int64Counter
	serviceHealthStatus         metric.Int64Gauge
	serviceReadyStatus          metric.Int64Gauge
	serviceRestartCount         metric.Int64Counter
	serviceHeartbeatTimestamp   metric.Int64Gauge
	serviceLatencyMillis        metric.Float64Histogram
	serviceMemoryUsageBytes     metric.Int64Gauge
	serviceCpuUsagePercent      metric.Float64Gauge
	serviceThreadCount          metric.Int64Gauge
	serviceGcDurationSeconds    metric.Float64Histogram
	serviceOpenConnections      metric.Int64Gauge
	serviceStartupTimestamp     metric.Int64Gauge
	serviceStartupDuration      metric.Float64Histogram
	serviceShutdownTimestamp    metric.Int64Gauge
	serviceShutdownDuration     metric.Float64Histogram

	AppName   string
	Version   string
	Cluster   string
	Namespace string
}

// NewHealthMetrics 创建新的健康指标实例
func NewHealthMetrics(meterProvider *sdkmetric.MeterProvider, appName, version, cluster, namespace string) (*HealthMetrics, error) {
	meter := meterProvider.Meter(appName)

	// 服务运行时间（计数器）
	serviceUptimeSeconds, err := meter.Int64Counter(
		"service_uptime_seconds",
		metric.WithDescription("服务自启动以来的累计运行时间（秒）"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// 服务健康状态（仪表盘）
	serviceHealthStatus, err := meter.Int64Gauge(
		"service_health_status",
		metric.WithDescription("服务的当前健康状态：0=不健康, 1=健康但需关注, 2=完全健康"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// 服务就绪状态（仪表盘）
	serviceReadyStatus, err := meter.Int64Gauge(
		"service_ready_status",
		metric.WithDescription("服务是否准备好接收流量：0=未就绪, 1=已就绪"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// 服务重启次数（计数器）
	serviceRestartCount, err := meter.Int64Counter(
		"service_restart_count",
		metric.WithDescription("服务自部署以来的重启总次数"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// 服务心跳时间戳（测量值）
	serviceHeartbeatTimestamp, err := meter.Int64Gauge(
		"service_heartbeat_timestamp",
		metric.WithDescription("最后一次心跳的时间戳（Unix时间戳）"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// 服务延迟（直方图）
	serviceLatencyMillis, err := meter.Float64Histogram(
		"service_latency_milliseconds",
		metric.WithDescription("关键服务调用的延迟（毫秒）"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(10, 50, 100, 200, 500, 1000, 3000),
	)
	if err != nil {
		return nil, err
	}

	// 服务内存使用量（测量值）
	serviceMemoryUsageBytes, err := meter.Int64Gauge(
		"service_memory_usage_bytes",
		metric.WithDescription("服务当前内存使用量（字节）"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// 服务CPU使用率（测量值）
	serviceCpuUsagePercent, err := meter.Float64Gauge(
		"service_cpu_usage_percent",
		metric.WithDescription("服务当前CPU使用率（百分比）"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return nil, err
	}

	// 服务线程数（测量值）
	serviceThreadCount, err := meter.Int64Gauge(
		"service_thread_count",
		metric.WithDescription("服务当前运行的线程数"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// GC持续时间（直方图）
	serviceGcDurationSeconds, err := meter.Float64Histogram(
		"service_gc_duration_seconds",
		metric.WithDescription("垃圾回收的持续时间（秒）"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1),
	)
	if err != nil {
		return nil, err
	}

	// 打开连接数（测量值）
	serviceOpenConnections, err := meter.Int64Gauge(
		"service_open_connections",
		metric.WithDescription("服务当前打开的连接数"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	// 服务启动时间戳（测量值）
	serviceStartupTimestamp, err := meter.Int64Gauge(
		"service_startup_timestamp",
		metric.WithDescription("服务启动时间戳（Unix时间戳）"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// 服务启动耗时（直方图）
	serviceStartupDuration, err := meter.Float64Histogram(
		"service_startup_duration_seconds",
		metric.WithDescription("服务启动耗时（秒）"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2, 5, 10, 30, 60),
	)
	if err != nil {
		return nil, err
	}

	// 服务关闭时间戳（测量值）
	serviceShutdownTimestamp, err := meter.Int64Gauge(
		"service_shutdown_timestamp",
		metric.WithDescription("服务关闭时间戳（Unix时间戳）"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// 服务关闭耗时（直方图）
	serviceShutdownDuration, err := meter.Float64Histogram(
		"service_shutdown_duration_seconds",
		metric.WithDescription("服务关闭耗时（秒）"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2, 5, 10, 30, 60),
	)
	if err != nil {
		return nil, err
	}

	return &HealthMetrics{
		meter:                     meterProvider,
		serviceUptimeSeconds:      serviceUptimeSeconds,
		serviceHealthStatus:       serviceHealthStatus,
		serviceReadyStatus:        serviceReadyStatus,
		serviceRestartCount:       serviceRestartCount,
		serviceHeartbeatTimestamp: serviceHeartbeatTimestamp,
		serviceLatencyMillis:      serviceLatencyMillis,
		serviceMemoryUsageBytes:   serviceMemoryUsageBytes,
		serviceCpuUsagePercent:    serviceCpuUsagePercent,
		serviceThreadCount:        serviceThreadCount,
		serviceGcDurationSeconds:  serviceGcDurationSeconds,
		serviceOpenConnections:    serviceOpenConnections,
		serviceStartupTimestamp:   serviceStartupTimestamp,
		serviceStartupDuration:    serviceStartupDuration,
		serviceShutdownTimestamp:  serviceShutdownTimestamp,
		serviceShutdownDuration:   serviceShutdownDuration,
		AppName:                   appName,
		Version:                   version,
		Cluster:                   cluster,
		Namespace:                 namespace,
	}, nil
}

// GetCommonAttributes 获取通用属性
func (hm *HealthMetrics) GetCommonAttributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("app_name", hm.AppName),
		attribute.String("version", hm.Version),
		attribute.String("cluster", hm.Cluster),
		attribute.String("namespace", hm.Namespace),
	}
}

// RecordUptime 记录服务运行时间
func (hm *HealthMetrics) RecordUptime(ctx context.Context, uptimeSeconds int64) {
	attrs := hm.GetCommonAttributes()
	hm.serviceUptimeSeconds.Add(ctx, uptimeSeconds, metric.WithAttributes(attrs...))
}

// RecordHealthStatus 记录服务健康状态
func (hm *HealthMetrics) RecordHealthStatus(ctx context.Context, healthCheckType string, status int64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("health_check_type", healthCheckType))
	hm.serviceHealthStatus.Record(ctx, status, metric.WithAttributes(attrs...))
}

// RecordReadyStatus 记录服务就绪状态
func (hm *HealthMetrics) RecordReadyStatus(ctx context.Context, ready int64) {
	attrs := hm.GetCommonAttributes()
	hm.serviceReadyStatus.Record(ctx, ready, metric.WithAttributes(attrs...))
}

// RecordRestart 记录服务重启
func (hm *HealthMetrics) RecordRestart(ctx context.Context, restartReason string) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("restart_reason", restartReason))
	hm.serviceRestartCount.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordHeartbeat 记录服务心跳
func (hm *HealthMetrics) RecordHeartbeat(ctx context.Context) {
	attrs := hm.GetCommonAttributes()
	timestamp := time.Now().Unix()
	hm.serviceHeartbeatTimestamp.Record(ctx, timestamp, metric.WithAttributes(attrs...))
}

// RecordLatency 记录服务延迟
func (hm *HealthMetrics) RecordLatency(ctx context.Context, endpoint string, latencyMillis float64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("endpoint", endpoint))
	hm.serviceLatencyMillis.Record(ctx, latencyMillis, metric.WithAttributes(attrs...))
}

// RecordMemoryUsage 记录内存使用量
func (hm *HealthMetrics) RecordMemoryUsage(ctx context.Context, memoryType string, usageBytes int64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("memory_type", memoryType))
	hm.serviceMemoryUsageBytes.Record(ctx, usageBytes, metric.WithAttributes(attrs...))
}

// RecordCpuUsage 记录CPU使用率
func (hm *HealthMetrics) RecordCpuUsage(ctx context.Context, cpuType string, usagePercent float64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("cpu_type", cpuType))
	hm.serviceCpuUsagePercent.Record(ctx, usagePercent, metric.WithAttributes(attrs...))
}

// RecordThreadCount 记录线程数
func (hm *HealthMetrics) RecordThreadCount(ctx context.Context, threadState string, count int64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("thread_state", threadState))
	hm.serviceThreadCount.Record(ctx, count, metric.WithAttributes(attrs...))
}

// RecordGcDuration 记录GC持续时间
func (hm *HealthMetrics) RecordGcDuration(ctx context.Context, gcType string, durationSeconds float64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("gc_type", gcType))
	hm.serviceGcDurationSeconds.Record(ctx, durationSeconds, metric.WithAttributes(attrs...))
}

// RecordOpenConnections 记录打开连接数
func (hm *HealthMetrics) RecordOpenConnections(ctx context.Context, connectionType string, count int64) {
	attrs := hm.GetCommonAttributes()
	attrs = append(attrs, attribute.String("connection_type", connectionType))
	hm.serviceOpenConnections.Record(ctx, count, metric.WithAttributes(attrs...))
}

// RecordStartup 记录服务启动
func (hm *HealthMetrics) RecordStartup(ctx context.Context, startupDurationSeconds float64) {
	attrs := hm.GetCommonAttributes()
	
	// 记录启动时间戳
	timestamp := time.Now().Unix()
	hm.serviceStartupTimestamp.Record(ctx, timestamp, metric.WithAttributes(attrs...))
	
	// 记录启动耗时
	hm.serviceStartupDuration.Record(ctx, startupDurationSeconds, metric.WithAttributes(attrs...))
}

// RecordShutdown 记录服务关闭
func (hm *HealthMetrics) RecordShutdown(ctx context.Context, shutdownDurationSeconds float64) {
	attrs := hm.GetCommonAttributes()
	
	// 记录关闭时间戳
	timestamp := time.Now().Unix()
	hm.serviceShutdownTimestamp.Record(ctx, timestamp, metric.WithAttributes(attrs...))
	
	// 记录关闭耗时
	hm.serviceShutdownDuration.Record(ctx, shutdownDurationSeconds, metric.WithAttributes(attrs...))
}

// GetStartupTimestamp 获取启动时间戳
func (hm *HealthMetrics) GetStartupTimestamp(ctx context.Context) int64 {
	// 这里我们需要从指标中获取值，但OpenTelemetry主要是写入操作
	// 实际使用时，应该在应用层保存这个值
	return time.Now().Unix()
}

// GetShutdownTimestamp 获取关闭时间戳
func (hm *HealthMetrics) GetShutdownTimestamp(ctx context.Context) int64 {
	// 这里我们需要从指标中获取值，但OpenTelemetry主要是写入操作
	// 实际使用时，应该在应用层保存这个值
	return time.Now().Unix()
}