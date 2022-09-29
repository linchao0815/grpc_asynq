// Code generated by protoc-gen-go-asynqgen. DO NOT EDIT.
// versions:
// protoc-gen-go-asynqgen v1.0.11

package demo

import (
	context "context"
	json "encoding/json"
	asynq "github.com/hibiken/asynq"
	proto "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

import (
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-logger"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"google.golang.org/grpc/metadata"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	otelcodes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"path"
	"fmt"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/attribute"
	"github.com/rookie-ninja/rk-grpc/v2/middleware/context"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the asynq package it is being compiled against.
var _ = new(context.Context)
var _ = new(asynq.Task)
var _ = new(emptypb.Empty)
var _ = new(proto.Message)
var _ = new(json.InvalidUTF8Error)

var (
	noopTracerProvider = oteltrace.NewNoopTracerProvider()
	tHolder            = newTraceHolder(nil)
)

const (
	spanKey       = "SpanKey"
	traceIdKey    = "TraceIdKey"
	tracerKey     = "TracerKey"
	propagatorKey = "PropagatorKey"
	providerKey   = "ProviderKey"
)

// trace
func RegisterTraceHolder(conf *TraceConfig) {
	tHolder = newTraceHolder(conf)
}

type TraceConfig struct {
	Asynq struct {
		Trace struct {
			Enabled        bool
			ServiceName    string
			ServiceVersion string
			Exporter       struct {
				File struct {
					Enabled    bool
					OutputPath string
				}
				Jaeger struct {
					Agent struct {
						Enabled bool
						Host    string
						Port    int
					}
					Collector struct {
						Enabled  bool
						Endpoint string
						Username string
						Password string
					}
				}
			}
		}
	}
}

func newTraceHolder(conf *TraceConfig) *traceHolder {
	if conf == nil {
		conf = &TraceConfig{}
		conf.Asynq.Trace.Enabled = true
	}

	mid := &traceHolder{}

	opts := toOptions(conf)

	for i := range opts {
		opts[i](mid)
	}

	if mid.exporter == nil {
		mid.exporter = newNoopExporter()
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

	return mid
}

type traceHolder struct {
	exporter   sdktrace.SpanExporter
	processor  sdktrace.SpanProcessor
	provider   *sdktrace.TracerProvider
	propagator propagation.TextMapPropagator
	tracer     oteltrace.Tracer
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
	if v := ctx.Value(traceIdKey); v != nil {
		if res, ok := v.(string); ok {
			return res
		}
	}
	return ""
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

func InjectSpanToHttpReq(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}

	newCtx := oteltrace.ContextWithRemoteSpanContext(req.Context(), GetSpan(ctx).SpanContext())
	GetPropagator(ctx).Inject(newCtx, propagation.HeaderCarrier(req.Header))
}

func InjectSpanToNewContext(ctx context.Context) context.Context {
	newCtx := oteltrace.ContextWithRemoteSpanContext(ctx, GetSpan(ctx).SpanContext())
	md := metadata.Pairs()
	GetPropagator(ctx).Inject(newCtx, &grpcMetadataCarrier{Md: &md})
	newCtx = metadata.NewOutgoingContext(newCtx, md)

	return newCtx
}

// toOptions convert BootConfig into Option list
func toOptions(config *TraceConfig) []option {
	opts := make([]option, 0)

	if config.Asynq.Trace.Enabled {
		var exporter sdktrace.SpanExporter

		if config.Asynq.Trace.Exporter.File.Enabled {
			exporter = newFileExporter(config.Asynq.Trace.Exporter.File.OutputPath)
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

			exporter = newJaegerExporter(jaeger.WithAgentEndpoint(opts...))
		}

		if config.Asynq.Trace.Exporter.Jaeger.Collector.Enabled {
			opts := []jaeger.CollectorEndpointOption{
				jaeger.WithUsername(config.Asynq.Trace.Exporter.Jaeger.Collector.Username),
				jaeger.WithPassword(config.Asynq.Trace.Exporter.Jaeger.Collector.Password),
			}

			if len(config.Asynq.Trace.Exporter.Jaeger.Collector.Endpoint) > 0 {
				opts = append(opts, jaeger.WithEndpoint(config.Asynq.Trace.Exporter.Jaeger.Collector.Endpoint))
			}

			exporter = newJaegerExporter(jaeger.WithCollectorEndpoint(opts...))
		}

		opts = append(opts, withExporter(exporter))
	}

	return opts
}

// Option is used while creating middleware as param
type option func(*traceHolder)

// WithExporter Provide sdktrace.SpanExporter.
func withExporter(exporter sdktrace.SpanExporter) option {
	return func(opt *traceHolder) {
		if exporter != nil {
			opt.exporter = exporter
		}
	}
}

// ***************** Global *****************

// NoopExporter noop
type noopExporter struct{}

// ExportSpans handles export of SpanSnapshots by dropping them.
func (nsb *noopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }

// Shutdown stops the exporter by doing nothing.
func (nsb *noopExporter) Shutdown(context.Context) error { return nil }

// NewNoopExporter create a noop exporter
func newNoopExporter() sdktrace.SpanExporter {
	return &noopExporter{}
}

// NewFileExporter create a file exporter whose default output is stdout.
func newFileExporter(outputPath string, opts ...stdouttrace.Option) sdktrace.SpanExporter {
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
func newJaegerExporter(opt jaeger.EndpointOption) sdktrace.SpanExporter {
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

type grpcMetadataCarrier struct {
	Md *metadata.MD
}

// Get value with key from grpc metadata.
func (carrier *grpcMetadataCarrier) Get(key string) string {
	values := carrier.Md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set value with key into grpc metadata.
func (carrier *grpcMetadataCarrier) Set(key string, value string) {
	carrier.Md.Set(key, value)
}

// Keys List keys in grpc metadata.
func (carrier *grpcMetadataCarrier) Keys() []string {
	out := make([]string, 0, len(*carrier.Md))
	for key := range *carrier.Md {
		out = append(out, key)
	}
	return out
}

func _handle_task_before(ctx context.Context, task *asynq.Task, in interface{}) (context.Context, oteltrace.Span, error) {
	wrap := &wrapPayload{
		Payload: in,
		Trace:   map[string]string{},
	}
	if err := json.Unmarshal(task.Payload(), &wrap); err != nil {
		return ctx, nil, fmt.Errorf("%s req=%s err=%s", task.Type(), wrap, err)
	}

	ctx = tHolder.propagator.Extract(ctx, propagation.MapCarrier(wrap.Trace))
	spanCtx := oteltrace.SpanContextFromContext(ctx)

	// create new span
	ctx, span := tHolder.tracer.Start(oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx), task.Type())
	defer span.End()

	ctx = context.WithValue(ctx, spanKey, span)
	ctx = context.WithValue(ctx, traceIdKey, span.SpanContext().TraceID().String())
	ctx = context.WithValue(ctx, tracerKey, tHolder.tracer)
	ctx = context.WithValue(ctx, propagatorKey, tHolder.propagator)
	ctx = context.WithValue(ctx, providerKey, tHolder.provider)

	md := metadata.Pairs()
	tHolder.propagator.Inject(ctx, &grpcMetadataCarrier{Md: &md})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx, span, nil
}

func _handle_task_after(span oteltrace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("%v", err))
	} else {
		span.SetStatus(codes.Ok, "success")
	}
}

type wrapPayload struct {
	Trace   map[string]string `json:"trace"`
	Payload interface{}       `json:",inline"`
}
type UserTaskServer interface {
	CreateUser(context.Context, *CreateUserPayload) error
	UpdateUser(context.Context, *UpdateUserPayload) error
}

func RegisterUserTaskServer(mux *asynq.ServeMux, srv UserTaskServer) {
	mux.HandleFunc("user:create", _User_CreateUser_Task_Handler(srv))
	mux.HandleFunc("user:update", _User_UpdateUser_Task_Handler(srv))
}

func _User_CreateUser_Task_Handler(srv UserTaskServer) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		in := &CreateUserPayload{}

		ctx, span, err := _handle_task_before(ctx, task, in)
		if err != nil {
			return err
		}

		err = srv.CreateUser(ctx, in)

		_handle_task_after(span, err)

		return err
	}
}

func _User_UpdateUser_Task_Handler(srv UserTaskServer) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		in := &UpdateUserPayload{}

		ctx, span, err := _handle_task_before(ctx, task, in)
		if err != nil {
			return err
		}

		err = srv.UpdateUser(ctx, in)

		_handle_task_after(span, err)

		return err
	}
}

