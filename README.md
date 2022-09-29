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
```

### 4.Send request and validate logs
```shell
curl -X GET 'http://localhost:2008/v1/user/create?name=hihi'   
{"traceId":"96ef777f259322d758821919529ec8cc"}
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