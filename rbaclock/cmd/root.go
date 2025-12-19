package cmd

import (
	"github.com/spf13/cobra"
	"os"

)

var (
	rootCmd = &cobra.Command{
		Use:   "rbaclock",
		Short: "See and evaluate RBAC permissions in Kubernetes clusters",
	}
)


func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
