package basemetrics

import (
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	// 服务启动时间
	serviceStartupTime = time.Now()
	// 服务开始关闭
	serviceShutdownTime = time.Unix(0, 0)
)

func SetServiceStartupTime(t time.Time) {
	serviceStartupTime = t
}

func SetServiceShutdownTime(t time.Time) {
	serviceShutdownTime = t
}

// GetServiceUptime 获取服务运行时间
func GetServiceUptime() time.Duration {
	return time.Since(serviceStartupTime)
}

// GetServiceShutdownTime 获取服务关闭时间
func GetServiceShutdownTime() time.Duration {
	if serviceShutdownTime.IsZero() {
		serviceShutdownTime = time.Now()
	}
	return time.Since(serviceShutdownTime)
}

type Metrics struct {
	AppMetrics    *AppMetrics
	ErrMetrics    *ErrMetrics
	HealthMetrics *HealthMetrics
}

func New(meterProvider *sdkmetric.MeterProvider, appName, version, cluster, namespace string) (*Metrics, error) {
	base := Metrics{}
	var err error
	base.AppMetrics, err = NewAppMetrics(meterProvider, appName, version, cluster, namespace)
	if err != nil {
		return nil, err
	}
	base.ErrMetrics, err = NewErrMetrics(meterProvider, appName, version, cluster, namespace)
	if err != nil {
		return nil, err
	}
	base.HealthMetrics, err = NewHealthMetrics(meterProvider, appName, version, cluster, namespace)
	if err != nil {
		return nil, err
	}
	return &base, nil
}
