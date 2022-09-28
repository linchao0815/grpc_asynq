package main

import (
	"context"
	_ "embed"
	"fmt"
	"myasynq"
	"os"

	"github.com/hibiken/asynq"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	pb "github.com/rookie-ninja/rk-demo/api/gen/v1"
	"github.com/rookie-ninja/rk-demo/grpcInterceptor"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed grpc-serverB.yaml
var grpcBootB []byte

var redisAddr = "192.168.44.203:6379"
var redis_DbId = 11

func main() {
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(grpcBootB))
	if len(redisAddr) == 0 {
		redisAddr = os.Getenv("redis")
		if len(redisAddr) == 0 {
			panic("please setup env variable for redis")
		}
	}
	// register grpc
	grpcEntry := rkgrpc.GetGrpcEntry("grpcServerB")
	grpcEntry.AddRegFuncGrpc(registerServerB)
	grpcEntry.AddRegFuncGw(pb.RegisterServerBHandlerFromEndpoint)
	grpcEntry.UnaryInterceptors = append(grpcEntry.UnaryInterceptors, grpcInterceptor.UnaryInterceptor)

	// Bootstrap
	boot.Bootstrap(context.TODO())
	// start asynq server
	asynqServer := startAsynqWorker(grpcEntry.LoggerEntry.Logger)
	boot.AddShutdownHookFunc("asynq worker", asynqServer.Shutdown)

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}

type ServerB_TaskJobServer struct {
	pb.ServerB_TaskJobServer
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
	//mux.HandleFunc(task.TypeDemo, task.HandleDemoTask)
	pb.RegisterServerB_TaskJobServer(mux, &ServerB_TaskJobServer{})

	// add jaeger middleware
	jaegerMid, err := myasynq.NewJaegerMid(grpcBootB)
	if err != nil {
		rkentry.ShutdownWithError(err)
	}
	mux.Use(jaegerMid)

	if err := srv.Start(mux); err != nil {
		rkentry.ShutdownWithError(err)
	}

	return srv
}

func registerServerB(server *grpc.Server) {
	pb.RegisterServerBServer(server, &ServerB{})
}

type ServerB struct {
	pb.UnimplementedServerBServer
}

func (s ServerB) CallB(ctx context.Context, req *pb.CallBReq) (*pb.CallBResp, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, DB: redis_DbId})
	defer client.Close()
	_, err := (pb.NewServerB_TaskJobClient(client)).CallB_Task(ctx, &pb.CallB_TaskReq{Param: req.Param})
	if err != nil {
		return &pb.CallBResp{Param: req.Param}, fmt.Errorf("CallB_Task err:%s", err)
	}
	return &pb.CallBResp{Param: req.Param}, nil
}

func (s *ServerB_TaskJobServer) CallB_Task(ctx context.Context, in *pb.CallB_TaskReq) (err error) {
	connA, _ := grpc.DialContext(ctx, "localhost:2008", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(grpcInterceptor.ClientInterceptor()))
	defer connA.Close()
	clientA := pb.NewServerAClient(connA)
	// 把 trace info 载入到 context 中
	//ctx = rkgrpcctx.InjectSpanToNewContext(ctx) will painc
	clientA.Login(ctx, &pb.LoginReq{Param: in.Param})

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, DB: redis_DbId - 1})
	defer client.Close()
	_, err = (pb.NewServerB_CallBackJobClient(client)).CallB_CallBack(ctx, &pb.CallBReply{Param: in.Param})
	if err != nil {
		return fmt.Errorf("CallB_Task err:%s", err)
	}
	return nil
}
