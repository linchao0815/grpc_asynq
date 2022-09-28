package main

import (
	"context"
	_ "embed"

	"encoding/json"
	"fmt"
	"myasynq"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	pb "github.com/rookie-ninja/rk-demo/api/gen/v1"
	"github.com/rookie-ninja/rk-demo/grpcInterceptor"
	"github.com/rookie-ninja/rk-demo/task"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

//go:embed grpc-serverA.yaml
var grpcBootA []byte

type Service struct {
	// pb.UnimplementedMyServerServer
	pb.UnimplementedServerAServer
}

var redisAddr = "192.168.44.203:6379"
var redis_DbId = 10

func main() {
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(grpcBootA))
	if len(redisAddr) == 0 {
		redisAddr = os.Getenv("redis")
		if len(redisAddr) == 0 {
			panic("please setup env variable for redis")
		}
	}
	// register grpc: Login & CallA
	grpcEntry := rkgrpc.GetGrpcEntry("grpcServerA")
	grpcEntry.AddRegFuncGrpc(registerServerA)
	grpcEntry.AddRegFuncGw(pb.RegisterServerAHandlerFromEndpoint)
	grpcEntry.UnaryInterceptors = append(grpcEntry.UnaryInterceptors, grpcInterceptor.UnaryInterceptor)
	// register grpc gateway: /v1/enqueue
	// grpcEntry.AddRegFuncGrpc(registerMyServer)
	// grpcEntry.AddRegFuncGw(pb.RegisterMyServerHandlerFromEndpoint)

	// Bootstrap
	boot.Bootstrap(context.TODO())

	// start asynq server
	asynqServer := startAsynqWorker(grpcEntry.LoggerEntry.Logger)
	boot.AddShutdownHookFunc("asynq worker", asynqServer.Shutdown)

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}

type ServerA_TaskJobServer struct {
	pb.ServerA_TaskJobServer
}

type ServerB_CallBackJobServer struct {
	pb.ServerB_CallBackJobServer
}

func startAsynqWorker(logger *zap.Logger) *asynq.Server {
	// start asynq service
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr, DB: redis_DbId},
		asynq.Config{
			Logger: logger.Sugar(),
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.TypeDemo, task.HandleDemoTask)
	pb.RegisterServerA_TaskJobServer(mux, &ServerA_TaskJobServer{})
	pb.RegisterServerB_CallBackJobServer(mux, &ServerB_CallBackJobServer{})

	// add jaeger middleware
	jaegerMid, err := myasynq.NewJaegerMid(grpcBootA)
	if err != nil {
		rkentry.ShutdownWithError(err)
	}
	mux.Use(jaegerMid)

	if err := srv.Start(mux); err != nil {
		rkentry.ShutdownWithError(err)
	}

	return srv
}

// func registerMyServer(server *grpc.Server) {
// 	pb.RegisterMyServerServer(server, &Service{})
// }
func (server *Service) Enqueue(ctx context.Context, req *pb.TaskReq) (*pb.TaskResp, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, DB: redis_DbId})
	defer client.Close()

	// get trace metadata
	header := http.Header{}
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, propagation.HeaderCarrier(header))

	// enqueue task
	//task, _ := task.NewDemoTask(header, req)
	payload, err := json.Marshal(myasynq.TaskPaylod{
		In:          req,
		TraceHeader: header,
	})
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(task.TypeDemo, payload)
	client.Enqueue(task)

	// 把 Trace 信息，存入 Metadata，以 Header 的形式返回给 httpclient
	for k, v := range header {
		rkgrpcctx.AddHeaderToClient(ctx, k, strings.Join(v, ","))
	}

	return &pb.TaskResp{}, nil
}

func (server *Service) GameTest(ctx context.Context, in *pb.GameTestReq) (*pb.GameTestResp, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, DB: redis_DbId})
	defer client.Close()
	_, err := (pb.NewServerA_TaskJobClient(client)).GameTest_Task(ctx, &pb.GameTest_TaskReq{Param: "test"})
	if err != nil {
		return &pb.GameTestResp{Param: in.Param}, fmt.Errorf("GameTest_Task err:%s", err)
	}

	return &pb.GameTestResp{Param: in.Param}, nil
}

func (s *ServerB_CallBackJobServer) CallB_CallBack(ctx context.Context, in *pb.CallBReply) error {
	task.ClientA_Login(ctx)
	return nil
}

func (s *ServerA_TaskJobServer) GameTest_Task(ctx context.Context, in *pb.GameTest_TaskReq) (err error) {
	connB, _ := grpc.DialContext(ctx, "localhost:2022", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(grpcInterceptor.ClientInterceptor()))
	defer connB.Close()
	clientB := pb.NewServerBClient(connB)
	clientB.CallB(ctx, &pb.CallBReq{})

	return nil
}
func registerServerA(server *grpc.Server) {
	pb.RegisterServerAServer(server, &Service{})
}

func (server *Service) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	md := metadata.Pairs()
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
	ctx = metadata.NewOutgoingContext(ctx, md)

	server.LoginAfter(ctx)
	return &pb.LoginResp{Param: req.Param}, nil
}

func (server *Service) LoginAfter(ctx context.Context) {
	span := rkgrpcctx.NewTraceSpan(ctx, "LoginAfter")
	defer rkgrpcctx.EndTraceSpan(ctx, span, true)

	time.Sleep(10 * time.Millisecond)
}

func (server *Service) CallA(ctx context.Context, req *pb.CallAReq) (*pb.CallAResp, error) {
	// create grpc client
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}
	// create connection with grpc-serverB
	connB, _ := grpc.Dial("localhost:2022", opts...)
	defer connB.Close()
	clientB := pb.NewServerBClient(connB)
	clientB.CallB(ctx, &pb.CallBReq{Param: req.Param})

	return &pb.CallAResp{Param: req.Param}, nil
}
