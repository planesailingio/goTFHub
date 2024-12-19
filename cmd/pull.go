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
		providers, err := commonlib.LoadProviders(yamlFile)
		modules, err := commonlib.LoadModules(yamlFile)
		if err != nil {
			fmt.Printf("Failed to load providers: %s\n", err)
			return
		}

		var wg sync.WaitGroup
		if parallelForks == 0 {
			parallelForks = 1
		}
		sem := make(chan struct{}, parallelForks) // Control concurrency

		for _, provider := range providers {
			wg.Add(1)
			sem <- struct{}{}
			go func(p commonlib.Provider) {
				defer wg.Done()
				providerPath := filepath.Join(outputDir, "providers")
				if err := commonlib.ProcessProvider(p, artefactCount, providerPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing provider %s: %v\n", p.Namespace, err)
				}
				<-sem
			}(provider)
		}

		for _, module := range modules {
			wg.Add(1)
			sem <- struct{}{}
			go func(m commonlib.Module) {
				defer wg.Done()
				modulePath := filepath.Join(outputDir, "modules")
				if err := commonlib.ProcessModule(m, artefactCount, modulePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing provider %s: %v\n", m.Namespace, err)
				}
				<-sem
			}(module)
		}

		wg.Wait()
		return
	},
}
var parallelForks int

func init() {
	pullCmd.Flags().StringVarP(&yamlFile, "artefact-path", "p", "./artefacts.yaml", "Path to a artefacts.yaml (optional)")
	pullCmd.Flags().StringVarP(&outputDir, "output-path", "o", outputDir, "Directory to save output files (optional)")
	pullCmd.Flags().IntVarP(&artefactCount, "count", "c", 5, "Number of versions to fetch per provider and modules")
	pullCmd.Flags().IntVarP(&parallelForks, "forks", "f", runtime.NumCPU(), "Number of parallel forks to fetch providers and modules")

	rootCmd.AddCommand(pullCmd)
}
