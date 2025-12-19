package schedule

import (
	cli "rbaclock/pkg/client"
	"rbaclock/pkg/group"
)

func SchedulePod() []*group.Group {
	groups := group.GroupPods()
	groupSize := len(groups)

	for i, node := range cli.WorkerNodes() {
		label := i % groupSize
		groups = addNode2Group(groups, label, node)
	}

	return groups
}
