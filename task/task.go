package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"

	"github.com/hibiken/asynq"
	greeter "github.com/rookie-ninja/rk-demo/api/gen/v1"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
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
	md := metadata.Pairs()
	pg := rkasynq.GetPropagator(ctx)
	pg.Inject(ctx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})

	ctx = metadata.NewOutgoingContext(ctx, md)

	clientA.Login(ctx, &greeter.LoginReq{})
}

func clientA_CallA(ctx context.Context) {
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
	md := metadata.Pairs()
	pg := rkasynq.GetPropagator(ctx)
	pg.Inject(ctx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})

	ctx = metadata.NewOutgoingContext(ctx, md)

	clientA.CallA(ctx, &greeter.CallAReq{})
}

func HandleDemoTask(ctx context.Context, t *asynq.Task) error {
	// sleep a while for testing
	time.Sleep(50 * time.Millisecond)

	var p DemoPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	rkentry.GlobalAppCtx.GetLoggerEntryDefault().Info("handle demo task", zap.String("traceId", rkasynq.GetTraceId(ctx)))

	// call func A & B
	CallWorkerA(ctx)
	CallWorkerB(ctx)

	// call clientA Login
	clientA_Login(ctx)

	// call clientA CallA
	clientA_CallA(ctx)
	return fmt.Errorf("HandleDemoTask err")
	return nil
}

func CallWorkerA(ctx context.Context) {
	newCtx, span := rkasynq.NewSpan(ctx, "workerA")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)

	CallWorkerAA(newCtx)
}

func CallWorkerAA(ctx context.Context) {
	_, span := rkasynq.NewSpan(ctx, "workerA-A")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)
}

func CallWorkerB(ctx context.Context) {
	_, span := rkasynq.NewSpan(ctx, "workerB")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)
}
