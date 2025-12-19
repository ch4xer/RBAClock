package cmd

import (
	"log"
	"rbaclock/pkg/measure"

	"github.com/spf13/cobra"
)

var (
	measureCmd = &cobra.Command{
		Use:   "measure",
		Short: "Measure the ERP on all nodes",
		Run:   runMeasure,
	}
)

func runMeasure(cmd *cobra.Command, args []string) {
	measure.RecordRiskVec()
	erp := measure.MeasureCluster()
	log.Printf("ERP: %f\n", erp)
	// for _, subClusterSize := range conf.SubClusterSizes {
	// 	car, clusters := measure.MeasureNode(subClusterSize)
	// 	log.Printf("CAR-%d: %f\n", subClusterSize, car)
	// 	for _, c := range clusters {
	// 		log.Printf("label: %d; sar: %v\n", c.Label, c.SAR("node"))
	// 	}
	// }
}

func init() {
	rootCmd.AddCommand(measureCmd)
}
