package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/trianglehasfoursides/bedroompop/consist"
	"github.com/trianglehasfoursides/bedroompop/dream"
	"github.com/trianglehasfoursides/bedroompop/flags"
	"github.com/trianglehasfoursides/bedroompop/pop"
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

	flag.StringVar(&flags.HTTPAddr, "http-address", ":7000", "")
	flag.StringVar(&flags.GRPCAddr, "grpc-address", ":7070", "")
	flag.StringVar(&flags.Join, "join", "", "")
	flag.StringVar(&flags.Username, "username", "soy", "")
	flag.StringVar(&flags.Password, "password", "pablo", "")
	flag.Parse()
	consist.Consist.Add(consist.Member(flags.GRPCAddr))
	if flags.Join != "" {
		for _, s := range strings.Split(flags.Join, " ") {
			consist.Consist.Add(consist.Member(s))
		}
	}
	slog.Info("starting Bedroompop")
	zap.L().Sugar().Infoln("starting bedroompop")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go dream.Start(ch)
	go pop.Start(ch)

	<-ch
	zap.L().Sugar().Infoln("stoping bedroompo")
	slog.Info("stoping Bedroompop")
}
