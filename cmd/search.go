/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	commonlib "github.com/planesailingio/gotfhub/lib"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to search the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var versions []string
		if terraformArtefactType == "provider" {
			versions, _ = commonlib.FetchProviderVersions(terraformArtifactNamespace)
			} else if terraformArtefactType == "module" {
			versions, _ = commonlib.FetchModuleVersions(terraformArtifactNamespace)
		}
		if artefactCount > len(versions) {
			artefactCount = len(versions)
		}
		for _,version := range versions[:artefactCount] {
			uri := fmt.Sprintf("%s:%s",terraformArtifactNamespace,version)
			fmt.Println(uri)
		}
	},
}

func init() {
	artefactCountDefault = 3
	searchCmd.Flags().StringVarP(&terraformArtifactNamespace, "namespace", "n", "", "Terraform artefact namespace e.g. hashicorp/tls (required)")
	searchCmd.Flags().IntVarP(&artefactCount, "count", "c", artefactCountDefault, "Number of versions to return (optional, default: 3)")
	searchCmd.Flags().StringVarP(&terraformArtefactType, "type", "t", "provider", "Terraform Artefact type either provider or module (default: provider)")
	searchCmd.MarkFlagRequired("namespace")
	rootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
