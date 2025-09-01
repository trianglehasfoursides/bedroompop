package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/trianglehasfoursides/bedroompop/config"
	"github.com/trianglehasfoursides/bedroompop/consist"
	"github.com/trianglehasfoursides/bedroompop/server"
)

func main() {
	flag.StringVar(&config.Name, "name", "", "")
	flag.StringVar(&config.HTTPAddr, "http-address", ":7000", "")
	flag.StringVar(&config.GRPCAddr, "grpc-address", ":7070", "")
	flag.StringVar(&config.GossipAddr, "gossip-address", "localhost:7777", "")
	flag.StringVar(&config.Join, "join", "", "")
	flag.StringVar(&config.Username, "username", "soy", "")
	flag.StringVar(&config.Password, "password", "pablo", "")
	flag.Parse()

	gossip, err := server.CreateGossip(config.GRPCAddr, config.GossipAddr, config.Name)
	if err != nil {
		return
	}

	consist.Consist.Add(consist.Member(config.GRPCAddr))
	if config.Join != "" {
		gossip.Join(strings.Split(config.Join, " "))
	}

	log.Info("starting Bedroompop")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go server.Start(ch)
	go server.GRPCStart(ch)

	<-ch
	log.Info("stoping Bedroompop")
}
