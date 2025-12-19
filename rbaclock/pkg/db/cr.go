package db

import (
	"context"
	"log"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/jackc/pgx/pgtype"
)

func init() {
	pool := dbPool()
	pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS cr_map (cr TEXT PRIMARY KEY, repo TEXT, builtin TEXT[])")
}

func UpsertCRMap(cr, repo string, builtins []string) {
	pool := dbPool()
	lowercaseBuiltins := []string{}
	for _, b := range builtins {
		lowercaseBuiltins = append(lowercaseBuiltins, LowerPlural(b))
	}
	query := `INSERT INTO cr_map (cr, repo, builtin) VALUES ($1, $2, $3) ON CONFLICT (cr) DO UPDATE SET repo = EXCLUDED.repo, builtin = EXCLUDED.builtin`

	_, err := pool.Exec(context.Background(), query, LowerPlural(cr), repo, lowercaseBuiltins)
	if err != nil {
		log.Fatalf("UpsertCRMap failed: %v\n", err)
	}
}

func QueryCRMap(cr string) []string {
	pool := dbPool()
	query := `SELECT builtin FROM cr_map WHERE cr = $1`
	var arr pgtype.TextArray
	err := pool.QueryRow(context.Background(), query, cr).Scan(&arr)
	if err != nil {
		return []string{cr}
	}
	var res []string
	for _, s := range arr.Elements {
		res = append(res, s.String)
	}
	return res
}

func LowerPlural(cr string) string {
	cr = strings.ToLower(cr)
	pl := pluralize.NewClient()
	return pl.Plural(cr)
}
