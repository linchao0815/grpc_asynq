package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	exporter, _ = jaeger.New(jaeger.WithAgentEndpoint())
	processor   = sdktrace.NewBatchSpanProcessor(exporter)
	propagator  = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})
	provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(processor),
		sdktrace.WithResource(
			sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				attribute.String("service.name", "httpclient"),
				attribute.String("service.version", "v1"),
			)),
	)
	tracer = provider.Tracer("demo", oteltrace.WithInstrumentationVersion(contrib.SemVersion()))
)

func main() {
	// 创建 root span
	ctx, span := tracer.Start(context.Background(), "root-send")

	fmt.Println(fmt.Sprintf("%s, traceId:%s, spanId:%s",
		"root-send",
		span.SpanContext().TraceID().String(),
		span.SpanContext().SpanID().String()))

	// enqueue 第一个 task
	//ctx = send("http://localhost:2008/v1/Enqueue", `{"Param":"first-send"}`, ctx, "first-send")
	ctx = send("http://localhost:2008/v1/GameTest", `{"Param":"first-send"}`, ctx, "first-send")

	// enqueue 第二个 task
	send("http://localhost:2002/v1/CallB", `{"Param":"second-send"}`, ctx, "second-send")

	// 不要在最上面用 defer span.End()，否则 jaeger 里看不到 root-send，因为 span.End() 之后来不及推送到 jaeger
	span.End()

	// 等待数据传到 jaeger
	time.Sleep(10 * time.Second)
}

func send(url, jsonStr string, ctx context.Context, spanName string) context.Context {
	fmt.Println("URL:>", url)

	// 创建 Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("Content-Type", "application/json")

	// 创建一个 Span，Inject 到 context 中
	carrier := propagation.HeaderCarrier(req.Header)
	ctx, span := tracer.Start(ctx, spanName)
	propagator.Inject(ctx, carrier)
	defer span.End()

	// 打印 traceId, spanId
	fmt.Println(fmt.Sprintf("%s, traceId:%s, spanId:%s",
		spanName,
		span.SpanContext().TraceID().String(),
		span.SpanContext().SpanID().String()))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(resp.Header))

	return ctx
}
