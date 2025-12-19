package measure

// import (
// 	"rbaclock/pkg/cluster"
// 	"rbaclock/pkg/db"
// )

// func MeasureSchedulePlan(clusters []*cluster.Cluster) float32 {
// 	sars := [][]float32{}
// 	for _, c := range clusters {
// 		sars = append(sars, MeasureCluster(c))
// 	}
// 	car := cluster.CAR(clusters, "pod")
// 	l1 := db.VectorL1(car)
// 	return l1
// }

// func MeasureCluster(c *cluster.Cluster) []float32 {
// 	// return sar metrics
// 	vector := []float32{}
// 	for _, p := range c.Pods {
// 		vec := db.QueryPodVec(p.Name)
// 		if len(vector) == 0 {
// 			vector = vec
// 		} else {
// 			vector = db.MergeVector(vector, vec)
// 		}
// 	}
// 	return vector
// }
