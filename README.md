# Example
In this example, we will start two gRPC user-server and after-server and implement distributed logging.

## Install
```shell
go get github.com/rookie-ninja/rk-boot/v2
go get github.com/rookie-ninja/rk-grpc/v2
```

## Quick start
### 1.Create user-server.yaml and after-server.yaml

```yaml
app:
  name: userServer
grpc:
  - name: userServer
    enabled: true
    port: 2008
    enableReflection: true
    enableRkGwOption: true
    sw:
      enabled: true
      path: "sw"
      jsonPath: "api/gen/v1"
    middleware:
      trace:
        enabled: true
        exporter:
          jaeger:
            agent:
              enabled: false
            collector:
              enabled: true
      logging:
        enabled: true
```


```yaml
app:
  name: afterServer  # 为了能让 jaeger 辨别不同的服务，这里必须取名，否则为：rk
grpc:
  - name: afterServer
    enabled: true
    port: 2022
    enableReflection: true
    enableRkGwOption: true
    sw:
      enabled: true
      path: "sw"
      jsonPath: "api/gen/v1"
    middleware:
      trace:
        enabled: true
        exporter:
          jaeger:
            agent:
              enabled: false
            collector:
              enabled: true
      logging:
        enabled: true
```

### 2.Create user-server.go and after-server.go
Please refer to [user-server.go](user-server.go), [grpc-serverB.go](after-server.go) and [http-server.go](after-server.go)

### 3.Start user-server and after-server

```shell
$ go run user-server.go
$ go run after-server.go
$ curl localhost:2008/HealthCheck 
$ go run httpclient.go
```

2022-10-20T13:58:42.210+0800    ERROR   panic/interceptor.go:42 panic occurs:
goroutine 45 [running]:
runtime/debug.Stack()
        C:/Program Files/Go/src/runtime/debug/stack.go:24 +0x65
github.com/rookie-ninja/rk-grpc/v2/middleware/panic.UnaryServerInterceptor.func1.1()
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/panic/interceptor.go:42 +0x208
panic({0xc37fe0, 0xef92e0})
        C:/Program Files/Go/src/runtime/panic.go:838 +0x207
context.WithValue({0x0?, 0x0?}, {0xc367e0?, 0xef90f8?}, {0xd3a3a0?, 0x155fdf8?})
        C:/Program Files/Go/src/context/context.go:525 +0x178
go.opentelemetry.io/otel/trace.ContextWithSpan(...)
        C:/Users/linchao/go/pkg/mod/go.opentelemetry.io/otel/trace@v1.10.0/context.go:25
go.opentelemetry.io/otel/trace.noopTracer.Start({}, {0x0, 0x0}, {0xc0001afd70?, 0x0?}, {0xb?, 0xc0001ecd00?, 0x30e287?})
        C:/Users/linchao/go/pkg/mod/go.opentelemetry.io/otel/trace@v1.10.0/noop.go:53 +0x7c
github.com/rookie-ninja/rk-grpc/v2/middleware/context.GetTraceSpan({0x0, 0x0})
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/context/context.go:162 +0x7a
github.com/rookie-ninja/rk-demo/grpcInterceptor.UnaryInterceptor({0x0, 0x0}, {0xce4700, 0xc0001afb00}, 0xc00021e300?, 0xc000010258?)
        S:/NFT/test/grpc_asynq/grpcInterceptor/grpcInterceptor.go:19 +0xe5
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0x0?, 0x0?}, {0xce4700?, 0xc0001afb00?})
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1129 +0x5b
github.com/rookie-ninja/rk-grpc/v2/middleware/tracing.UnaryServerInterceptor.func1({0xf05328?, 0xc0001afb60?}, {0xce4700, 0xc0001afb00}, 0xc000512300, 0xc000012200)
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/tracing/server_interceptor.go:54 +0x794
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0xf05328?, 0xc0001afb60?}, {0xce4700?, 0xc0001afb00?})
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1132 +0x83
github.com/rookie-ninja/rk-grpc/v2/middleware/panic.UnaryServerInterceptor.func1({0xf05328?, 0xc0001afb60?}, {0xce4700, 0xc0001afb00}, 0xc0f0e0?, 0xc000012200)
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/panic/interceptor.go:46 +0x19e
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0xf05328?, 0xc0001afb60?}, {0xce4700?, 0xc0001afb00?})
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1132 +0x83
github.com/rookie-ninja/rk-grpc/v2/middleware/log.UnaryServerInterceptor.func1({0xf05328?, 0xc0001afad0?}, {0xce4700, 0xc0001afb00}, 0xc000512300, 0xc000012200)
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/log/server_interceptor.go:54 +0x8ef
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0xf05328?, 0xc0001afad0?}, {0xce4700?, 0xc0001afb00?})
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1132 +0x83
google.golang.org/grpc.chainUnaryInterceptors.func1({0xf05328, 0xc0001afad0}, {0xce4700, 0xc0001afb00}, 0xc000512300, 0xc000010258)
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1134 +0x12b
github.com/rookie-ninja/rk-demo/api/gen/v1._User_HealthCheck_Handler({0xcc0ee0?, 0xc0001f9480}, {0xf05328, 0xc0001afad0}, 0xc000146380, 0xc000115a00)
        S:/NFT/test/grpc_asynq/api/gen/v1/user_grpc.pb.go:116 +0x138
google.golang.org/grpc.(*Server).processUnaryRPC(0xc00047c780, {0xf08ef0, 0xc0000736c0}, 0xc0005f0000, 0xc0002a31a0, 0x14f05c0, 0x0)
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1295 +0xb0b
google.golang.org/grpc.(*Server).handleStream(0xc00047c780, {0xf08ef0, 0xc0000736c0}, 0xc0005f0000, 0x0)
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1636 +0xa1b
google.golang.org/grpc.(*Server).serveStreams.func1.2()
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:932 +0x98
created by google.golang.org/grpc.(*Server).serveStreams.func1
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:930 +0x28a
        {"error": "rpc error: code = Internal desc = cannot create context from nil parent"}
### 4.Send request and validate logs
```shell
curl -X GET 'http://localhost:2008/v1/user/create?name=hihi'   
{"traceId":"96ef777f259322d758821919529ec8cc"}

curl -X 'POST' 'http://localhost:2008/v1/CreateUser' -H 'accept: application/json' -H 'Content-Type: application/json'  -d '{"name": "string"}'
curl -X 'POST' 'http://localhost:2008/v1/CreateUser' -d '{"name": "string"}'

```

- check logs at user-server

```shell
------------------------------------------------------------------------
...
ids={"eventId":"96becb7b-8ddb-4e64-91ed-869022913487","traceId":"96ef777f259322d758821919529ec8cc"}
...
operation=/api.v1.User/CreateUser
...
```

- check logs at task server

```shell
traceId: 96ef777f259322d758821919529ec8cc
```

- check logs at after-server

```shell
------------------------------------------------------------------------
...
ids={"eventId":"a75d0c6d-b54d-476a-b16c-f4ede087ef3a","traceId":"96ef777f259322d758821919529ec8cc"}
...
operation=/api.v1.After/CreateAfter
...
```