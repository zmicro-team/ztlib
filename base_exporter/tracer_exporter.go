package base_exporter

import (
	"context"
	"fmt"
	"time"

	pluginmanager "github.com/zmicro-team/ztlib/plugin_manager"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type OtlpTraceHttpConfig struct {
	AuthMode           AuthMode // auth mode (basic_auth)
	UserName           string
	Password           string
	Endpoint           string
	Environment        string
	ServiceName        string
	Version            string
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

type TracerExporter struct {
	conf *OtlpTraceHttpConfig
	// resource             *sdkresource.Resource
	batchOpts            []sdktrace.BatchSpanProcessorOption
	exporter             *otlptrace.Exporter
	TracerProvider       oteltrace.TracerProvider
	TracerErrProvider    oteltrace.TracerProvider
	TextMapPropagator    propagation.TextMapPropagator
	samplingCall         []SamplingCall
	spanProcessorEndCall []SpanProcessorEndCall
	pluginManager        *pluginmanager.PluginManager
}

type SamplingCall func(p sdktrace.SamplingParameters) (*sdktrace.SamplingResult, error)

type SpanProcessorEndCall func(s sdktrace.ReadOnlySpan) bool

// tracer_exporter
func NewTracerExporter(c *OtlpTraceHttpConfig) *TracerExporter {
	var ctx = context.Background()
	headers := buildAuthHeaders(c.AuthMode, c.UserName, c.Password)
	var opts = []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(c.Endpoint),
		otlptracehttp.WithHeaders(headers),
	}

	if c.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if c.Timeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(c.Timeout))
	} else {
		opts = append(opts, otlptracehttp.WithTimeout(time.Second*5))
	}

	if c.RetryConfig != nil && c.RetryConfig.Enabled {
		opts = append(opts,
			otlptracehttp.WithRetry(
				otlptracehttp.RetryConfig{
					Enabled:         c.RetryConfig.Enabled,
					InitialInterval: c.RetryConfig.InitialInterval,
					MaxInterval:     c.RetryConfig.MaxInterval,
					MaxElapsedTime:  c.RetryConfig.MaxElapsedTime,
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
	if c.BatchTimeout > 0 {
		batchOpts = append(batchOpts, sdktrace.WithBatchTimeout(c.BatchTimeout))
	} else {
		batchOpts = append(batchOpts, sdktrace.WithBatchTimeout(5*time.Second))
	}
	if c.MaxExportBatchSize > 0 {
		batchOpts = append(batchOpts, sdktrace.WithMaxExportBatchSize(c.MaxExportBatchSize))
	} else {
		batchOpts = append(batchOpts, sdktrace.WithMaxExportBatchSize(512))
	}

	// 创建TracerExporter实例，包含samplingCall字段的初始化
	exporterInst := &TracerExporter{
		// resource:          res,
		conf:              c,
		batchOpts:         batchOpts,
		exporter:          exporter,
		samplingCall:      make([]SamplingCall, 0), // 初始化为空切片
		TextMapPropagator: propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	}

	// 使用exporterInst作为参数传递给newDefaultSampler
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, batchOpts...),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(newDefaultSampler(exporterInst, c.RatioBased)),
		// 添加自定义的SpanProcessor来过滤
		sdktrace.WithSpanProcessor(newDefaultSpanProcessor(
			exporterInst,
			sdktrace.NewBatchSpanProcessor(exporter),
		)),
	)

	// 设置TracerProvider字段
	exporterInst.TracerProvider = tracerProvider

	return exporterInst
}

func (be *TracerExporter) NewExporter(c *OtlpTraceHttpConfig) (*otlptrace.Exporter, error) {
	var ctx = context.Background()
	headers := buildAuthHeaders(c.AuthMode, c.UserName, c.Password)
	var opts = []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(c.Endpoint),
		otlptracehttp.WithHeaders(headers),
	}

	if c.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if c.Timeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(c.Timeout))
	} else {
		opts = append(opts, otlptracehttp.WithTimeout(time.Second*5))
	}

	if c.RetryConfig != nil && c.RetryConfig.Enabled {
		opts = append(opts,
			otlptracehttp.WithRetry(
				otlptracehttp.RetryConfig{
					Enabled:         c.RetryConfig.Enabled,
					InitialInterval: c.RetryConfig.InitialInterval,
					MaxInterval:     c.RetryConfig.MaxInterval,
					MaxElapsedTime:  c.RetryConfig.MaxElapsedTime,
				},
			),
		)
	}

	return otlptracehttp.New(ctx, opts...)
}

