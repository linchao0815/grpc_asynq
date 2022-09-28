# Example
In this example, we will start two gRPC servers serverA and serverB and implement distributed logging.

## Install
```shell
go get github.com/rookie-ninja/rk-boot/v2
go get github.com/rookie-ninja/rk-grpc/v2
go get github.com/rookie-ninja/rk-gin/v2

# install protoc-gen-go-asynqgen
go install github.com/linchao0815/protoc-gen-go-asynqgen@v1.0.11
```

## Quick start
### 1.Create bootA.yaml and bootB.yaml

```yaml
grpc:
  - name: grpcServerA
    enabled: true
    port: 2008
    middleware:
      trace:
        enabled: true
        exporter:
          jaeger:
            agent:
              enabled: true
      logging:
        enabled: true
```

- grpc-serverB.yaml

```yaml
grpc:
  - name: grpcServerB
    enabled: true
    port: 2022
    middleware:
      trace:
        enabled: true
        exporter:
          jaeger:
            agent:
              enabled: true
      logging:
        enabled: true
```

### 2.Create grpc-serverA.go, grpc-serverB.go and http-server.go
Please refer to [grpc-serverA.go](grpc-serverA.go), [grpc-serverB.go](grpc-serverB.go) and [http-server.go](http-server.go)

> How to propagate span from http-server to grpc-serverA and grpc-serverB?
> 
> First, get span context from http-server which generated automatically by rk-boot. Second, inject span context into metadata and create new context
> ```go
> grpcCtx := trace.ContextWithRemoteSpanContext(context.Background(), rkginctx.GetTraceSpan(ctx).SpanContext())
> md := metadata.Pairs()
> rkginctx.GetTracerPropagator(ctx).Inject(grpcCtx, &rkgrpcctx.GrpcMetadataCarrier{Md: &md})
> grpcCtx = metadata.NewOutgoingContext(grpcCtx, md)
> ```

### 3.Start serverA and serverB

```shell
$ go run grpc-serverA.go
$ go run grpc-serverB.go
$ go run httpclient.go
```
