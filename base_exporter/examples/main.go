package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	base_exporter "github.com/zmicro-team/ztlib/base_exporter"
	basemetrics "github.com/zmicro-team/ztlib/base_metrics"
	pluginmanager "github.com/zmicro-team/ztlib/plugin_manager"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv2 "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.opentelemetry.io/otel/trace"
)

// fs导入一个配置目录
//
//go:embed conf/_test_config.yaml conf/conf.yaml
var config embed.FS
var pluginManager = pluginmanager.NewPluginManager()

// 模拟一个gin
func main() {
	var ctx = context.Background()
	basemetrics.SetServiceStartupTime(time.Now())
	eventPlugin := pluginmanager.NewEventPlugin(
		"test-event",
		pluginmanager.DefaultEventPluginConfig(),
	)
	pluginManager.RegisterPlugin(eventPlugin)
	pluginManager.StartAll(ctx)
	r := gin.Default()
	r.Use(
		testGin(),
		func(c *gin.Context) {
			c.Header("X-Trace-Id", base_exporter.FromTraceId(c.Request.Context()))
			c.Next()
		},
	)

	gr := r.Group("/api")
	// curl -X GET http://localhost:8380/api/v1
	gr.GET("/v1", func(c *gin.Context) {
		c.String(200, "hello")
	})
	// curl -X GET http://localhost:8380/api/error
	gr.GET("/error", func(c *gin.Context) {
		// 返回错误
		c.String(500, "error")
	})
	// curl
	// curl -X GET http://localhost:8380/
	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello")
	})
	// 生成curl
	// curl -X GET http://localhost:8380/error
	r.GET("/error", func(c *gin.Context) {
		// 返回错误

		pluginManager.StopAll()
		c.String(500, "error")
	})
	basemetrics.SetServiceShutdownTime(time.Now())
	r.Run(":8381")
}

func testGin() func(c *gin.Context) {
	exporterConf := &base_exporter.Config{}
	f, err := config.Open("conf/_test_config.yaml")
	if err != nil {
		panic(err.Error())
	}
	vip := viper.New()
	vip.SetConfigType("yaml")
	err = vip.ReadConfig(f)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%v\n", vip.Get("exporter"))
	err = vip.UnmarshalKey("exporter", exporterConf)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%v\n", exporterConf)
	exporter := base_exporter.NewBaseExporter(exporterConf)

	// 专门传递错误的导出器
	errTracerProvider, err :=
		exporter.TracerExporter.NewIndependentErrTracerProvider(sdktrace.AlwaysSample(), "errors")

	if err != nil {
		panic(err.Error())
	}

	baseMetrics, err := basemetrics.New(exporter.PrometheusExporter.Provider,
		exporter.GetName(),
		exporter.Config.Version,
		exporter.Config.Environment,
		exporter.Config.Environment)
	if err != nil {
		panic(err.Error())
	}

	baseMetrics.HealthMetrics.
		RecordStartup(context.Background(), (54000 * time.Millisecond).Seconds())

	var cfgTracerProvider = exporter.TracerExporter.TracerProvider
	tracer := cfgTracerProvider.Tracer(
		"gin",
	)
	var cfgPropagators = exporter.TracerExporter.TextMapPropagator
	// 订阅关闭事件
	go func() {
		ctx := context.Background()
		p, b := pluginManager.GetPlugin("test-event")
		if !b {
			log.Default().Printf("plugin not found")
			return
		}
		eventPlugin := p.(*pluginmanager.EventPlugin)
		eventCh, unsubscribe := eventPlugin.SubscribeToEvents(pluginmanager.EventTypeShutdown)
		defer unsubscribe() // 确保取消订阅
		startupDuration := basemetrics.GetServiceShutdownTime()
		baseMetrics.HealthMetrics.RecordStartup(ctx, startupDuration.Seconds())
		data := <-eventCh
		baseMetrics.HealthMetrics.RecordShutdown(ctx,
			basemetrics.GetServiceShutdownTime().Seconds())

		log.Default().Printf("shutdown %v \n", data)
		exporter.PrometheusExporter.Provider.ForceFlush(ctx)
		time.Sleep(2 * time.Second)
		exporter.PrometheusExporter.Provider.Shutdown(ctx)
	}()
	// 心跳订阅
	go func() {
		ctx := context.Background()
		p, b := pluginManager.GetPlugin("test-event")
		if !b {
			log.Default().Printf("plugin not found")
			return
		}
		eventPlugin := p.(*pluginmanager.EventPlugin)
		eventCh, unsubscribe := eventPlugin.SubscribeToEvents(pluginmanager.EventTypeTick)
		_ = unsubscribe
		for data := range eventCh {
			log.Default().Printf("tick %v \n", data)
			baseMetrics.HealthMetrics.RecordHeartbeat(ctx)
		}
	}()
	return func(c *gin.Context) {
		fmt.Println(base_exporter.FromTraceId(c))
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()
		ctx := cfgPropagators.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
		opts := []trace.SpanStartOption{
			trace.WithAttributes(
				semconv2.URLScheme(c.Request.URL.Scheme),
				semconv2.HostName(c.Request.Host),
				semconv2.URLFull(c.FullPath()),
				semconv2.HTTPRequestMethodOriginal(c.Request.Method),
				semconv2.ServiceName(exporter.GetName()),
				semconv2.UserAgentOriginal(c.Request.UserAgent()),
				semconv2.ClientAddress(c.ClientIP()),
				semconv2.ServiceVersion(exporter.Config.Version),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		}
		spanName := c.FullPath()
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()
		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)
		fmt.Println("1", base_exporter.FromTraceId(c.Request.Context()))
		// serve the request to the next middleware

		c.Next()

		status := c.Writer.Status()
		spanStatus := codes.Code(status)
		span.SetStatus(spanStatus, spanStatus.String())
		span.SetAttributes(
			semconv2.HTTPResponseStatusCode(status),
		)
		if status >= 400 {
			// 导出到另外一个导出器
			pid := base_exporter.FromTraceId(c.Request.Context())
			fmt.Println("2", pid)
			errTracer := errTracerProvider.Tracer("gin-err")
			_, errSpan := errTracer.Start(ctx, spanName, opts...)
			errSpan.SetStatus(spanStatus, spanStatus.String())
			errSpan.SetAttributes(
				semconv2.HTTPResponseStatusCode(status),
			)
			defer errSpan.End()
		}
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}
	}
}
