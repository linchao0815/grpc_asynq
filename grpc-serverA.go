package main

import (
	"context"
	_ "embed"
	"net/http"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	greeter "github.com/rookie-ninja/rk-demo/api/gen/v1"
	"github.com/rookie-ninja/rk-demo/task"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
	rkasynq "github.com/rookie-ninja/rk-repo/asynq"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//go:embed grpc-serverA.yaml
var grpcBootA []byte

type Service struct {
	greeter.UnimplementedMyServerServer
	greeter.UnimplementedServerAServer
}

func main() {
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(grpcBootA))

	// register grpc: Login & CallA
	grpcEntry := rkgrpc.GetGrpcEntry("grpcServerA")
	grpcEntry.AddRegFuncGrpc(registerServerA)

	// register grpc gateway: /v1/enqueue
	grpcEntry.AddRegFuncGrpc(registerMyServer)
	grpcEntry.AddRegFuncGw(greeter.RegisterMyServerHandlerFromEndpoint)

	// Bootstrap
	boot.Bootstrap(context.TODO())

	// start asynq server
	asynqServer := startAsynqWorker(grpcEntry.LoggerEntry.Logger)
	boot.AddShutdownHookFunc("asynq worker", asynqServer.Shutdown)

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}
func startAsynqWorker(logger *zap.Logger) *asynq.Server {
	// start asynq service
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{
			Logger: logger.Sugar(),
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.TypeDemo, task.HandleDemoTask)

	// add jaeger middleware
	jaegerMid, err := rkasynq.NewJaegerMid(grpcBootA)
	if err != nil {
		rkentry.ShutdownWithError(err)
	}
	mux.Use(jaegerMid)

	if err := srv.Start(mux); err != nil {
		rkentry.ShutdownWithError(err)
	}

	return srv
}

func registerMyServer(server *grpc.Server) {
	greeter.RegisterMyServerServer(server, &Service{})
}

func (server *Service) Enqueue(ctx context.Context, req *greeter.TaskReq) (*greeter.TaskResp, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer client.Close()

	// get trace metadata
	header := http.Header{}
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, propagation.HeaderCarrier(header))

	// enqueue task
	task, _ := task.NewDemoTask(header)
	client.Enqueue(task)

	// 把 Trace 信息，存入 Metadata，以 Header 的形式返回给 httpclient
	for k, v := range header {
		rkgrpcctx.AddHeaderToClient(ctx, k, strings.Join(v, ","))
	}

	return &greeter.TaskResp{}, nil
}

func registerServerA(server *grpc.Server) {
	greeter.RegisterServerAServer(server, &Service{})
}

func (server *Service) Login(ctx context.Context, req *greeter.LoginReq) (*greeter.LoginResp, error) {
	md := metadata.Pairs()
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
	ctx = metadata.NewOutgoingContext(ctx, md)

	server.LoginAfter(ctx)
	return &greeter.LoginResp{}, nil
}

func (server *Service) LoginAfter(ctx context.Context) {
	span := rkgrpcctx.NewTraceSpan(ctx, "LoginAfter")
	defer rkgrpcctx.EndTraceSpan(ctx, span, true)

	time.Sleep(10 * time.Millisecond)
}

func (server *Service) CallA(ctx context.Context, req *greeter.CallAReq) (*greeter.CallAResp, error) {
	// create grpc client
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}
	// create connection with grpc-serverB
	connB, _ := grpc.Dial("localhost:2022", opts...)
	defer connB.Close()
	clientB := greeter.NewServerBClient(connB)

	// 把 trace info 载入到 context 中
	ctx = rkgrpcctx.InjectSpanToNewContext(ctx)
	clientB.CallB(ctx, &greeter.CallBReq{})

	return &greeter.CallAResp{}, nil
}
