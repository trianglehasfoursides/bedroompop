package pop

import (
	"context"
	"net"
	"os"

	"github.com/trianglehasfoursides/bedroompop/database"
	"github.com/trianglehasfoursides/bedroompop/flags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	result, err := database.Query(req.GetName(), req.GetQuery())
	if err != nil {
		return nil, err
	}
	return &ResponseQuery{Result: string(result)}, nil
}

func (s *server) Exec(c context.Context, req *RequestQueryExec) (*ResponseExec, error) {
	result, err := database.Exec(req.GetName(), req.GetQuery())
	if err != nil {
		return nil, err
	}
	return &ResponseExec{Result: result}, nil
}

func (s *server) mustEmbedUnimplementedPopServiceServer() {}

func Start(ch chan os.Signal) {
	listener, err := net.Listen("tcp", flags.GRPCAddr)
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
