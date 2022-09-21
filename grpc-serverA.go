package main

import (
	"context"
	_ "embed"
	"net/http"
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

	// register grpc
	grpcEntry := rkgrpc.GetGrpcEntry("grpcServerA")
	grpcEntry.AddRegFuncGrpc(registerServerA)
	//grpcEntry.AddRegFuncGw(greeter.RegisterServerAHandlerFromEndpoint)
	// register grpc
	//grpcEntry := rkgrpc.GetGrpcEntry("greeter")
	grpcEntry.AddRegFuncGrpc(registerMyServer)
	grpcEntry.AddRegFuncGw(greeter.RegisterMyServerHandlerFromEndpoint)

	// Bootstrap
	boot.Bootstrap(context.TODO())
	asynqServer := startAsynqWorker(grpcEntry.LoggerEntry.Logger, grpcBootA)
	boot.AddShutdownHookFunc("asynq worker", asynqServer.Stop)
	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}
func startAsynqWorker(logger *zap.Logger, grpcBootA []byte) *asynq.Server {
	// start asynq service
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "192.168.44.203:6379"},
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
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "192.168.44.203:6379"})
	defer client.Close()

	// get trace metadata
	header := http.Header{}
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, propagation.HeaderCarrier(header))

	err := server.enqueueTask(client, header)

	return &greeter.TaskResp{}, err
}

func (server *Service) enqueueTask(client *asynq.Client, header http.Header) error {
	task, err := task.NewDemoTask(header)
	_, err = client.Enqueue(task)
	return err
}

func registerServerA(server *grpc.Server) {
	greeter.RegisterServerAServer(server, &Service{})
}

func CallFuncAA(ctx context.Context) {
	_, span := rkasynq.NewSpan(ctx, "funcAA")
	defer rkasynq.EndSpan(span, true)

	time.Sleep(10 * time.Millisecond)
}

func (server *Service) Login(ctx context.Context, req *greeter.LoginReq) (*greeter.LoginResp, error) {
	md := metadata.Pairs()
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
	ctx = metadata.NewOutgoingContext(ctx, md)
	//server.CallA(ctx, &greeter.CallAReq{})
	// ctx, span := rkasynq.NewSpan(ctx, "Login")
	// defer rkasynq.EndSpan(span, true)

	//time.Sleep(10 * time.Millisecond)
	CallFuncAA(ctx)
	return &greeter.LoginResp{}, nil
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
	// eject span context from gin context and inject into grpc ctx
	//grpcCtx := trace.ContextWithRemoteSpanContext(context.Background(), rkginctx.GetTraceSpan(ctx).SpanContext())
	grpcCtx := ctx
	md := metadata.Pairs()
	rkgrpcctx.GetTracerPropagator(ctx).Inject(grpcCtx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
	grpcCtx = metadata.NewOutgoingContext(grpcCtx, md)
	// call gRPC server B
	clientB.CallB(grpcCtx, &greeter.CallBReq{})
	server.Enqueue(ctx, &greeter.TaskReq{})
	return &greeter.CallAResp{}, nil
}
