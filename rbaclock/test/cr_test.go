package test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"rbaclock/conf"
	"rbaclock/pkg/crchecker"
	"strings"
	"testing"
)

func TestCRChecker(t *testing.T) {
	fmt.Println("TestCRChecker")
	path := "/home/ch4ser/Projects/rbaclock/experiment/source"
	subDirs := listSubDirs(path)
	for _, subDir := range subDirs {
		database := crchecker.GenerateDB(subDir)
		parts := strings.Split(subDir, "/")
		repo := parts[len(parts)-1]
		crchecker.Check(repo, database, conf.CheckerQL)
	}
}

func listSubDirs(root string) []string {
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	var subDirs []string
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() && !strings.Contains(dirEntry.Name(), ".") {
			subDirs = append(subDirs, filepath.Join(root, dirEntry.Name()))
		}
	}
	return subDirs
}
