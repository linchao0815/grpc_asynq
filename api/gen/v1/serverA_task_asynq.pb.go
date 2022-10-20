// Code generated by protoc-gen-go-asynqgen. DO NOT EDIT.
// versions:
// protoc-gen-go-asynqgen v1.0.10

package v1

import (
	context "context"
	json "encoding/json"
	asynq "github.com/hibiken/asynq"
	proto "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

import (
	"fmt"
	"net/http"
	"myasynq"
	"strings"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/attribute"
	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the asynq package it is being compiled against.
var _ = new(context.Context)
var _ = new(asynq.Task)
var _ = new(emptypb.Empty)
var _ = new(proto.Message)
var _ = new(json.InvalidUTF8Error)

type ServerA_TaskJobServer interface {
	GameTest_Task(context.Context, *GameTest_TaskReq) error
}

func RegisterServerA_TaskJobServer(mux *asynq.ServeMux, srv ServerA_TaskJobServer) {
	mux.HandleFunc("ServerA_Task:GameTest_Task", _ServerA_Task_GameTest_Task_Job_Handler(srv))
}

func _ServerA_Task_GameTest_Task_Job_Handler(srv ServerA_TaskJobServer) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		var in GameTest_TaskReq
		t := &myasynq.TaskPaylod{In: &in}
		if err := json.Unmarshal(task.Payload(), &t); err != nil {
			return fmt.Errorf("%s req=%s err=%s", task.Type(), t, err)
		}
		ctx, span := myasynq.NewSpan(ctx, "GameTest_Task")
		err := srv.GameTest_Task(ctx, t.In.(*GameTest_TaskReq))
		span.SetAttributes(attribute.String("req", myasynq.ToMarshal(t)))
		myasynq.EndSpan(span, err == nil)
		return err
	}
}

type ServerA_TaskSvcJob struct{}

var ServerA_TaskJob ServerA_TaskSvcJob

func (j *ServerA_TaskSvcJob) GameTest_Task(ctx context.Context, in *GameTest_TaskReq, opts ...asynq.Option) (*asynq.Task, *http.Header, error) {
	// get trace metadata
	header := http.Header{}
	rkgrpcctx.GetTracerPropagator(ctx).Inject(ctx, propagation.HeaderCarrier(header))
	payload, err := json.Marshal(myasynq.TaskPaylod{
		In:          in,
		TraceHeader: header,
	})
	if err != nil {
		return nil, nil, err
	}

	task := asynq.NewTask("ServerA_Task:GameTest_Task", payload, opts...)
	return task, &header, nil
}

type ServerA_TaskJobClient interface {
	GameTest_Task(ctx context.Context, req *GameTest_TaskReq, opts ...asynq.Option) (info *asynq.TaskInfo, err error)
}

type ServerA_TaskJobClientImpl struct {
	cc *asynq.Client
}

func NewServerA_TaskJobClient(client *asynq.Client) ServerA_TaskJobClient {
	return &ServerA_TaskJobClientImpl{client}
}

func (c *ServerA_TaskJobClientImpl) GameTest_Task(ctx context.Context, in *GameTest_TaskReq, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task, header, err := ServerA_TaskJob.GameTest_Task(ctx, in, opts...)
	if err != nil {
		return nil, fmt.Errorf("ServerA_TaskJob.GameTest_Task req:%s err:%s", in, err)
	}
	info, err := c.cc.Enqueue(task)
	if err != nil {
		return nil, fmt.Errorf("ServerA_TaskJob.GameTest_Task Enqueue req:%s err:%s", in, err)
	}
	// 把 Trace 信息，存入 Metadata，以 Header 的形式返回给 httpclient
	for k, v := range *header {
		rkgrpcctx.AddHeaderToClient(ctx, k, strings.Join(v, ","))
	}
	return info, nil
}