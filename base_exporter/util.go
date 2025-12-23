package base_exporter

import (
	"context"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type defaultSampler struct {
	baseSampler sdktrace.Sampler
	be          *TracerExporter
}

// 添加be参数，以便在ShouldSample中访问samplingCall数组
func newDefaultSampler(be *TracerExporter, baseRatio float64) sdktrace.Sampler {
	base := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(baseRatio),
		// 为根span添加特殊规则
		sdktrace.WithRemoteParentSampled(sdktrace.AlwaysSample()), // 远程父span已采样则采样
		sdktrace.WithLocalParentSampled(sdktrace.AlwaysSample()),  // 本地父span已采样则采样
	)
	return &defaultSampler{
		baseSampler: base,
		be:          be,
	}
}

func (ds *defaultSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	// 首先检查samplingCall数组中的函数
	if ds.be != nil && ds.be.samplingCall != nil {
		for _, fn := range ds.be.samplingCall {
			if fn == nil {
				continue
			}
			shouldSample, err := fn(p)
			if err != nil {
				continue
			}
			if shouldSample == nil {
				continue
			}
			return *shouldSample
		}
	}

	// 插件逻辑保持原样
	if ds.be != nil && ds.be.pluginManager != nil {
		if plugin, ok := ds.be.pluginManager.UnsafeGetPlugin("sampler"); ok {
			pluginResult, err := plugin.SyncEvents(context.Background(), p, time.Second*2)
			if err == nil {
				if b, ok := pluginResult.(sdktrace.SamplingResult); ok {
					return b
				}
			}
		}
	}

	// 如果没有则使用基础采样器
	return ds.baseSampler.ShouldSample(p)
}

func (ds *defaultSampler) Description() string {
	return "defaultSampler"
}

type defaultSpanProcessor struct {
	nextProcessor sdktrace.SpanProcessor
	be            *TracerExporter
}

func newDefaultSpanProcessor(be *TracerExporter, next sdktrace.SpanProcessor) *defaultSpanProcessor {
	return &defaultSpanProcessor{
		nextProcessor: next,
		be:            be,
	}
}

func (p *defaultSpanProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	// 传递给下一个处理器
	if p.nextProcessor != nil {
		p.nextProcessor.OnStart(ctx, s)
	}

}

func (p *defaultSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p.be != nil && p.be.spanProcessorEndCall != nil && len(p.be.spanProcessorEndCall) > 0 {
		for _, fn := range p.be.spanProcessorEndCall {
			if fn == nil {
				continue
			}
			fn(s)
		}
	}

	if p.be != nil && p.be.pluginManager != nil {
		skipCount := 0
		if plugin, ok := p.be.pluginManager.UnsafeGetPlugin("span_processor_end"); ok {
			plugin.SyncEvents(context.Background(), p, time.Second*5)
		}
		if skipCount >= 1 {
			return
		}
	}

	// 传递给下一个处理器
	if p.nextProcessor != nil {
		p.nextProcessor.OnEnd(s)
	}
}

func (p *defaultSpanProcessor) Shutdown(ctx context.Context) error {
	if p.nextProcessor != nil {
		return p.nextProcessor.Shutdown(ctx)
	}
	return nil
}

func (p *defaultSpanProcessor) ForceFlush(ctx context.Context) error {
	if p.nextProcessor != nil {
		return p.nextProcessor.ForceFlush(ctx)
	}
	return nil
}
