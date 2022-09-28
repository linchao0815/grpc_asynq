package myasynq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/hibiken/asynq"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rklogger "github.com/rookie-ninja/rk-logger"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v3"
)

type TaskPaylod struct {
	//RequestId   string
	//TraceId     string
	TraceHeader http.Header `json:"traceHeader"`
	In          interface{} `json:"in"`
}

func ToMarshal(obj interface{}) (str string) {
	res, err := json.Marshal(obj)
	if err != nil {
		str = fmt.Sprint(obj)
	} else {
		str = string(res)
	}
	return str
}

var (
	noopTracerProvider = oteltrace.NewNoopTracerProvider()
)

const (
	spanKey       = "SpanKey"
	traceIdKey    = "TraceIdKey"
	tracerKey     = "TracerKey"
	propagatorKey = "PropagatorKey"
	providerKey   = "ProviderKey"
)

type TraceConfig struct {
	Asynq struct {
		Trace struct {
			Enabled        bool   `yaml:"enabled" json:"enabled"`
			ServiceName    string `yaml:"serviceName"`
			ServiceVersion string `yaml:"serviceVersion"`
			Exporter       struct {
				File struct {
					Enabled    bool   `yaml:"enabled" json:"enabled"`
					OutputPath string `yaml:"outputPath" json:"outputPath"`
				} `yaml:"file" json:"file"`
				Jaeger struct {
					Agent struct {
						Enabled bool   `yaml:"enabled" json:"enabled"`
						Host    string `yaml:"host" json:"host"`
						Port    int    `yaml:"port" json:"port"`
					} `yaml:"agent" json:"agent"`
					Collector struct {
						Enabled  bool   `yaml:"enabled" json:"enabled"`
						Endpoint string `yaml:"endpoint" json:"endpoint"`
						Username string `yaml:"username" json:"username"`
						Password string `yaml:"password" json:"password"`
					} `yaml:"collector" json:"collector"`
				} `yaml:"jaeger" json:"jaeger"`
			} `yaml:"exporter" json:"exporter"`
		} `yaml:"trace"`
	} `yaml:"asynq"`
}

func NewJaegerMid(traceRaw []byte) (asynq.MiddlewareFunc, error) {
	conf := &TraceConfig{}
	err := yaml.Unmarshal(traceRaw, conf)

	if err != nil {
		return nil, err
	}

	mid := &TraceMiddleware{}

	opts := ToOptions(conf)

	for i := range opts {
		opts[i](mid)
	}

	if mid.exporter == nil {
		mid.exporter = NewNoopExporter()
	}

	if mid.processor == nil {
		mid.processor = sdktrace.NewBatchSpanProcessor(mid.exporter)
	}

	if mid.provider == nil {
		mid.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(mid.processor),
			sdktrace.WithResource(
				sdkresource.NewWithAttributes(
					semconv.SchemaURL,
					attribute.String("service.name", conf.Asynq.Trace.ServiceName),
					attribute.String("service.version", conf.Asynq.Trace.ServiceVersion),
				)),
		)
	}

	mid.tracer = mid.provider.Tracer(conf.Asynq.Trace.ServiceName, oteltrace.WithInstrumentationVersion(contrib.SemVersion()))

	if mid.propagator == nil {
		mid.propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{})
	}

	return mid.Middleware, nil
}

type TraceMiddleware struct {
	exporter   sdktrace.SpanExporter
	processor  sdktrace.SpanProcessor
	provider   *sdktrace.TracerProvider
	propagator propagation.TextMapPropagator
	tracer     oteltrace.Tracer
}

func (m *TraceMiddleware) Middleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		var p TaskPaylod
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		ctx = m.propagator.Extract(ctx, propagation.HeaderCarrier(p.TraceHeader))
		spanCtx := oteltrace.SpanContextFromContext(ctx)

		// create new span
		ctx, span := m.tracer.Start(oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx), t.Type())
		defer span.End()

		ctx = context.WithValue(ctx, spanKey, span)
		ctx = context.WithValue(ctx, traceIdKey, span.SpanContext().TraceID())
		ctx = context.WithValue(ctx, tracerKey, m.tracer)
		ctx = context.WithValue(ctx, propagatorKey, m.propagator)
		ctx = context.WithValue(ctx, providerKey, m.provider)

		err := h.ProcessTask(ctx, t)

		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%v", err))
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		return err
	})
}

func GetSpan(ctx context.Context) oteltrace.Span {
	if v := ctx.Value(spanKey); v != nil {
		if res, ok := v.(oteltrace.Span); ok {
			return res
		}
	}

	_, span := noopTracerProvider.Tracer("rk-trace-noop").Start(ctx, "noop-span")
	return span
}

func GetTraceId(ctx context.Context) string {
	return GetSpan(ctx).SpanContext().TraceID().String()
}

func GetTracer(ctx context.Context) oteltrace.Tracer {
	if v := ctx.Value(tracerKey); v != nil {
		if res, ok := v.(oteltrace.Tracer); ok {
			return res
		}
	}

	return noopTracerProvider.Tracer("rk-trace-noop")
}

