package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	greeter "github.com/rookie-ninja/rk-demo/api/gen/v1"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
	rkasynq "github.com/rookie-ninja/rk-repo/asynq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ************ Task ************

const (
	TypeDemo = "demo-task"
)

type DemoPayload struct {
	TraceHeader http.Header `json:"traceHeader"`
}

func NewDemoTask(header http.Header) (*asynq.Task, error) {
	payload, err := json.Marshal(DemoPayload{
		TraceHeader: header,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDemo, payload), nil
}

func clientA_Login(ctx context.Context) {
	// create grpc client
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}
	// create connection with grpc-serverA
	connA, _ := grpc.Dial("localhost:2008", opts...)
	defer connA.Close()
	clientA := greeter.NewServerAClient(connA)
	// eject span context from gin context and inject into grpc ctx
	//grpcCtx := trace.ContextWithRemoteSpanContext(context.Background(), rkginctx.GetTraceSpan(ctx).SpanContext())
	md := metadata.Pairs()
	pg := rkgrpcctx.GetTracerPropagator(ctx)
	if pg != nil {
		pg.Inject(ctx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
	} else {
		fmt.Printf("rkgrpcctx.GetTracerPropagator() =nil\n")
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	// call gRPC server B
	clientA.Login(ctx, &greeter.LoginReq{})
}

func HandleDemoTask(ctx context.Context, t *asynq.Task) error {
	// sleep a while for testing
	time.Sleep(50 * time.Millisecond)

	var p DemoPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	rkentry.GlobalAppCtx.GetLoggerEntryDefault().Info("handle demo task", zap.String("traceId", rkasynq.GetTraceId(ctx)))

	CallFuncA(ctx)
	CallFuncB(ctx)
	clientA_Login(ctx)

	return nil
}

func CallFuncA(ctx context.Context) {
	newCtx, span := rkasynq.NewSpan(ctx, "funcA")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)

	CallFuncAA(newCtx)
}

func CallFuncAA(ctx context.Context) {
	_, span := rkasynq.NewSpan(ctx, "funcAA")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)
}

func CallFuncB(ctx context.Context) {
	_, span := rkasynq.NewSpan(ctx, "funcB")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)
}
