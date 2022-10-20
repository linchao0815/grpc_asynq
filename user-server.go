package main

import (
	"context"
	_ "embed"
	"fmt"
	"myasynq"

	"github.com/hibiken/asynq"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	demo "github.com/rookie-ninja/rk-demo/api/gen/v1"
	"github.com/rookie-ninja/rk-demo/grpcInterceptor"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed user-server.yaml
var userServerConf []byte

func main() {
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(userServerConf))

	// register user server
	grpcEntry := rkgrpc.GetGrpcEntry("userServer")
	grpcEntry.AddRegFuncGrpc(func(server *grpc.Server) {
		demo.RegisterUserServer(server, &userServer{
			client: demo.NewUserTaskClient(asynq.NewClient(asynq.RedisClientOpt{Addr: "192.168.44.203:6379", DB: 10})),
		})
	})
	grpcEntry.AddRegFuncGw(demo.RegisterUserHandlerFromEndpoint)
	grpcEntry.UnaryInterceptors = append(grpcEntry.UnaryInterceptors, grpcInterceptor.UnaryInterceptor)

	// register and start user task server
	taskServer := NewUserTaskServer()
	demo.RegisterUserTaskServer(taskServer.mux, taskServer)

	// Bootstrap
	boot.Bootstrap(context.TODO())

	boot.AddShutdownHookFunc("asynq task server", taskServer.server.Shutdown)

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}

func NewUserTaskServer() *userTaskServer {
	// register trace holder in asynq
	conf := &myasynq.TraceConfig{}
	conf.Asynq.Trace.Enabled = true
	conf.Asynq.Trace.ServiceName = "task"
	conf.Asynq.Trace.ServiceVersion = "v1"
	conf.Asynq.Trace.Exporter.Jaeger.Collector.Enabled = true
	conf.Asynq.Trace.Exporter.Jaeger.Collector.Endpoint = "http://jaeger-query-nfttest.towergame.com:8087/api/traces"
	myasynq.RegisterTraceHolder(conf)

	// start asynq service
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "192.168.44.203:6379", DB: 10},
		asynq.Config{
			Logger: rkentry.GlobalAppCtx.GetLoggerEntryDefault().Sugar(),
		},
	)

	mux := asynq.NewServeMux()

	res := &userTaskServer{
		server: srv,
		mux:    mux,
	}

	if err := srv.Start(mux); err != nil {
		rkentry.ShutdownWithError(err)
	}

	return res
}

type userTaskServer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func (t *userTaskServer) CreateUser(ctx context.Context, payload *demo.CreateUserPayload) error {
	fmt.Println(fmt.Sprintf("traceId: %s", myasynq.GetTraceId(ctx)))

	// call after server
	// after, _ := grpc.Dial("localhost:2022", []grpc.DialOption{
	// 	grpc.WithBlock(),
	// 	grpc.WithInsecure(),
	// }...)
	after, _ := grpc.DialContext(ctx, "localhost:2022", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(grpcInterceptor.ClientInterceptor()))
	defer after.Close()

	client := demo.NewAfterClient(after)

	_, err := client.CreateAfter(ctx, &demo.CreateAfterReq{})

	return err
}

func (t *userTaskServer) UpdateUser(ctx context.Context, payload *demo.UpdateUserPayload) error {
	// after, _ := grpc.Dial("localhost:2022", []grpc.DialOption{
	// 	grpc.WithBlock(),
	// 	grpc.WithInsecure(),
	// }...)
	after, _ := grpc.DialContext(ctx, "localhost:2022", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(grpcInterceptor.ClientInterceptor()))
	defer after.Close()

	client := demo.NewAfterClient(after)

	_, err := client.UpdateAfter(myasynq.InjectSpanToNewContext(ctx), &demo.UpdateAfterReq{})

	return err
}

type userServer struct {
	demo.UnimplementedUserServer
	client demo.UserTaskClient
}

func (s *userServer) CreateUser(ctx context.Context, payload *demo.CreateUserPayload) (*demo.Response, error) {
	_, span, err := s.client.CreateUser(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &demo.Response{
		TraceId: span.SpanContext().TraceID().String(),
	}, nil
}

func (s *userServer) UpdateUser(ctx context.Context, payload *demo.UpdateUserPayload) (*demo.Response, error) {
	_, span, err := s.client.UpdateUser(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &demo.Response{
		TraceId: span.SpanContext().TraceID().String(),
	}, nil
}

func (s *userServer) HealthCheck(ctx context.Context, payload *demo.Empty) (*demo.Empty, error) {
	return &demo.Empty{}, nil
}