type UserTaskClient interface {
	CreateUser(ctx context.Context, req *CreateUserPayload, opts ...asynq.Option) (info *asynq.TaskInfo, span oteltrace.Span, err error)
	UpdateUser(ctx context.Context, req *UpdateUserPayload, opts ...asynq.Option) (info *asynq.TaskInfo, span oteltrace.Span, err error)
}

type UserTaskClientImpl struct {
	cc *asynq.Client
}

func NewUserTaskClient(client *asynq.Client) UserTaskClient {
	return &UserTaskClientImpl{client}
}

func (c *UserTaskClientImpl) CreateUser(ctx context.Context, in *CreateUserPayload, opts ...asynq.Option) (*asynq.TaskInfo, oteltrace.Span, error) {
	if rkgrpcctx.GetTracerPropagator(ctx) != nil {
		ctx = rkgrpcctx.InjectSpanToNewContext(ctx)
	}

	spanCtx := oteltrace.SpanContextFromContext(ctx)
	ctx, span := tHolder.tracer.Start(oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx), "CreateUserClient")
	defer span.End()

	// get trace metadata
	m := make(map[string]string)
	tHolder.propagator.Inject(ctx, propagation.MapCarrier(m))

	wrap, err := json.Marshal(wrapPayload{
		Trace:   m,
		Payload: in,
	})
	if err != nil {
		return nil, nil, err
	}

	task := asynq.NewTask("user:create", wrap, opts...)

	info, err := c.cc.Enqueue(task)
	if err != nil {
		return nil, nil, err
	}
	return info, span, nil
}

func (c *UserTaskClientImpl) UpdateUser(ctx context.Context, in *UpdateUserPayload, opts ...asynq.Option) (*asynq.TaskInfo, oteltrace.Span, error) {
	if rkgrpcctx.GetTracerPropagator(ctx) != nil {
		ctx = rkgrpcctx.InjectSpanToNewContext(ctx)
	}

	spanCtx := oteltrace.SpanContextFromContext(ctx)
	ctx, span := tHolder.tracer.Start(oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx), "UpdateUserClient")
	defer span.End()

	// get trace metadata
	m := make(map[string]string)
	tHolder.propagator.Inject(ctx, propagation.MapCarrier(m))

	wrap, err := json.Marshal(wrapPayload{
		Trace:   m,
		Payload: in,
	})
	if err != nil {
		return nil, nil, err
	}

	task := asynq.NewTask("user:update", wrap, opts...)

	info, err := c.cc.Enqueue(task)
	if err != nil {
		return nil, nil, err
	}
	return info, span, nil
}
