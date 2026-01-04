package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/spf13/cobra"

	s3tablesinternal "s3t/internal/s3tables"
)

var (
	// s3tablesClient holds the initialized S3 Tables client
	s3tablesClient s3tablesinternal.S3TablesAPI
)

var rootCmd = &cobra.Command{
	Use:   "s3t",
	Short: "S3 Tables CLI - Create and manage S3 Tables resources",
	Long: `A CLI tool for creating Amazon S3 Tables resources (Table Bucket, Namespace, Table) hierarchically.

This tool uses the default AWS credential chain for authentication.
Configure your credentials using:
  - AWS CLI: aws configure
  - Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
  - IAM roles (for EC2/ECS/Lambda)`,
	PersistentPreRunE: initAWSClient,
	SilenceUsage:      true,
	SilenceErrors:     true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// initAWSClient initializes the AWS S3 Tables client using the default credential chain
func initAWSClient(cmd *cobra.Command, args []string) error {
	// Skip client initialization for help commands
	if cmd.Name() == "help" || cmd.Name() == "completion" {
		return nil
	}

	ctx := context.Background()

	// Load AWS configuration using default credential chain
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w\n\nPlease configure AWS credentials using:\n  - AWS CLI: aws configure\n  - Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY\n  - IAM roles (for EC2/ECS/Lambda)", err)
	}

	// Create S3 Tables client
	s3tablesClient = s3tables.NewFromConfig(cfg)

	return nil
}

// getS3TablesClient returns the initialized S3 Tables client
func getS3TablesClient() s3tablesinternal.S3TablesAPI {
	return s3tablesClient
}

// SetS3TablesClient sets the S3 Tables client (useful for testing)
func SetS3TablesClient(client s3tablesinternal.S3TablesAPI) {
	s3tablesClient = client
}

func init() {
	// Add version flag
	rootCmd.Version = "0.1.0"

	// Customize error output
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetOut(os.Stdout)
}
