package test

import (
	"log"
	"rbaclock/pkg/measure"
	"rbaclock/pkg/schedule"
	"testing"
)

func TestSchedule(t *testing.T) {
	measure.RecordRiskVec()
	clusters := schedule.SchedulePod()
	for _, c := range clusters {
		log.Printf("label: %d; sar: %v\n", c.Label, c.SAR("pod"))
	}
}

// func TestScheduleManual(t *testing.T) {
// 	node := client.Node("vm-m05")
// 	_ = client.UpsertNodeAnchoraa(node, "rbaclock", "vm-m05")
// 	err := client.UpsertPodAnchor("kubeplus", "kubeplus-deployment-5c44d7c467-9kh29", "rbaclock", "vm-m05")
// 	if err != nil {
// 		t.Error(err)
// 	}
// }
