package cmd

import (
	"fmt"

	"rbaclock/pkg/scan"

	"github.com/spf13/cobra"
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan potential harmness caused by RBAC privileges on each node",
		Run:   runScan,
	}
)

func runScan(cmd *cobra.Command, args []string) {
	nsAlerts := scan.Scan()
	for ns, alerts := range nsAlerts {
		fmt.Printf("Namespace: %s\n", ns)
		for _, alert := range alerts {
			if len(alert.Risks) == 0 {
				continue
			}
			fmt.Printf("\tServiceAccount: %s\n", alert.ServiceAccount)
			for _, e := range alert.Risks {
				fmt.Printf("\t\t%s;\n", e.Desc)
			}
			fmt.Printf("\t\t%s\n", alert.Privileges)
			fmt.Printf("\n")
		}
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
