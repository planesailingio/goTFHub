/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	commonlib "github.com/planesailingio/gotfhub/lib"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the transferred Terraform Providers and Module artefacts",
	Long:  `Push the transferred Terraform Providers and Module artefacts`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := commonlib.UploadDirectory(backend, bucket, outputDir, prefix); err != nil {
			fmt.Println(err.Error())
		}

	},
}

func init() {
	pushCmd.Flags().StringVarP(&outputDir, "local-path", "p", outputDir, "Directory to save output files (Optional)")
	pushCmd.Flags().StringVar(&bucket, "bucket", "", "Desired S3 bucket to store Terraform artefacts (Required)")
	pushCmd.Flags().StringVar(&backend, "backend", "", "Desired S3 bucket to store Terraform artefacts (Optional)")
	pushCmd.Flags().StringVar(&prefix, "prefix", prefix, "Desired S3 bucket key prefix to store Terraform artefacts (Optional)")
	pushCmd.MarkFlagRequired("bucket")
	// serveCmd.MarkFlagRequired("bucket")
	rootCmd.AddCommand(pushCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