func (be *TracerExporter) NewIndependentErrTracerProvider(sampler sdktrace.Sampler, suffix string) (oteltrace.TracerProvider, error) {
	return be.newIndependentTracerProvider(be.conf, sampler, "errors")
}

// 创建独立的 TracerProvider
func (be *TracerExporter) newIndependentTracerProvider(c *OtlpTraceHttpConfig, sampler sdktrace.Sampler, suffix string) (oteltrace.TracerProvider, error) {
	// 创建新的 exporter
	newExporter, err := be.NewExporter(c)
	if err != nil {
		return nil, err
	}

	// 创建新的 resource（可选：可以修改服务名等）
	ctx := context.Background()
	newRes, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceNameKey.String(fmt.Sprintf("%s-%s-%s", c.ServiceName, c.Environment, suffix)),
			semconv.ServiceVersionKey.String(c.Version),
			semconv.DeploymentEnvironmentKey.String(c.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建新的 TracerProvider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(newExporter, be.batchOpts...),
		sdktrace.WithResource(newRes),
		sdktrace.WithSampler(sampler),
	)

	return tracerProvider, nil
}

// 添加AddSamplingCall方法，用于注册自定义采样函数
func (be *TracerExporter) AddSamplingCall(fn SamplingCall) {
	if be.samplingCall == nil {
		be.samplingCall = make([]SamplingCall, 0)
	}
	be.samplingCall = append(be.samplingCall, fn)
}

// 添加AddSpanProcessorEndCall方法，用于注册自定义SpanProcessorEndCall函数
func (be *TracerExporter) AddSpanProcessorEndCall(fn SpanProcessorEndCall) {
	if be.spanProcessorEndCall == nil {
		be.spanProcessorEndCall = make([]SpanProcessorEndCall, 0)
	}
	be.spanProcessorEndCall = append(be.spanProcessorEndCall, fn)
}

// 添加插件管理方法
func (be *TracerExporter) RegisterPlugin(plugin pluginmanager.CollectorPlugin) error {
	if be.pluginManager == nil {
		be.pluginManager = pluginmanager.NewPluginManager()
	}
	return be.pluginManager.RegisterPlugin(plugin)
}

func (be *TracerExporter) UnregisterPlugin(name string) error {
	if be.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	return be.pluginManager.UnregisterPlugin(name)
}

func (be *TracerExporter) StartPlugins(ctx context.Context) error {
	if be.pluginManager == nil {
		be.pluginManager = pluginmanager.NewPluginManager()
	}
	return be.pluginManager.StartAll(ctx)
}

func (be *TracerExporter) StopPlugins() error {
	if be.pluginManager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	return be.pluginManager.StopAll()
}

func (be *TracerExporter) GetPluginList() []string {
	if be.pluginManager == nil {
		return []string{}
	}
	return be.pluginManager.GetPlugins()
}

func (be *TracerExporter) GetPlugin(name string) (pluginmanager.CollectorPlugin, bool) {
	if be.pluginManager == nil {
		return nil, false
	}
	return be.pluginManager.GetPlugin(name)
}

// 同步事件相关方法
func (be *TracerExporter) GetSyncEvent(name string, value any, timeout time.Duration) (any, error) {
	if be.pluginManager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}
	return be.pluginManager.GetSyncEvent(name, value, timeout)
}

func (be *TracerExporter) GetSyncEvents(pluginNames []string, value any, timeout time.Duration) (map[string]any, error) {
	if be.pluginManager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}
	return be.pluginManager.GetSyncEvents(pluginNames, value, timeout)
}

func (be *TracerExporter) GetAllSyncEvents(value any, timeout time.Duration) (map[string]any, error) {
	if be.pluginManager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}
	return be.pluginManager.GetAllSyncEvents(value, timeout)
}

func (be *TracerExporter) SupportsSyncEvents(name string) bool {
	if be.pluginManager == nil {
		return false
	}
	return be.pluginManager.SupportsSyncEvents(name)
}
