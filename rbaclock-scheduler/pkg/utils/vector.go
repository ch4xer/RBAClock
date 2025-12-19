package utils

import (
	log "k8s.io/klog/v2"
)

func MergeVector(vec1, vec2 []float32) []float32 {
	if len(vec1) != len(vec2) {
		log.Fatalf("vectors have different lengths, %d and %d\n", len(vec1), len(vec2))
	}
	for i := range vec1 {
		if vec1[i] < vec2[i] {
			vec1[i] = vec2[i]
		}
	}
	return vec1
}

func AddVector(vec1, vec2 []float32) []float32 {
	if len(vec1) < len(vec2) {
		vec1 = append(vec1, make([]float32, len(vec2)-len(vec1))...)
	} else if len(vec1) > len(vec2) {
		vec2 = append(vec2, make([]float32, len(vec1)-len(vec2))...)
	}

	for i := range vec1 {
		vec1[i] += vec2[i]
	}
	return vec1
}

func XORVector(vec1, vec2 []float32) []float32 {
	if len(vec1) != len(vec2) {
		return []float32{}
	}
	result := []float32{}
	for i := range vec1 {
		if vec1[i] > vec2[i] {
			result = append(result, vec1[i]-vec2[i])
		} else {
			result = append(result, vec2[i]-vec1[i])
		}
	}
	return result
}

func VectorL1(vec []float32) float32 {
	var result float32
	for _, v := range vec {
		result += v
	}
	return result
}
