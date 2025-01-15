/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	commonlib "github.com/planesailingio/gotfhub/lib"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate an example artefacts.yaml file",
	Long:  `Quickly generate a artefacts.yaml with a bunch of example on-premise services to get you started.`,
	Run:   runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

var outputFile string

func init() {
	initCmd.Flags().StringVarP(&outputFile, "output", "o", "artefacts.yaml", "Path to output artefacts.yaml file")
}

func runInit(cmd *cobra.Command, args []string) {
	example := commonlib.ArtefactList{
		Providers: []commonlib.Provider{
			{Namespace: "coder/coder"},
			{Namespace: "coder/coderd"},
			{Namespace: "cyrilgdn/postgresql"},
			{Namespace: "cyrilgdn/rabbitmq"},
			{Namespace: "datadrivers/nexus"},
			{Namespace: "drfaust92/confluence"},
			{Namespace: "elastic/elasticstack"},
			{Namespace: "fortinetdev/fortios"},
			{Namespace: "fourplusone/jira"},
			{Namespace: "gavinbunney/kubectl"},
			{Namespace: "gitlabhq/gitlab", Version: "3.9.1"},
			{Namespace: "gitlabhq/gitlab"},
			{Namespace: "goharbor/harbor"},
			{Namespace: "grafana/grafana"},
			{Namespace: "hashicorp/archive"},
			{Namespace: "hashicorp/assert"},
			{Namespace: "hashicorp/aws", Version: "5.81.0"},
			{Namespace: "hashicorp/cloudinit"},
			{Namespace: "hashicorp/external"},
			{Namespace: "hashicorp/helm"},
			{Namespace: "hashicorp/http"},
			{Namespace: "hashicorp/kubernetes"},
			{Namespace: "hashicorp/local"},
			{Namespace: "hashicorp/null"},
			{Namespace: "hashicorp/random"},
			{Namespace: "hashicorp/template"},
			{Namespace: "hashicorp/time"},
			{Namespace: "hashicorp/tls"},
			{Namespace: "hashicorp/vault"},
			{Namespace: "hashicorp/vsphere"},
			{Namespace: "mrparkers/keycloak"},
			{Namespace: "mrparkers/keycloak"},
			{Namespace: "opensearch-project/opensearch"},
			{Namespace: "rancher/rancher2"},
		},
		Modules: []commonlib.Module{
			{
				Namespace: "terraform-aws-modules/s3-bucket/aws",
				Version:   "4.2.2",
			},
		},
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Errorf("failed to create file: %w", err)
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(example); err != nil {
		fmt.Errorf("failed to write YAML: %w", err)
		return
	}

	fmt.Printf("Example %s created at: ./%s\n", outputFile, outputFile)
	return
}
