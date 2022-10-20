// Code generated by protoc-gen-go-asynqgen. DO NOT EDIT.
// versions:
// protoc-gen-go-asynqgen v1.0.12

package v1

import (
	context "context"
	json "encoding/json"
	asynq "github.com/hibiken/asynq"
	proto "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

import (
	"myasynq"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	rkgrpcctx "github.com/rookie-ninja/rk-grpc/v2/middleware/context"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the asynq package it is being compiled against.
var _ = new(context.Context)
var _ = new(asynq.Task)
var _ = new(emptypb.Empty)
var _ = new(proto.Message)
var _ = new(json.InvalidUTF8Error)

type UserTaskServer interface {
	CreateUser(context.Context, *CreateUserPayload) error
	UpdateUser(context.Context, *UpdateUserPayload) error
}

func RegisterUserTaskServer(mux *asynq.ServeMux, srv UserTaskServer) {
	mux.HandleFunc("user:create", _User_CreateUser_Task_Handler(srv))
	mux.HandleFunc("user:update", _User_UpdateUser_Task_Handler(srv))
}

func _User_CreateUser_Task_Handler(srv UserTaskServer) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		in := &CreateUserPayload{}

		ctx, span, err := myasynq.Handle_task_before(ctx, task, in)
		if err != nil {
			return err
		}

		err = srv.CreateUser(ctx, in)

		myasynq.Handle_task_after(span, err)

		return err
	}
}

func _User_UpdateUser_Task_Handler(srv UserTaskServer) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		in := &UpdateUserPayload{}

		ctx, span, err := myasynq.Handle_task_before(ctx, task, in)
		if err != nil {
			return err
		}

		err = srv.UpdateUser(ctx, in)

		myasynq.Handle_task_after(span, err)

		return err
	}
}

type UserTaskClient interface {
	CreateUser(ctx context.Context, req *CreateUserPayload, opts ...asynq.Option) (info *asynq.TaskInfo, span oteltrace.Span, err error)
	UpdateUser(ctx context.Context, req *UpdateUserPayload, opts ...asynq.Option) (info *asynq.TaskInfo, span oteltrace.Span, err error)
}

type UserTaskClientImpl struct {
	cc *asynq.Client
}

func NewUserTaskClient(client *asynq.Client) UserTaskClient {
	return &UserTaskClientImpl{client}
}

func (c *UserTaskClientImpl) CreateUser(ctx context.Context, in *CreateUserPayload, opts ...asynq.Option) (*asynq.TaskInfo, oteltrace.Span, error) {
	if rkgrpcctx.GetTracerPropagator(ctx) != nil {
		ctx = rkgrpcctx.InjectSpanToNewContext(ctx)
	}

	spanCtx := oteltrace.SpanContextFromContext(ctx)
	ctx, span := myasynq.HolderTracer().Start(oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx), "CreateUserClient")
	defer span.End()

	// get trace metadata
	m := make(map[string]string)
	myasynq.HolderPropagator().Inject(ctx, propagation.MapCarrier(m))

	wrap, err := json.Marshal(myasynq.WrapPayload{
		Trace:   m,
		Payload: in,
	})
	if err != nil {
		return nil, nil, err
	}

	task := asynq.NewTask("user:create", wrap, opts...)

	info, err := c.cc.Enqueue(task)
	if err != nil {
		return nil, nil, err
	}
	return info, span, nil
}

func (c *UserTaskClientImpl) UpdateUser(ctx context.Context, in *UpdateUserPayload, opts ...asynq.Option) (*asynq.TaskInfo, oteltrace.Span, error) {
	if rkgrpcctx.GetTracerPropagator(ctx) != nil {
		ctx = rkgrpcctx.InjectSpanToNewContext(ctx)
	}

	spanCtx := oteltrace.SpanContextFromContext(ctx)
	ctx, span := myasynq.HolderTracer().Start(oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx), "UpdateUserClient")
	defer span.End()

	// get trace metadata
	m := make(map[string]string)
	myasynq.HolderPropagator().Inject(ctx, propagation.MapCarrier(m))

	wrap, err := json.Marshal(myasynq.WrapPayload{
		Trace:   m,
		Payload: in,
	})
	if err != nil {
		return nil, nil, err
	}

	task := asynq.NewTask("user:update", wrap, opts...)

	info, err := c.cc.Enqueue(task)
	if err != nil {
		return nil, nil, err
	}
	return info, span, nil
}
