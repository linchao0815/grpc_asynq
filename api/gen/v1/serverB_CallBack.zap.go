// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: serverB_CallBack.proto

package v1

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "github.com/linchao0815/protoc-gen-go-asynqgen/proto"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *CallBReply) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "Param" // field Param = 1
	enc.AddString(keyName, m.Param)

	return nil
}
