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

2022-10-20T12:18:31.335+0800    ERROR   panic/interceptor.go:42                                                   panic occurs:
goroutine 104 [running]:
runtime/debug.Stack()
        C:/Program Files/Go/src/runtime/debug/stack.go:24 +0x65
github.com/rookie-ninja/rk-grpc/v2/middleware/panic.UnaryServerInterceptor.func1.1()
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/panic/interceptor.go:42 +0x208
panic({0xa47fe0, 0xd092e0})
        C:/Program Files/Go/src/runtime/panic.go:838 +0x207
context.WithValue({0x0?, 0x0?}, {0xa467e0?, 0xd090f8?}, {0xb4a3a0?, 0x136fdf8?})
        C:/Program Files/Go/src/context/context.go:525 +0x178
go.opentelemetry.io/otel/trace.ContextWithSpan(...)
        C:/Users/linchao/go/pkg/mod/go.opentelemetry.io/otel/trace@v1.10.0/context.go:25
go.opentelemetry.io/otel/trace.noopTracer.Start({}, {0x0, 0x0}, {0xc000721530?, 0x0?}, {0xb?, 0xc00041ad00?, 0x11e287?})
        C:/Users/linchao/go/pkg/mod/go.opentelemetry.io/otel/trace@v1.10.0/noop.go:53 +0x7c
github.com/rookie-ninja/rk-grpc/v2/middleware/context.GetTraceSpan({0x0, 0x0})
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/context/context.go:162 +0x7a
github.com/rookie-ninja/rk-demo/grpcInterceptor.UnaryInterceptor({0x0, 0x0}, {0xaf4700, 0xc0007212c0}, 0xc0003f6300?, 0xc000282588?)
        S:/NFT/grpc_asynq/grpcInterceptor/grpcInterceptor.go:19 +0xe5
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0x0?, 0x0?}, {0xaf4700?, 0xc0007212c0?})
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1129 +0x5b
github.com/rookie-ninja/rk-grpc/v2/middleware/tracing.UnaryServerInterceptor.func1({0xd15328?, 0xc000721320?}, {0xaf4700, 0xc0007212c0}, 0xc00055e560, 0xc000089180)
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/tracing/server_interceptor.go:54 +0x794
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0xd15328?, 0xc000721320?}, {0xaf4700?, 0xc0007212c0?})      
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1132 +0x83
github.com/rookie-ninja/rk-grpc/v2/middleware/panic.UnaryServerInterceptor.func1({0xd15328?, 0xc000721320?}, {0xaf4700, 0xc0007212c0}, 0xa1f0e0?, 0xc000089180)
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/panic/interceptor.go:46 +0x19e
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0xd15328?, 0xc000721320?}, {0xaf4700?, 0xc0007212c0?})      
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1132 +0x83
github.com/rookie-ninja/rk-grpc/v2/middleware/log.UnaryServerInterceptor.func1({0xd15328?, 0xc000721290?}, {0xaf4700, 0xc0007212c0}, 0xc00055e560, 0xc000089180)
        C:/Users/linchao/go/pkg/mod/github.com/rookie-ninja/rk-grpc/v2@v2.2.8/middleware/log/server_interceptor.go:54 +0x8ef
google.golang.org/grpc.chainUnaryInterceptors.func1.1({0xd15328?, 0xc000721290?}, {0xaf4700?, 0xc0007212c0?})      
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1132 +0x83
google.golang.org/grpc.chainUnaryInterceptors.func1({0xd15328, 0xc000721290}, {0xaf4700, 0xc0007212c0}, 0xc00055e560, 0xc000282588)
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1134 +0x12b
github.com/rookie-ninja/rk-demo/api/gen/v1._User_HealthCheck_Handler({0xad0ee0?, 0xc0001f9180}, {0xd15328, 0xc000721290}, 0xc0002a4850, 0xc0002a3a00)
        S:/NFT/grpc_asynq/api/gen/v1/user_grpc.pb.go:116 +0x138
google.golang.org/grpc.(*Server).processUnaryRPC(0xc000476780, {0xd18ef0, 0xc0004e6d00}, 0xc0000ca480, 0xc000113710, 0x13005c0, 0x0)
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1295 +0xb0b
google.golang.org/grpc.(*Server).handleStream(0xc000476780, {0xd18ef0, 0xc0004e6d00}, 0xc0000ca480, 0x0)
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:1636 +0xa1b
google.golang.org/grpc.(*Server).serveStreams.func1.2()
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:932 +0x98
created by google.golang.org/grpc.(*Server).serveStreams.func1
        C:/Users/linchao/go/pkg/mod/google.golang.org/grpc@v1.48.0/server.go:930 +0x28a
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