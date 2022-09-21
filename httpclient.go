package main

import (
	"bytes"
	"fmt"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"net/http"
	"time"
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
	send("http://localhost:2008/v1/enqueue", `{""}`)

	// wait to send trace to jaeger
	time.Sleep(10 * time.Second)
}

func send(url, jsonStr string) string {
	fmt.Println("URL:>", url)

	// 创建 Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("Content-Type", "application/json")

	// 创建一个 Span，Inject 到 context 中
	carrier := propagation.HeaderCarrier(req.Header)
	ctx, span := tracer.Start(req.Context(), "enqueue")
	propagator.Inject(ctx, carrier)
	defer span.End()

	// 打印 traceId
	fmt.Println(fmt.Sprintf("traceId:%s", span.SpanContext().TraceID().String()))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
