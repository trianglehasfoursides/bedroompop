package main

import (
	"fmt"
	"os"
	"time"

	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serverAddr string

func main() {
	// Initialize root command
	rootCmd := &cobra.Command{
		Use:   "mathrock",
		Short: "CLI client for interacting with the Mathrock server",
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")

	// Add subcommands
	rootCmd.AddCommand(
		listDatabasesCmd(),
		createDatabaseCmd(),
		getDatabaseCmd(),
		deleteDatabaseCmd(),
	)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// listDatabasesCmd defines the command to list all databases
func listDatabasesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all databases",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/databases", serverAddr))
			if err != nil {
				zap.L().Error("Failed to list databases", zap.Error(err))
				return
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		},
	}
}

// createDatabaseCmd defines the command to create a new database
func createDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new database",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			payload := fmt.Sprintf(`{"name": "%s"}`, name)
			resp, err := http.Post(fmt.Sprintf("%s/api/v1/databases", serverAddr), "application/json", bytes.NewBuffer([]byte(payload)))
			if err != nil {
				zap.L().Error("Failed to create database", zap.Error(err))
				return
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		},
	}
}

// getDatabaseCmd defines the command to get a specific database by name
func getDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [name]",
		Short: "Get details of a specific database",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/databases/%s", serverAddr, name))
			if err != nil {
				zap.L().Error("Failed to get database", zap.Error(err))
				return
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		},
	}
}

// deleteDatabaseCmd defines the command to delete a specific database by name
func deleteDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a specific database",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			client := &http.Client{Timeout: 10 * time.Second}
			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/databases/%s", serverAddr, name), nil)
			if err != nil {
				zap.L().Error("Failed to create delete request", zap.Error(err))
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				zap.L().Error("Failed to delete database", zap.Error(err))
				return
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		},
	}
}
