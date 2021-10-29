/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/helm"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version of BigBang Cluster and BigBang CLI.",
	Long:  `Print version of BigBang Cluster and BigBang CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		clientVersionOnly, _ := cmd.Flags().GetBool("client")
		fmt.Println("bbctl version 0.0.1")
		if !clientVersionOnly {
			bbChartVersion()
		}

	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("client", "c", false, "Print bbctl version only")
}

func bbChartVersion() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	opt := &helm.Options{
		Namespace:        "bigbang",
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
		RestConfig:       config,
	}

	helmClient, err := helm.New(opt)
	if err != nil {
		panic(err)
	}

	release, _ := helmClient.GetRelease("bigbang")
	fmt.Printf("%s version %s\n", release.Chart.Metadata.Name, release.Chart.Metadata.Version)

	releases, _ := helmClient.GetList()
	_ = releases
}
