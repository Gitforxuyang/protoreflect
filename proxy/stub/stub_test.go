package stub

import (
	"context"
	"reflect"
	"testing"

	"github.com/Gitforxuyang/protoreflect/errors"
	"github.com/Gitforxuyang/protoreflect/metadata"
	"github.com/Gitforxuyang/protoreflect/proxy/reflection"
	"github.com/Gitforxuyang/protoreflect/proxy/test"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/test/grpc_testing"
)

func TestNewStub(t *testing.T) {
	cc, err := grpc.Dial("localhost:5000", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err.Error())
	}
	NewStub(grpcdynamic.NewStub(cc))
}

func TestStub_InvokeRPC(t *testing.T) {
	cases := []struct {
		name           string
		methodName     string
		outputMsgIsNil bool
		error
	}{
		{
			name:           "success",
			methodName:     "EmptyCall",
			outputMsgIsNil: false,
			error:          nil,
		},
		{
			name:           "grpc error",
			methodName:     "UnaryCall",
			outputMsgIsNil: true,
			error: &errors.GRPCError{
				StatusCode: int(codes.Unimplemented),
				Message:    "unary unimplemented",
			},
		},
	}
	fileDesc := test.NewFileDescriptor(t, test.File)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			serviceDesc := reflection.ServiceDescriptorFromFileDescriptor(fileDesc, test.TestService)
			if serviceDesc == nil {
				t.Fatal("service descriptor is nil")
			}
			methodDesc, err := serviceDesc.FindMethodByName(tc.methodName)
			if err != nil {
				t.Fatal(err.Error())
			}
			inputMsgDesc := methodDesc.GetInputType()
			inputMsg := inputMsgDesc.NewMessage()
			ctx := context.Background()

			stub := &stubImpl{
				stub: &test.MockGrpcdynamicStub{},
			}
			invocation := &reflection.MethodInvocation{
				MethodDescriptor: methodDesc,
				Message:          inputMsg,
			}
			outputMsg, err := stub.InvokeRPC(ctx, invocation, (*metadata.Metadata)(&map[string][]string{}))
			if err != nil {
				switch v := err.(type) {
				case *errors.ProxyError:
					expected := tc.error.(*errors.ProxyError)
					if got, want := v, expected; !reflect.DeepEqual(got, want) {
						t.Fatalf("got %v, want %v", got, want)
					}
				case *errors.GRPCError:
					expected := tc.error.(*errors.GRPCError)
					if got, want := v, expected; !reflect.DeepEqual(got, want) {
						t.Fatalf("got %#v, want %#v", got, want)
					}
				}
			}
			if got, want := outputMsg == nil, tc.outputMsgIsNil; got != want {
				t.Fatalf("got %t, want %t", got, want)
			}
		})
	}
}