func GetPropagator(ctx context.Context) propagation.TextMapPropagator {
	if v := ctx.Value(propagatorKey); v != nil {
		if res, ok := v.(propagation.TextMapPropagator); ok {
			return res
		}
	}

	return nil
}

func GetProvider(ctx context.Context) *sdktrace.TracerProvider {
	if v := ctx.Value(providerKey); v != nil {
		if res, ok := v.(*sdktrace.TracerProvider); ok {
			return res
		}
	}

	return nil
}

func NewSpan(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	return GetTracer(ctx).Start(ctx, name)
}

func EndSpan(span oteltrace.Span, success bool) {
	if success {
		span.SetStatus(otelcodes.Ok, otelcodes.Ok.String())
	}

	span.End()
}

// ToOptions convert BootConfig into Option list
func ToOptions(config *TraceConfig) []Option {
	opts := make([]Option, 0)

	if config.Asynq.Trace.Enabled {
		var exporter sdktrace.SpanExporter

		if config.Asynq.Trace.Exporter.File.Enabled {
			exporter = NewFileExporter(config.Asynq.Trace.Exporter.File.OutputPath)
		}

		if config.Asynq.Trace.Exporter.Jaeger.Agent.Enabled {
			opts := make([]jaeger.AgentEndpointOption, 0)
			if len(config.Asynq.Trace.Exporter.Jaeger.Agent.Host) > 0 {
				opts = append(opts,
					jaeger.WithAgentHost(config.Asynq.Trace.Exporter.Jaeger.Agent.Host))
			}
			if config.Asynq.Trace.Exporter.Jaeger.Agent.Port > 0 {
				opts = append(opts, jaeger.WithAgentPort(fmt.Sprintf("%d", config.Asynq.Trace.Exporter.Jaeger.Agent.Port)))
			}

			exporter = NewJaegerExporter(jaeger.WithAgentEndpoint(opts...))
		}

		if config.Asynq.Trace.Exporter.Jaeger.Collector.Enabled {
			opts := []jaeger.CollectorEndpointOption{
				jaeger.WithUsername(config.Asynq.Trace.Exporter.Jaeger.Collector.Username),
				jaeger.WithPassword(config.Asynq.Trace.Exporter.Jaeger.Collector.Password),
			}

			if len(config.Asynq.Trace.Exporter.Jaeger.Collector.Endpoint) > 0 {
				opts = append(opts, jaeger.WithEndpoint(config.Asynq.Trace.Exporter.Jaeger.Collector.Endpoint))
			}

			exporter = NewJaegerExporter(jaeger.WithCollectorEndpoint(opts...))
		}

		opts = append(opts, WithExporter(exporter))
	}

	return opts
}

// Option is used while creating middleware as param
type Option func(*TraceMiddleware)

// WithExporter Provide sdktrace.SpanExporter.
func WithExporter(exporter sdktrace.SpanExporter) Option {
	return func(opt *TraceMiddleware) {
		if exporter != nil {
			opt.exporter = exporter
		}
	}
}

// ***************** Global *****************

// NoopExporter noop
type NoopExporter struct{}

// ExportSpans handles export of SpanSnapshots by dropping them.
func (nsb *NoopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }

// Shutdown stops the exporter by doing nothing.
func (nsb *NoopExporter) Shutdown(context.Context) error { return nil }

// NewNoopExporter create a noop exporter
func NewNoopExporter() sdktrace.SpanExporter {
	return &NoopExporter{}
}

// NewFileExporter create a file exporter whose default output is stdout.
func NewFileExporter(outputPath string, opts ...stdouttrace.Option) sdktrace.SpanExporter {
	if opts == nil {
		opts = make([]stdouttrace.Option, 0)
	}

	if outputPath == "" {
		outputPath = "stdout"
	}

	if outputPath == "stdout" {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	} else {
		// init lumberjack logger
		writer := rklogger.NewLumberjackConfigDefault()
		if !path.IsAbs(outputPath) {
			wd, _ := os.Getwd()
			outputPath = path.Join(wd, outputPath)
		}

		writer.Filename = outputPath

		opts = append(opts, stdouttrace.WithWriter(writer))
	}

	exporter, _ := stdouttrace.New(opts...)

	return exporter
}

// NewJaegerExporter Create jaeger exporter with bellow condition.
//
// 1: If no option provided, then export to jaeger agent at localhost:6831
// 2: Jaeger agent
//    If no jaeger agent host was provided, then use localhost
//    If no jaeger agent port was provided, then use 6831
// 3: Jaeger collector
//    If no jaeger collector endpoint was provided, then use http://localhost:14268/api/traces
func NewJaegerExporter(opt jaeger.EndpointOption) sdktrace.SpanExporter {
	// Assign default jaeger agent endpoint which is localhost:6831
	if opt == nil {
		opt = jaeger.WithAgentEndpoint()
	}

	exporter, err := jaeger.New(opt)

	if err != nil {
		rkentry.ShutdownWithError(err)
	}

	return exporter
}
