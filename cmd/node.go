package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	nodeCmd.AddCommand(startCmd)
}

var (
	nodeCmd = &cobra.Command{
		Use: "node",
	}

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the Mathrock Node",
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("").Value.String() != "master" || cmd.Flag("").Value.String() != "slave" {
				zap.L().Sugar().Fatal("can only be master or slave")
				return
			}
		},
	}
)
