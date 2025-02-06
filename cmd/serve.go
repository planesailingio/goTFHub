/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	commonlib "github.com/planesailingio/gotfhub/lib"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the provider and module Terraform artefacts",
	Long:  `Serve the provider and module Terraform artefacts`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
		commonlib.Serve(backend, bucket, outputDir, "")
	},
}

func init() {

	serveCmd.Flags().StringVar(&backend, "backend", "aws", "A backend to use. Currently supports aws and minio (Optional)")
	serveCmd.Flags().StringVar(&bucket, "bucket", "", "Desired S3 bucket to store Terraform artefacts (Required)")
	// serveCmd.Flags().StringVar(&bucket, "bucket", "", "Desired S3 bucket to store Terraform artefacts (Required)")
	serveCmd.MarkFlagRequired("bucket")
	rootCmd.AddCommand(serveCmd)
}
