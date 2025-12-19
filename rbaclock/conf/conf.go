package conf

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// var Kubeconfig = "./conf/kubeconfig_tencent"
// var SubClusterSize = 3
// var IgnoreNamespaces = []string{
// 	// "kube-system",
// 	// "kube-public",
// 	// "kube-node-lease",
// }
// var IgnorePods = []string{
// 	"kube-proxy",
// 	"kubernetes-proxy",
// 	"kindnet",
// }
// var GlobalKey = "rbaclock"
// var Database = "postgres://user:pass@localhost:5432/rbaclock"

type WeightConfig struct {
	Container float32 `yaml:"container"`
	Node      float32 `yaml:"node"`
}

var (
	Kubeconfig       string
	Deploy           bool
	IgnoreNamespaces []string
	IgnorePods       []string
	GlobalKey        string
	Database         string
	Weight           WeightConfig
	CheckerQL        string
	SourceCode       string
	ClusterScript    string
)

func init() {
	pwd, _ := os.Getwd()
	CdRootDir(pwd)
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	var cfg struct {
		Kubeconfig       string       `yaml:"kubeconfig"`
		SubClusterSizes  []int        `yaml:"subClusterSizes"`
		Deploy           bool         `yaml:"deploy"`
		IgnoreNamespaces []string     `yaml:"ignoreNamespaces"`
		IgnorePods       []string     `yaml:"ignorePods"`
		GlobalKey        string       `yaml:"globalKey"`
		Database         string       `yaml:"database"`
		Weight           WeightConfig `yaml:"weight"`
		CheckerQL        string       `yaml:"checkerQL"`
		SourceCode       string       `yaml:"sourceCode"`
		ClusterScript    string       `yaml:"clusterScript"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	Kubeconfig = cfg.Kubeconfig
	Deploy = cfg.Deploy
	IgnoreNamespaces = cfg.IgnoreNamespaces
	IgnorePods = cfg.IgnorePods
	GlobalKey = cfg.GlobalKey
	Database = cfg.Database
	CheckerQL = cfg.CheckerQL
	SourceCode = cfg.SourceCode
	Weight = cfg.Weight
	ClusterScript = cfg.ClusterScript
}

func CdRootDir(path string) {
	// check if there is .git folder
	// if not, cd to the parent dir
	// if yes, cd to the root dir
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		if path == "/" {
			panic("not a git repo")
		}
		CdRootDir(filepath.Dir(path))
	} else {
		err := os.Chdir(path)
		if err != nil {
			panic(err)
		}
	}
}
