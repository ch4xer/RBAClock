package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
	"log"
	"rbaclock/conf"
	"sync"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

func dbPool() *pgxpool.Pool {
	once.Do(func() {
		var err error
		ctx := context.Background()
		config, err := pgxpool.ParseConfig(conf.Database)
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

func Query(query string) {
	pool := dbPool()
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			log.Fatalf("Scan failed: %v\n", err)
		}
	}
}

func Exec(query string) {
	pool := dbPool()
	_, err := pool.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("Exec failed: %v\n", err)
	}
}
