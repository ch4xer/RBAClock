package measure

import (
	"sort"
)

func risk2Vec(riskMap map[string][]float32, weightMap map[string]float32) []float32 {
	for o, risks := range riskMap {
		for i, risk := range risks {
			riskMap[o][i] = risk * weightMap[o]
		}
	}
	keys := make([]string, 0, len(riskMap))
	for k := range riskMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	var riskVec []float32
	for _, k := range keys {
		riskVec = append(riskVec, riskMap[k]...)
	}
	return riskVec
}

// return index sorted by vectors
// func sortVec(indexes []string, vecs [][]float32) []string {
// 	for i := range indexes {
// 		for j := i + 1; j < len(indexes); j++ {
// 			if db.XORVector(vecs[i], vecs[j]) < 0 {
// 				indexes[i], indexes[j] = indexes[j], indexes[i]
// 				vecs[i], vecs[j] = vecs[j], vecs[i]
// 			}
// 		}
// 	}
// 	return indexes
// }

// group_num -> {node -> vector}
// func groupSimilarVec(target map[string][]float32, size int) map[int]map[string][]float32 {
// 	var indexes []string
// 	var values [][]float32
// 	for k, v := range target {
// 		indexes = append(indexes, k)
// 		values = append(values, v)
// 	}
// 	indexes = sortVec(indexes, values)
// 	groups := map[int]map[string][]float32{}
// 	for i := range len(indexes)/size + 1 {
// 		for j := i * size; j < (i+1)*size && j < len(indexes); j++ {
// 			if _, ok := groups[i]; !ok {
// 				groups[i] = make(map[string][]float32)
// 			}
// 			groups[i][indexes[j]] = target[indexes[j]]
// 		}
// 	}
// 	return groups
// }
