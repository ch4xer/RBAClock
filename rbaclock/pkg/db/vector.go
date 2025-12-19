package db

import (
	"context"
	"fmt"
	"log"

	"github.com/pgvector/pgvector-go"
)

func InitVectorTable(vecSize int) {
	CleanVectorTable()
	ctx := context.Background()
	pool := dbPool()
	_, err := pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		log.Fatalf("create extension failed: %v\n", err)
	}
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS knowlege_vec (pod TEXT PRIMARY KEY, serviceaccount TEXT, namespace TEXT, node TEXT, vec VECTOR(%d))`, vecSize)
	_, err = pool.Exec(ctx, stmt)
	if err != nil {
		log.Fatalf("InitVectorTable failed: %v\n", err)
	}
}

func CleanVectorTable() {
	pool := dbPool()
	_, err := pool.Exec(context.Background(), "DROP TABLE IF EXISTS knowlege_vec")
	if err != nil {
		log.Fatalf("CleanVectorTable failed: %v\n", err)
	}
}

func UpsertPodVec(pod, serviceaccount, namespace, node string, vec []float32) {
	pool := dbPool()
	query := `INSERT INTO knowlege_vec (pod, serviceaccount, namespace, node, vec) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (pod) DO UPDATE SET serviceaccount=EXCLUDED.serviceaccount, namespace = EXCLUDED.namespace, node = EXCLUDED.node, vec = EXCLUDED.vec`
	_, err := pool.Exec(context.Background(), query, pod, serviceaccount, namespace, node, pgvector.NewVector(vec))
	if err != nil {
		log.Fatalf("UpsertPodVec failed: %v\n", err)
	}
}

func QuerySAVec(serviceaccount string) []float32 {
	pool := dbPool()
	query := `SELECT vec FROM knowlege_vec WHERE serviceaccount = $1`
	var vec pgvector.Vector
	err := pool.QueryRow(context.Background(), query, serviceaccount).Scan(&vec)
	if err != nil {
		log.Printf("QuerySAVec failed: %v\n", err)
		return []float32{}
	}
	return vec.Slice()
}

func QueryNodeVec(node string) []float32 {
	pool := dbPool()
	query := `SELECT vec FROM knowlege_vec WHERE node = $1`
	rows, err := pool.Query(context.Background(), query, node)
	if err != nil {
		log.Printf("QueryNodeVec failed: %v\n", err)
		return []float32{}
	}
	defer rows.Close()
	var resultVec []float32
	for rows.Next() {
		var vec pgvector.Vector
		if err := rows.Scan(&vec); err != nil {
			return []float32{}
		}
		if len(resultVec) == 0 {
			resultVec = vec.Slice()
		} else {
			resultVec = MergeVector(resultVec, vec.Slice())
		}
	}
	return resultVec
}

func MergeVector(vec1, vec2 []float32) []float32 {
	if len(vec1) != len(vec2) {
		log.Fatalf("mergeVector failed: vectors have different lengths")
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
		panic("Vector length mismatch")
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

func QueryPodVec(pod string) []float32 {
	pool := dbPool()
	query := `SELECT vec FROM knowlege_vec WHERE pod = $1`
	var vec pgvector.Vector
	err := pool.QueryRow(context.Background(), query, pod).Scan(&vec)
	if err != nil {
		log.Printf("QueryPodVec failed: %v\n", err)
		return []float32{}
	}
	return vec.Slice()
}

func QueryPodNamespace(pod string) string {
	pool := dbPool()
	query := `SELECT namespace FROM knowlege_vec WHERE pod = $1`
	var namespace string
	err := pool.QueryRow(context.Background(), query, pod).Scan(&namespace)
	if err != nil {
		log.Fatalf("QueryPodNamespace failed: %v\n", err)
	}
	return namespace
}
