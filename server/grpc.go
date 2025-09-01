package server

import (
	"context"
	"net"
	"os"

	"github.com/trianglehasfoursides/bedroompop/config"
	"github.com/trianglehasfoursides/bedroompop/database"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type server struct{}

func (s *server) Create(c context.Context, req *RequestCreate) (*DDLResponse, error) {
	if err := database.Create(req.Name, req.Migration); err != nil {
		return nil, err
	}
	return &DDLResponse{Msg: "sucess"}, nil
}

func (s *server) Drop(c context.Context, req *RequestGetDrop) (*DDLResponse, error) {
	if err := database.Drop(req.GetName()); err != nil {
		return nil, err
	}
	return &DDLResponse{Msg: "sucess"}, nil
}

func (s *server) Get(c context.Context, req *RequestGetDrop) (*DDLResponse, error) {
	if err := database.Get(req.GetName()); err != nil {
		return nil, err
	}
	return &DDLResponse{Msg: "sucess"}, nil
}

func (s *server) Query(c context.Context, req *RequestQueryExec) (*ResponseQuery, error) {
	if len(req.Args) > 0 {
		reqAny := make([]any, len(req.Args))
		for i, arg := range req.Args {
			msg, err := anypb.UnmarshalNew(arg, proto.UnmarshalOptions{})
			if err != nil {
				return nil, err
			}
			switch v := msg.(type) {
			case *wrapperspb.StringValue:
				reqAny[i] = v.GetValue()
			case *wrapperspb.DoubleValue:
				reqAny[i] = v.Value
			case *wrapperspb.BoolValue:
				reqAny[i] = v.Value
			default:
				reqAny[i] = nil
			}
		}
		result, err := database.Query(req.GetName(), req.GetQuery(), reqAny...)
		if err != nil {
			return nil, err
		}
		return &ResponseQuery{Result: result}, nil
	}

	result, err := database.Query(req.GetName(), req.GetQuery())
	if err != nil {
		return nil, err
	}
	return &ResponseQuery{Result: result}, nil
}

func (s *server) Exec(c context.Context, req *RequestQueryExec) (*ResponseExec, error) {
	if len(req.Args) > 0 {
		reqAny := make([]any, len(req.Args))
		for i, arg := range req.Args {
			msg, err := anypb.UnmarshalNew(arg, proto.UnmarshalOptions{})
			if err != nil {
				return nil, err
			}
			switch v := msg.(type) {
			case *wrapperspb.StringValue:
				reqAny[i] = v.GetValue()
			case *wrapperspb.DoubleValue:
				reqAny[i] = v.Value
			case *wrapperspb.BoolValue:
				reqAny[i] = v.Value
			default:
				reqAny[i] = nil
			}
		}
		result, err := database.Exec(req.GetName(), req.GetQuery(), reqAny...)
		if err != nil {
			return nil, err
		}
		return &ResponseExec{Result: result}, nil
	}
	result, err := database.Exec(req.GetName(), req.GetQuery())
	if err != nil {
		return nil, err
	}
	return &ResponseExec{Result: result}, nil
}

func (s *server) mustEmbedUnimplementedPopServiceServer() {}

func GRPCStart(ch chan os.Signal) {
	listener, err := net.Listen("tcp", config.GRPCAddr)
	if err != nil {
		zap.L().Sugar().Panic(err.Error())
	}
	popServer := grpc.NewServer()
	popService := &server{}
	RegisterPopServiceServer(popServer, popService)

	go func() {
		select {
		case _ = <-ch:
			popServer.GracefulStop()
		}
	}()

	if err := popServer.Serve(listener); err != nil {
		zap.L().Sugar().Panic(err.Error())
	}
}
