package main

import (
	"os"
	plugin "scheduler/pkg/plugin"

	"k8s.io/component-base/cli"
	"k8s.io/klog/v2" // Add klog import

	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

func main() {
	// Initialize klog
	klog.InitFlags(nil)
	defer klog.Flush()

	// Register custom plugins to the scheduler framework.
	// Later they can consist of scheduler profile(s) and hence
	// used by various kinds of workloads.
	command := app.NewSchedulerCommand(
		app.WithPlugin(plugin.Name, plugin.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
