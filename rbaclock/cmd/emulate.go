package cmd

import (
	"rbaclock/pkg/emulate"
	"rbaclock/pkg/measure"

	"github.com/spf13/cobra"
)

var (
	emulateCmd = &cobra.Command{
		Use:   "emulate",
		Short: "Emulate the risk of each ndoe",
		Run:   runEmulate,
	}
)

func runEmulate(cmd *cobra.Command, args []string) {
	measure.RecordRiskVec()
	emulate.Emulate()
}

func init() {
	rootCmd.AddCommand(emulateCmd)
}
