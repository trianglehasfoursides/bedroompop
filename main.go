package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/trianglehasfoursides/bedroompop/config"
	"github.com/trianglehasfoursides/bedroompop/consist"
	"github.com/trianglehasfoursides/bedroompop/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	ws, err := os.OpenFile("bedroompop.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("can't open the log file")
	}

	core := zapcore.NewCore(enc, ws, zapcore.InfoLevel)

	logger := zap.New(core)
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	flag.StringVar(&config.Name, "name", "", "")
	flag.StringVar(&config.HTTPAddr, "http-address", ":7000", "")
	flag.StringVar(&config.GRPCAddr, "grpc-address", ":7070", "")
	flag.StringVar(&config.Join, "join", "", "")
	flag.StringVar(&config.Username, "username", "soy", "")
	flag.StringVar(&config.Password, "password", "pablo", "")
	flag.Parse()

	gossip, err := server.CreateGossip(config.GRPCAddr)
	if err != nil {
		return
	}

	consist.Consist.Add(consist.Member(config.GRPCAddr))
	if config.Join != "" {
	}

	slog.Info("starting Bedroompop")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go server.Start(ch)
	go server.GRPCStart(ch)

	<-ch
	slog.Info("stoping Bedroompop")
}
