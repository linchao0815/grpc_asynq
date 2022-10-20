package main

import (
	"context"
	_ "embed"

	rkboot "github.com/rookie-ninja/rk-boot/v2"
	demo "github.com/rookie-ninja/rk-demo/api/gen/v1"
	"github.com/rookie-ninja/rk-demo/grpcInterceptor"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"google.golang.org/grpc"
)

//go:embed after-server.yaml
var afterServerConf []byte

func main() {
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(afterServerConf))

	// register grpc
	grpcEntry := rkgrpc.GetGrpcEntry("afterServer")
	grpcEntry.AddRegFuncGrpc(func(server *grpc.Server) {
		demo.RegisterAfterServer(server, &afterServer{})
	})
	grpcEntry.UnaryInterceptors = append(grpcEntry.UnaryInterceptors, grpcInterceptor.UnaryInterceptor)
	// Bootstrap
	boot.Bootstrap(context.TODO())

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}

type afterServer struct {
	demo.UnimplementedAfterServer
}

func (a afterServer) CreateAfter(ctx context.Context, req *demo.CreateAfterReq) (*demo.CreateAfterResp, error) {
	return &demo.CreateAfterResp{}, nil
}

func (a afterServer) UpdateAfter(ctx context.Context, req *demo.UpdateAfterReq) (*demo.UpdateAfterResp, error) {
	return &demo.UpdateAfterResp{}, nil
}
