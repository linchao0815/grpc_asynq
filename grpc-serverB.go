package main

import (
	"context"
	_ "embed"

	rkboot "github.com/rookie-ninja/rk-boot/v2"
	greeter "github.com/rookie-ninja/rk-demo/api/gen/v1"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//go:embed grpc-serverB.yaml
var grpcBootB []byte

func main() {
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(grpcBootB))

	// register grpc
	grpcEntry := rkgrpc.GetGrpcEntry("grpcServerB")
	grpcEntry.AddRegFuncGrpc(registerServerB)
	grpcEntry.AddRegFuncGw(greeter.RegisterServerBHandlerFromEndpoint)

	// Bootstrap
	boot.Bootstrap(context.TODO())

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.TODO())
}

func registerServerB(server *grpc.Server) {
	greeter.RegisterServerBServer(server, &ServerB{})
}

type ServerB struct{}

func (s ServerB) CallB(ctx context.Context, req *greeter.CallBReq) (*greeter.CallBResp, error) {
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
	grpcCtx := ctx
	md := metadata.Pairs()
	rkgrpcctx.GetTracerPropagator(ctx).Inject(grpcCtx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
	grpcCtx = metadata.NewOutgoingContext(grpcCtx, md)
	// call gRPC server B
	clientA.Login(grpcCtx, &greeter.LoginReq{})
	return &greeter.CallBResp{}, nil
}
