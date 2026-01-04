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

	// Global flags for AWS configuration
	awsProfile string
	awsRegion  string
)

var rootCmd = &cobra.Command{
	Use:   "s3t",
	Short: "S3 Tables CLI - Create and manage S3 Tables resources",
	Long: `A CLI tool for creating Amazon S3 Tables resources (Table Bucket, Namespace, Table) hierarchically.

This tool uses the default AWS credential chain for authentication.
Configure your credentials using:
  - AWS CLI: aws configure
  - Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
  - IAM roles (for EC2/ECS/Lambda)

Global Options:
  --profile  Use a specific AWS profile from ~/.aws/credentials or ~/.aws/config
  --region   Override the AWS region for API calls

Examples:
  # Use default credentials and region
  s3t create my-bucket my-namespace my-table

  # Use a specific AWS profile
  s3t --profile my-profile create my-bucket my-namespace my-table

  # Use a specific region
  s3t --region ap-northeast-1 create my-bucket my-namespace my-table

  # Combine profile and region
  s3t --profile my-profile --region us-west-2 create my-bucket my-namespace my-table`,
	PersistentPreRunE: initAWSClient,
	SilenceUsage:      true,
	SilenceErrors:     true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// buildConfigOptions creates config options based on profile and region flags.
// Returns a slice of config.LoadOptions functions to be passed to config.LoadDefaultConfig.
func buildConfigOptions(profile, region string) []func(*config.LoadOptions) error {
	var opts []func(*config.LoadOptions) error

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	return opts
}

// handleConfigError wraps AWS configuration errors with user-friendly messages.
// When a profile is specified, it returns a profile-specific error message.
// Otherwise, it returns a general configuration error message with guidance.
func handleConfigError(err error, profile string) error {
	if profile != "" {
		return fmt.Errorf("failed to load AWS profile '%s': %w\n\nPlease ensure the profile exists in ~/.aws/credentials or ~/.aws/config", profile, err)
	}
	return fmt.Errorf("failed to load AWS configuration: %w\n\nPlease configure AWS credentials using:\n  - AWS CLI: aws configure\n  - Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY\n  - IAM roles (for EC2/ECS/Lambda)", err)
}

// initAWSClient initializes the AWS S3 Tables client using the default credential chain
func initAWSClient(cmd *cobra.Command, args []string) error {
	// Skip client initialization for help commands
	if cmd.Name() == "help" || cmd.Name() == "completion" {
		return nil
	}

	ctx := context.Background()

	// Build config options based on flags
	configOpts := buildConfigOptions(awsProfile, awsRegion)

	// Load AWS configuration using default credential chain with options
	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return handleConfigError(err, awsProfile)
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
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&awsProfile, "profile", "", "AWS profile name to use for authentication")
	rootCmd.PersistentFlags().StringVar(&awsRegion, "region", "", "AWS region to use for API calls")

	// Add version flag
	rootCmd.Version = "0.1.0"

	// Customize error output
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetOut(os.Stdout)
}
