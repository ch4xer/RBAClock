package cmd

import (
	"log"
	"os"
	"path/filepath"
	"rbaclock/conf"
	"rbaclock/pkg/crchecker"
	"strings"

	"github.com/spf13/cobra"
)

var crCheckCmd = &cobra.Command{
	Use:   "crcheck",
	Short: "Find the Mapping between Custom Resources and Built-in Resources across all projects within the specified path",
	Run:   runCrCheck,
}

func runCrCheck(cmd *cobra.Command, args []string) {
	subDirs := listSubDirs(conf.SourceCode)
	log.Println(subDirs)
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

func init() {
	rootCmd.AddCommand(crCheckCmd)
}
