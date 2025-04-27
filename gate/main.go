package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/trianglehasfoursides/mathrock/gate/server"
)

var (
	httpAddr   string
	gossipAddr string
	nodeName   string
	gossipPort int
)

func main() {
	// Initialize root command
	rootCmd := &cobra.Command{
		Use:   "mathrock-gate",
		Short: "CLI for managing the Mathrock Gate server",
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&httpAddr, "http-addr", ":8080", "HTTP server address")
	rootCmd.PersistentFlags().StringVar(&gossipAddr, "gossip-addr", "127.0.0.1", "Gossip protocol bind address")
	rootCmd.PersistentFlags().IntVar(&gossipPort, "gossip-port", 7946, "Gossip protocol bind port")
	rootCmd.PersistentFlags().StringVar(&nodeName, "node-name", "node-1", "Name of the gossip node")

	// Add subcommands
	rootCmd.AddCommand(startServerCmd())

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// startServerCmd defines the command to start the server
func startServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the HTTP and Gossip servers",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger
			logger, _ := zap.NewProduction()
			defer logger.Sync()
			zap.ReplaceGlobals(logger)

			// Create context for graceful shutdown
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle OS signals for graceful shutdown
			go func() {
				signalChan := make(chan os.Signal, 1)
				signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
				<-signalChan
				zap.L().Info("Received shutdown signal")
				cancel()
			}()

			// Start Gossip server
			go func() {
				zap.L().Info("Starting Gossip server...", zap.String("address", gossipAddr), zap.Int("port", gossipPort), zap.String("node", nodeName))
				server.Start(gossipAddr, nodeName, gossipPort)
			}()

			// Start HTTP server
			zap.L().Info("Starting HTTP server...", zap.String("address", httpAddr))
			if err := server.StartHTTPServer(ctx, httpAddr); err != nil {
				zap.L().Fatal("Failed to start HTTP server", zap.Error(err))
			}
		},
	}
}
