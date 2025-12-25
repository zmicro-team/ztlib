# Base Metrics Package

这个包包含了基于 OpenTelemetry 的基础指标收集库，提供了三个主要的指标类别：

## 文件结构

### http_metrics.go
- **结构体**: `AppMetrics`
- **功能**: 应用程序级别的 HTTP 指标
- **包含指标**:
  - `http_requests_total`: HTTP 请求总数
  - `http_error_rate`: HTTP 错误率
  - `http_request_duration_seconds`: HTTP 请求响应时间直方图
- **适用场景**: Web 应用、API 服务的请求级别监控

### service_metrics.go
- **结构体**: `HealthMetrics`
- **功能**: 服务健康状态、资源使用和生命周期指标
- **包含指标**:
  - `service_uptime_seconds`: 服务运行时间
  - `service_health_status`: 服务健康状态
  - `service_ready_status`: 服务就绪状态
  - `service_restart_count`: 服务重启次数
  - `service_heartbeat_timestamp`: 服务心跳时间戳
  - `service_latency_milliseconds`: 服务延迟
  - `service_memory_usage_bytes`: 内存使用量
  - `service_cpu_usage_percent`: CPU 使用率
  - `service_thread_count`: 线程数
  - `service_gc_duration_seconds`: GC 持续时间
  - `service_open_connections`: 打开连接数
  - `service_startup_timestamp`: 服务启动时间戳
  - `service_startup_duration_seconds`: 服务启动耗时
  - `service_shutdown_timestamp`: 服务关闭时间戳
  - `service_shutdown_duration_seconds`: 服务关闭耗时
- **适用场景**: 微服务健康检查、性能监控、资源管理、生命周期管理

### err_metrics.go
- **结构体**: `ErrMetrics`
- **功能**: 错误相关的指标
- **包含指标**:
  - `error_counter_total`: 错误总数
  - `error_rate_per_minute`: 每分钟错误率
  - `error_recovery_time_seconds`: 错误恢复时间
  - `consecutive_errors`: 连续错误数量
- **适用场景**: 错误监控、故障恢复、系统稳定性分析

## 使用方法

所有指标结构体都包含相同的基础标签：
- `app_name`: 应用名称
- `version`: 应用版本
- `cluster`: 集群名称
- `namespace`: 命名空间

创建指标实例的示例：
```go
// 创建应用指标
appMetrics, err := basemetrics.NewAppMetrics(meterProvider, "my-app", "v1.0.0", "prod", "default")

// 创建健康指标
healthMetrics, err := basemetrics.NewHealthMetrics(meterProvider, "my-app", "v1.0.0", "prod", "default")

// 创建错误指标
errMetrics, err := basemetrics.NewErrMetrics(meterProvider, "my-app", "v1.0.0", "prod", "default")
```

### 启动和关闭指标使用示例

```go
// 服务启动时
func startService() {
    startTime := time.Now()
    
    // ... 执行启动逻辑 ...
    
    startupDuration := time.Since(startTime).Seconds()
    
    // 记录启动指标
    healthMetrics.RecordStartup(ctx, startupDuration)
    healthMetrics.RecordReadyStatus(ctx, 1) // 1 = 就绪
    healthMetrics.RecordHeartbeat(ctx)
}

// 优雅关闭时
func shutdownService() {
    shutdownStartTime := time.Now()
    
    // ... 执行关闭逻辑 ...
    
    shutdownDuration := time.Since(shutdownStartTime).Seconds()
    
    // 记录关闭指标
    healthMetrics.RecordReadyStatus(ctx, 0) // 0 = 未就绪
    healthMetrics.RecordShutdown(ctx, shutdownDuration)
}
```

## 设计原则

1. **一致性**: 所有指标使用相同的标签体系
2. **可扩展性**: 易于添加新的指标类型
3. **标准化**: 遵循 OpenTelemetry 标准
4. **易用性**: 简单的 API 接口