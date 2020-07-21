/**
 * @Author: karen
 * @Email: hainazhitong@foxmail.com
 * @Version: 1.0.0
 * @Date: 2020/7/20 15:55
 * @Description: //TODO
 */

package validator

import (
	"context"
	"reflect"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// https://github.com/envoyproxy/protoc-gen-validate
// Client和Server的拦截验证

const funcName = "Validate"

func validate(req interface{}) error {
	t := reflect.TypeOf(req)
	if m, ok := t.MethodByName(funcName); ok {
		if e := m.Func.Call([]reflect.Value{reflect.ValueOf(req)}); len(e) > 0 {
			errInter := e[0].Interface()
			if errInter != nil {
				return errInter.(error)
			}
		}
	}
	return nil
}

// UnaryClientInterceptor returns a new unary client interceptor for validate.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := validate(req); err != nil {
			return status.Errorf(codes.InvalidArgument, err.Error())
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for validate.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validate(req); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return handler(ctx, req)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{stream}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if err := validate(m); err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	return nil
}
