package group

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"rbaclock/conf"
	cli "rbaclock/pkg/client"
	"rbaclock/pkg/db"
	"strings"

	v1 "k8s.io/api/core/v1"
)

type Group struct {
	Label int
	Nodes []*v1.Node
	Pods  []*v1.Pod
}

func (c *Group) Show() {
	nodesString := ""
	podsString := ""
	for _, n := range c.Nodes {
		nodesString += n.Name + " "
	}
	for _, p := range c.Pods {
		podsString += p.Name + " "
	}
	log.Printf("\tLabel: %d\n", c.Label)
	log.Printf("\tNode: %s\n", nodesString)
	log.Printf("\tPod: %s\n", podsString)
}

func (c *Group) SAR(mode string) []float32 {
	vector := []float32{}
	if mode == "node" {
		// calculate by nodes
		for _, n := range c.Nodes {
			vec := db.QueryNodeVec(n.Name)
			if len(vector) == 0 {
				vector = vec
			} else {
				vector = db.MergeVector(vector, vec)
			}
		}
	} else {
		// calculate by pods
		for _, p := range c.Pods {
			vec := db.QueryPodVec(p.Name)
			if len(vector) == 0 {
				vector = vec
			} else {
				vector = db.MergeVector(vector, vec)
			}
		}
	}
	return vector
}

func (c *Group) Deploy() {
	key := conf.GlobalKey
	value := fmt.Sprintf("cluster-%d", c.Label)
	for _, node := range c.Nodes {
		cli.UpsertNodeAnchor(node, key, value)
	}
	for _, pod := range c.Pods {
		cli.UpsertPodAnchor(pod, key, value)
	}
}

func ClusterNum(subClusterSize int) int {
	nodes := cli.WorkerNodes()
	clusterNum := len(nodes) / subClusterSize
	if len(nodes)%subClusterSize != 0 {
		clusterNum += 1
	}
	return clusterNum
}

func GroupPods() []*Group {
	result := []*Group{}
	clusters := groupWithPy()
	for i, nsPodNames := range clusters {
		group := &Group{
			Label: i,
		}
		for _, nsPod := range nsPodNames {
			ns := strings.Split(nsPod, ":")[0]
			podName := strings.Split(nsPod, ":")[1]
			pod := cli.Pod(ns, podName)
			group.Pods = append(group.Pods, pod)
		}
		result = append(result, group)
	}
	return result
}

// func ClusterNodes(clusterNum int) []*Cluster {
// 	result := []*Cluster{}
// 	clusters := clusterWithScript(clusterNum, "node")
// 	for i, nodes := range clusters {
// 		cluster := &Cluster{
// 			Label: i,
// 		}
// 		for _, nodeName := range nodes {
// 			node := cli.Node(nodeName)
// 			cluster.Nodes = append(cluster.Nodes, node)
// 		}
// 		result = append(result, cluster)
// 	}
// 	return result
// }

func groupWithPy() [][]string {
	cmd := exec.Command("python", conf.ClusterScript, "-t", "pod", "-n", "10")
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	if e := cmd.Run(); e != nil {
		log.Fatal(err.String())
	}
	result := parseClusterResult(out.String())
	return result
}

func parseClusterResult(jsonStr string) [][]string {
	var result [][]string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Fatalf("Parse json error: %v\n", err)
	}

	return result
}

func CAR(clusters []*Group, mode string) []float32 {
	sars := [][]float32{}
	for _, c := range clusters {
		sars = append(sars, c.SAR(mode))
	}
	var cache []float32
	for _, vec := range sars {
		if len(cache) == 0 {
			cache = make([]float32, len(vec))
			copy(cache, vec)
		} else {
			cache = db.AddVector(cache, vec)
		}
	}
	for i := range cache {
		cache[i] /= float32(len(sars))
	}
	return cache
}
