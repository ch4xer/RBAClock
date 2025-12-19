package cmd

import (
	"rbaclock/conf"
	"rbaclock/pkg/measure"
	"rbaclock/pkg/schedule"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Reschedule pods with clustering(grouping)",
	Run:   runSchedule,
}

func runSchedule(cmd *cobra.Command, args []string) {
	measure.RecordRiskVec()
	groups := schedule.SchedulePod()
	for _, g := range groups {
		g.Show()
	}

	// measure again!
	// car := measure.MeasureSchedulePlan(groups)

	// try to deploy
	for _, c := range groups {
		if conf.Deploy {
			c.Deploy()
		}
	}
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}
