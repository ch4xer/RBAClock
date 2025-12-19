package crchecker

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"rbaclock/pkg/db"
	"rbaclock/pkg/privilege"
	"strings"
)

func GenerateDB(repo string) string {
	codeqlDB := fmt.Sprintf("%s/codeql_database", repo)
	cmd := exec.Command("codeql", "database", "create", codeqlDB, "--language=go", "--source-root", repo)
	var err bytes.Buffer
	cmd.Stderr = &err

	if e := cmd.Run(); e != nil {
		if !strings.Contains(err.String(), "exists and is not an empty directory") {
			log.Fatal(err.String())
		}
	}
	log.Printf("Database for %s generated\n", repo)
	return codeqlDB
}

func Check(repo, database, query string) {
	outputPath := "/tmp/output.bqrs"
	cmdSeg1 := fmt.Sprintf("--output=%s", outputPath)
	cmdSeg2 := fmt.Sprintf("--database=%s", database)

	cmd := exec.Command("codeql", "query", "run", cmdSeg2, query, cmdSeg1)
	var err bytes.Buffer
	cmd.Stderr = &err

	if e := cmd.Run(); e != nil {
		log.Fatal(err.String())
	}

	log.Printf("Query %s executed\n", query)
	crMap := Decode(outputPath)
	for cr, builtins := range crMap {
		db.UpsertCRMap(cr, repo, builtins)
	}
}

func Decode(output string) map[string][]string {
	tmpOutput := "/tmp/output.csv"
	cmd := exec.Command("codeql", "bqrs", "decode", "--format=csv", output)
	var err bytes.Buffer
	var out bytes.Buffer
	cmd.Stderr = &err
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		log.Fatal(err.String())
	}

	if e := os.WriteFile(tmpOutput, out.Bytes(), 0655); e != nil {
		log.Fatal(e)
	}

	f, _ := os.Open(tmpOutput)
	defer f.Close()

	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	CRMap := make(map[string][]string)
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		if record[0] == "col0" {
			continue
		}

		cr := extractResource(record[0])
		builtin := extractResource(record[1])
		if CRMap[cr] == nil {
			CRMap[cr] = []string{builtin}
		} else {
			CRMap[cr] = append(CRMap[cr], builtin)
		}
	}

	resultMap := make(map[string][]string)
	for cr, builtins := range CRMap {
		// some cr have same name as builtin, skip
		if privilege.IsCriticalBuiltin(db.LowerPlural(cr)) {
			continue
		}
		resultMap[cr] = builtins
	}
	return resultMap
}

func extractResource(seg string) string {
	s0 := strings.Split(seg, "/")
	s1 := strings.Split(s0[len(s0)-1], ".")
	return s1[len(s1)-1]
}
