/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	commonlib "github.com/planesailingio/gotfhub/lib"
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:     "pull",
	Aliases: []string{"get", "download"},
	Short:   "Pull down specific Terraform providers",
	Long:    `This is used to pull down providers`,
	Run: func(cmd *cobra.Command, args []string) {
		providerPath := filepath.Join(outputDir, "providers")
		modulePath := filepath.Join(outputDir, "modules")

		if terraformArtifactNamespace != "" {
			switch terraformArtefactType {
			case "provider": 
				item := commonlib.Provider{
					Namespace: terraformArtifactNamespace,
				}
				if terraformArtifactVersion != "" {
					item.Version = terraformArtifactVersion
				}
				if err := commonlib.ProcessProvider(item, artefactCount, providerPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing provider %s: %v\n", item.Namespace, err)
				}
			case "module":
				item := commonlib.Module{
					Namespace: terraformArtifactNamespace,
				}
				if terraformArtifactVersion != "" {
					item.Version = terraformArtifactVersion
				}

				if err := commonlib.ProcessModule(item, artefactCount, modulePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing module %s: %v\n", item.Namespace, err)
				}
			}


		} else {
			providers, err := commonlib.LoadProviders(yamlFile)
			modules, err := commonlib.LoadModules(yamlFile)
			if err != nil {
				fmt.Printf("Failed to load providers: %s\n", err)
				return
			}

			workCh := make(chan interface{}) // Channel for both providers and modules
			var wg sync.WaitGroup

			// Start consumer goroutines
			for i := 0; i < parallelForks; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for work := range workCh {
						switch item := work.(type) {
						case commonlib.Provider:
							if err := commonlib.ProcessProvider(item, artefactCount, providerPath); err != nil {
								fmt.Fprintf(os.Stderr, "Error processing provider %s: %v\n", item.Namespace, err)
							}
						case commonlib.Module:
							if err := commonlib.ProcessModule(item, artefactCount, modulePath); err != nil {
								fmt.Fprintf(os.Stderr, "Error processing module %s: %v\n", item.Namespace, err)
							}
						}
					}
				}()
			}

			// Produce work for providers
			for _, provider := range providers {
				workCh <- provider
			}

			// Produce work for modules
			for _, module := range modules {
				workCh <- module
			}

			close(workCh) // Close the channel after all work is queued
			wg.Wait()     // Wait for all consumers to finish
			return
		} 
	},
}


func init() {
	pullCmd.Flags().StringVarP(&yamlFile, "artefact-path", "p", "./artefacts.yaml", "Path to a artefacts.yaml (optional)")
	pullCmd.Flags().StringVarP(&outputDir, "output-path", "o", outputDir, "Directory to save output files (optional)")
	pullCmd.Flags().IntVarP(&artefactCount, "count", "c", artefactCountDefault, "Number of versions to fetch per provider and modules")
	pullCmd.Flags().IntVarP(&parallelForks, "forks", "f", runtime.NumCPU(), "Number of parallel forks to fetch providers and modules")
	pullCmd.Flags().StringVarP(&terraformArtifactNamespace, "namespace", "n", "", "Number of parallel forks to fetch providers and modules")
	pullCmd.Flags().StringVarP(&terraformArtifactVersion, "version", "v", "", "Version of the artefact type (used with namespace)")
	pullCmd.Flags().StringVarP(&terraformArtefactType, "type", "t", "provider", "Number of parallel forks to fetch providers and modules")

	rootCmd.AddCommand(pullCmd)
}
