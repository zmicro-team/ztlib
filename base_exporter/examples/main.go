package main

import (
	"embed"
	"fmt"
	"os"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	base_exporter "github.com/zmicro-team/ztlib/base_exporter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv2 "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.opentelemetry.io/otel/trace"
)

// fs导入一个配置目录
//
//go:embed conf/_test_config.yaml conf/conf.yaml
var config embed.FS

// 模拟一个gin
func main() {
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
		c.String(500, "error")
	})

	r.Run(":8380")
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
	meter := exporter.PrometheusExporter.Provider.Meter(exporter.GetName())
	// *** 关键修改：在中间件外部预先创建指标 ***
	requestCounter, err := meter.Int64Counter(
		"http.server.request_count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create request counter: %v", err))
	}

	// failureCounter, err := meter.Int64Counter(
	// 	"http.server.request_failed_count",
	// 	metric.WithDescription("Total number of failed requests"),
	// 	metric.WithUnit("1"),
	// )
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to create failure counter: %v", err))
	// }
	var cfgTracerProvider = exporter.TracerExporter.TracerProvider
	tracer := cfgTracerProvider.Tracer(
		"gin",
	)
	var cfgPropagators = exporter.TracerExporter.TextMapPropagator
	return func(c *gin.Context) {
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
		hostname, _ := os.Hostname()
		// serve the request to the next middleware
		requestCounter.Add(ctx, 1, metric.WithAttributes(
			semconv2.HostName(c.Request.Host),
			semconv2.ServiceName(hostname),
		))
		c.Next()

		status := c.Writer.Status()
		spanStatus := codes.Code(status)
		span.SetStatus(spanStatus, spanStatus.String())
		span.SetAttributes(
			semconv2.HTTPResponseStatusCode(status),
		)
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}
	}
}
