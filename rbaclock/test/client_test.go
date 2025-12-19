package test

import (
	"os"
	"path/filepath"
	cli "rbaclock/pkg/client"
	"strings"
	"testing"
)

func TestAddAffinity(t *testing.T) {
	pods := cli.Pods("gitlab")
	for _, pod := range pods {
		if len(pod.OwnerReferences) > 0 {
			if strings.Contains(pod.OwnerReferences[0].Name, "my-gitlab-gitlab-shell") {
				t.Log("Find gitlab shell")
				owner := cli.Owner(pod)
				// cli.AddOwnerAffinity(owner, "keyy", "valuee")
				t.Log(owner)
			}
		}
	}
}

func init() {
	rootDir, _ := os.Getwd()
	rootDir = filepath.Dir(rootDir)
	_ = os.Chdir(rootDir)
}
