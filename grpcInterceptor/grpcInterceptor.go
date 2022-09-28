package grpcInterceptor

import (
	"context"
	"myasynq"
	"strings"

	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (reply interface{}, err error) {
	if strings.HasSuffix(info.FullMethod, "HeathCheck") {
		return handler(ctx, req)
	}

	reply, err = handler(ctx, req)
	span := rkgrpcctx.GetTraceSpan(ctx)
	if span != nil {
		span.SetAttributes(attribute.String("req", myasynq.ToMarshal(req)), attribute.String("reply", myasynq.ToMarshal(reply)))
	}

	return reply, err
}

func ClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, request, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx := rkgrpcctx.InjectSpanToNewContext(ctx) // Inject current trace information into context
		return invoker(newCtx, method, request, reply, cc, opts...)
	}
}
