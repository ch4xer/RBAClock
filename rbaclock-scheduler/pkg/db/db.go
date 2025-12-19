package db

import (
	"context"
	"fmt"
	"scheduler/pkg/utils"
	"sync"

	log "k8s.io/klog/v2"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
)

var (
	pool     *pgxpool.Pool
	once     sync.Once
	database = "postgres://user:pass@localhost:5432/rbaclock?sslmode=disable"
	vecLen   int
)

func dbPool() *pgxpool.Pool {
	once.Do(func() {
		var err error
		ctx := context.Background()
		config, err := pgxpool.ParseConfig(database)
		if err != nil {
			log.Fatalf("Unable to parse db config: %v\n", err)
		}
		config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			return pgxvector.RegisterTypes(ctx, conn)
		}
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}
	})
	return pool
}

func init() {
	ctx := context.Background()
	pool := dbPool()
	var vec pgvector.Vector
	err := pool.QueryRow(ctx, "SELECT vec from knowlege_vec LIMIT 1").Scan(&vec)
	if err != nil {
		log.Errorf("Query knowlege_vec failed: %v", err)
	}
	vecLen = len(vec.Slice())
	InitStateTable(vecLen)
}

func InitStateTable(vecSize int) {
	CleanVectorTable()
	ctx := context.Background()
	pool := dbPool()
	_, err := pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		log.Fatalf("create extension failed: %v\n", err)
	}
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS state_vec (pod TEXT PRIMARY KEY, serviceaccount TEXT, namespace TEXT, node TEXT, vec VECTOR(%d))`, vecSize)
	_, err = pool.Exec(ctx, stmt)
	if err != nil {
		log.Fatalf("InitVectorTable failed: %v\n", err)
	}
}

func CleanVectorTable() {
	pool := dbPool()
	_, err := pool.Exec(context.Background(), "DROP TABLE IF EXISTS state_vec")
	if err != nil {
		log.Fatalf("CleanVectorTable failed: %v\n", err)
	}
}

func UpsertPodVec(pod, serviceaccount, namespace, node string, vec []float32) {
	pool := dbPool()
	query := `INSERT INTO state_vec (pod, serviceaccount, namespace, node, vec) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (pod) DO UPDATE SET serviceaccount=EXCLUDED.serviceaccount, namespace = EXCLUDED.namespace, node = EXCLUDED.node, vec = EXCLUDED.vec`
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
		log.Infof("QuerySAVec for %v failed: %v\n", serviceaccount, err)
		return make([]float32, vecLen)
	}
	return vec.Slice()
}

func QueryNodeVec(node string) []float32 {
	pool := dbPool()
	query := `SELECT vec FROM state_vec WHERE node = $1`
	rows, err := pool.Query(context.Background(), query, node)
	// no vector for the node
	if err != nil {
		log.Errorf("QueryNodeVec failed: %v\n", err)
		return make([]float32, vecLen)
	}
	defer rows.Close()
	resultVec := make([]float32, vecLen)
	for rows.Next() {
		var vec pgvector.Vector
		if err := rows.Scan(&vec); err != nil {
			log.Infof("Scan data error for %s", node)
			return resultVec
		}
		if len(resultVec) == 0 {
			resultVec = vec.Slice()
		} else {
			resultVec = utils.MergeVector(resultVec, vec.Slice())
		}
	}
	return resultVec
}

func QueryPodVecsOnNode(node string) [][]float32 {
	pool := dbPool()
	query := `SELECT vec FROM state_vec WHERE node = $1`
	rows, err := pool.Query(context.Background(), query, node)
	if err != nil {
		log.Errorf("QueryPodVecsOnNode failed: %v\n", err)
	}
	defer rows.Close()
	var resultVecs [][]float32
	for rows.Next() {
		var vec pgvector.Vector
		if err := rows.Scan(&vec); err != nil {
			log.Errorf("QueryPodVecsOnNode failed: %v\n", err)
			return resultVecs
		}
		resultVecs = append(resultVecs, vec.Slice())
	}
	return resultVecs
}
