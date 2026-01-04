package cmd

import (
	"context"
	"fmt"

	"s3t/internal/s3tables"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [table-bucket] [namespace] [table]",
	Short: "List S3 Tables resources",
	Long: `List S3 Tables resources (Table Bucket, Namespace, Table) hierarchically.

Arguments are optional:
  - No arguments: List all table buckets, then interactively navigate
  - table-bucket: List namespaces in the specified bucket
  - table-bucket namespace: List tables in the specified namespace
  - table-bucket namespace table: Show table details

Interactive Features:
  - Type to filter: Press "/" then type to filter resources in real-time
  - .. (Back): Select this option to go back to previous level
  - Ctrl+C: Exit the application
  - Enter: Select the highlighted resource

Examples:
  # List all table buckets interactively
  s3t list

  # Start from a specific bucket
  s3t list my-bucket

  # List tables in a specific namespace
  s3t list my-bucket my-namespace

  # Show details of a specific table
  s3t list my-bucket my-namespace my-table`,
	Args: cobra.MaximumNArgs(3),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	client := getS3TablesClient()
	if client == nil {
		return fmt.Errorf("S3 Tables client not initialized")
	}

	ctx := context.Background()
	lister := s3tables.NewS3TablesLister(client)
	selector := s3tables.NewFilterablePromptSelector()
	controller := s3tables.NewNavigationController(lister, selector)

	switch len(args) {
	case 0:
		// Start from Table Bucket level
		return controller.Navigate(ctx, s3tables.LevelTableBucket)
	case 1:
		// Start from Namespace level with specified bucket
		bucketARN, err := lister.GetTableBucketARN(ctx, args[0])
		if err != nil {
			return err
		}
		controller.SetInitialState(args[0], bucketARN, "")
		return controller.Navigate(ctx, s3tables.LevelNamespace)
	case 2:
		// Start from Table level with specified bucket and namespace
		bucketARN, err := lister.GetTableBucketARN(ctx, args[0])
		if err != nil {
			return err
		}
		controller.SetInitialState(args[0], bucketARN, args[1])
		return controller.Navigate(ctx, s3tables.LevelTable)
	case 3:
		// Show table details directly
		return showTableDetails(ctx, lister, args[0], args[1], args[2])
	default:
		return fmt.Errorf("too many arguments")
	}
}

// showTableDetails displays detailed information about a specific table
func showTableDetails(ctx context.Context, lister *s3tables.S3TablesLister, tableBucketName, namespace, tableName string) error {
	// Get table bucket ARN
	tableBucketARN, err := lister.GetTableBucketARN(ctx, tableBucketName)
	if err != nil {
		return err
	}

	table, err := lister.GetTableDetails(ctx, tableBucketARN, namespace, tableName)
	if err != nil {
		return err
	}

	// Display table details
	fmt.Printf("\nTable Details:\n")
	fmt.Println()
	fmt.Printf("  Name:      %s\n", table.Name)
	fmt.Printf("  Namespace: %s\n", table.Namespace)
	fmt.Printf("  ARN:       %s\n", table.ARN)
	fmt.Printf("  Type:      %s\n", table.Type)
	fmt.Printf("  Created:   %s\n", table.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	return nil
}
