package cmd

import (
	"context"
	"fmt"

	"s3t/internal/s3tables"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <table-bucket> <namespace> <table>",
	Short: "Create S3 Tables resources",
	Long: `Create S3 Tables resources (Table Bucket, Namespace, Table) hierarchically.

This command creates the specified resources in order:
1. Table Bucket (if not exists)
2. Namespace (if not exists)
3. Table (if not exists)

Existing resources are detected and skipped with a notification.`,
	Args: cobra.ExactArgs(3),
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	tableBucket := args[0]
	namespace := args[1]
	table := args[2]

	// Validate input arguments
	if err := s3tables.ValidateAll(tableBucket, namespace, table); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Get the S3 Tables client from context (set by root command)
	client := getS3TablesClient()
	if client == nil {
		return fmt.Errorf("S3 Tables client not initialized")
	}

	// Create the S3TablesCreator and execute
	creator := s3tables.NewS3TablesCreator(client)
	ctx := context.Background()

	result, err := creator.Create(ctx, tableBucket, namespace, table)
	if err != nil {
		return err
	}

	// Output results
	printResult(result)
	return nil
}

// printResult outputs the creation result in a user-friendly format
func printResult(result *s3tables.CreateResult) {
	fmt.Println()
	fmt.Println("=== S3 Tables Resource Creation Summary ===")
	fmt.Println()

	// Print each message
	for _, msg := range result.Messages {
		fmt.Printf("  • %s\n", msg)
	}

	fmt.Println()

	// Print summary
	created := 0
	existed := 0

	if result.TableBucketCreated {
		created++
	} else if result.TableBucketARN != "" {
		existed++
	}

	if result.NamespaceCreated {
		created++
	} else {
		existed++
	}

	if result.TableCreated {
		created++
	} else if result.TableARN != "" {
		existed++
	}

	if created > 0 {
		fmt.Printf("Created: %d resource(s)\n", created)
	}
	if existed > 0 {
		fmt.Printf("Already existed: %d resource(s)\n", existed)
	}

	// Print ARNs if available
	if result.TableBucketARN != "" {
		fmt.Printf("\nTable Bucket ARN: %s\n", result.TableBucketARN)
	}
	if result.TableARN != "" {
		fmt.Printf("Table ARN: %s\n", result.TableARN)
	}
}
