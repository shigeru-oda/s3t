package s3tables

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/aws/smithy-go"
)

// S3TablesAPI defines the interface for AWS S3 Tables API operations
type S3TablesAPI interface {
	ListTableBuckets(ctx context.Context, params *s3tables.ListTableBucketsInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTableBucketsOutput, error)
	GetTableBucket(ctx context.Context, params *s3tables.GetTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableBucketOutput, error)
	CreateTableBucket(ctx context.Context, params *s3tables.CreateTableBucketInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableBucketOutput, error)
	GetNamespace(ctx context.Context, params *s3tables.GetNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.GetNamespaceOutput, error)
	CreateNamespace(ctx context.Context, params *s3tables.CreateNamespaceInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateNamespaceOutput, error)
	GetTable(ctx context.Context, params *s3tables.GetTableInput, optFns ...func(*s3tables.Options)) (*s3tables.GetTableOutput, error)
	CreateTable(ctx context.Context, params *s3tables.CreateTableInput, optFns ...func(*s3tables.Options)) (*s3tables.CreateTableOutput, error)
	ListNamespaces(ctx context.Context, params *s3tables.ListNamespacesInput, optFns ...func(*s3tables.Options)) (*s3tables.ListNamespacesOutput, error)
	ListTables(ctx context.Context, params *s3tables.ListTablesInput, optFns ...func(*s3tables.Options)) (*s3tables.ListTablesOutput, error)
}

// CreateResult represents the result of resource creation
type CreateResult struct {
	TableBucketARN     string
	TableARN           string
	Messages           []string
	TableBucketCreated bool
	NamespaceCreated   bool
	TableCreated       bool
}

// S3TablesCreator manages S3 Tables resource creation
type S3TablesCreator struct {
	client S3TablesAPI
}

// NewS3TablesCreator creates a new S3TablesCreator instance
func NewS3TablesCreator(client S3TablesAPI) *S3TablesCreator {
	return &S3TablesCreator{client: client}
}

// isNotFoundError checks if the error is a NotFoundException from AWS API
func isNotFoundError(err error) bool {
	var nfe *types.NotFoundException
	if errors.As(err, &nfe) {
		return true
	}
	// Also check for smithy API error with NotFound code
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "NotFoundException"
	}
	return false
}

// checkTableBucketExists checks if a Table Bucket exists and returns its ARN if it does
// Uses ListTableBuckets with prefix filter to find the bucket by name
func (c *S3TablesCreator) checkTableBucketExists(ctx context.Context, tableBucket string) (exists bool, arn string, err error) {
	output, err := c.client.ListTableBuckets(ctx, &s3tables.ListTableBucketsInput{
		Prefix: aws.String(tableBucket),
	})
	if err != nil {
		return false, "", WrapError("ListTableBuckets", err)
	}

	// Find exact match in the results
	for _, bucket := range output.TableBuckets {
		if aws.ToString(bucket.Name) == tableBucket {
			return true, aws.ToString(bucket.Arn), nil
		}
	}

	return false, "", nil
}

// checkNamespaceExists checks if a Namespace exists under the given Table Bucket
func (c *S3TablesCreator) checkNamespaceExists(ctx context.Context, tableBucketARN, namespace string) (exists bool, err error) {
	_, err = c.client.GetNamespace(ctx, &s3tables.GetNamespaceInput{
		TableBucketARN: aws.String(tableBucketARN),
		Namespace:      aws.String(namespace),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, WrapError("GetNamespace", err)
	}
	return true, nil
}

// checkTableExists checks if a Table exists under the given Namespace
func (c *S3TablesCreator) checkTableExists(ctx context.Context, tableBucketARN, namespace, table string) (exists bool, tableARN string, err error) {
	output, err := c.client.GetTable(ctx, &s3tables.GetTableInput{
		TableBucketARN: aws.String(tableBucketARN),
		Namespace:      aws.String(namespace),
		Name:           aws.String(table),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, "", nil
		}
		return false, "", WrapError("GetTable", err)
	}
	return true, aws.ToString(output.TableARN), nil
}

// Create creates S3 Tables resources hierarchically: Table Bucket → Namespace → Table
// It checks for existing resources and only creates what's needed
func (c *S3TablesCreator) Create(ctx context.Context, tableBucket, namespace, table string) (*CreateResult, error) {
	result := &CreateResult{
		Messages: make([]string, 0),
	}

	// Step 1: Check/Create Table Bucket
	tableBucketARN, err := c.ensureTableBucket(ctx, tableBucket, result)
	if err != nil {
		return nil, err
	}

	// Step 2: Check/Create Namespace
	err = c.ensureNamespace(ctx, tableBucketARN, namespace, result)
	if err != nil {
		return nil, err
	}

	// Step 3: Check/Create Table
	err = c.ensureTable(ctx, tableBucketARN, namespace, table, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ensureTableBucket ensures the Table Bucket exists, creating it if necessary
func (c *S3TablesCreator) ensureTableBucket(ctx context.Context, tableBucket string, result *CreateResult) (string, error) {
	exists, arn, err := c.checkTableBucketExists(ctx, tableBucket)
	if err != nil {
		return "", err
	}

	if exists {
		result.TableBucketARN = arn
		result.Messages = append(result.Messages, fmt.Sprintf("Table Bucket '%s' already exists", tableBucket))
		return arn, nil
	}

	// Create Table Bucket
	output, err := c.client.CreateTableBucket(ctx, &s3tables.CreateTableBucketInput{
		Name: aws.String(tableBucket),
	})
	if err != nil {
		return "", WrapError("CreateTableBucket", err)
	}

	result.TableBucketCreated = true
	result.TableBucketARN = aws.ToString(output.Arn)
	result.Messages = append(result.Messages, fmt.Sprintf("Table Bucket '%s' created", tableBucket))
	return result.TableBucketARN, nil
}

// ensureNamespace ensures the Namespace exists, creating it if necessary
func (c *S3TablesCreator) ensureNamespace(ctx context.Context, tableBucketARN, namespace string, result *CreateResult) error {
	exists, err := c.checkNamespaceExists(ctx, tableBucketARN, namespace)
	if err != nil {
		return err
	}

	if exists {
		result.Messages = append(result.Messages, fmt.Sprintf("Namespace '%s' already exists", namespace))
		return nil
	}

	// Create Namespace
	_, err = c.client.CreateNamespace(ctx, &s3tables.CreateNamespaceInput{
		TableBucketARN: aws.String(tableBucketARN),
		Namespace:      []string{namespace},
	})
	if err != nil {
		return WrapError("CreateNamespace", err)
	}

	result.NamespaceCreated = true
	result.Messages = append(result.Messages, fmt.Sprintf("Namespace '%s' created", namespace))
	return nil
}

// ensureTable ensures the Table exists, creating it if necessary
func (c *S3TablesCreator) ensureTable(ctx context.Context, tableBucketARN, namespace, table string, result *CreateResult) error {
	exists, tableARN, err := c.checkTableExists(ctx, tableBucketARN, namespace, table)
	if err != nil {
		return err
	}

	if exists {
		result.TableARN = tableARN
		result.Messages = append(result.Messages, fmt.Sprintf("Table '%s' already exists", table))
		return nil
	}

	// Create Table
	output, err := c.client.CreateTable(ctx, &s3tables.CreateTableInput{
		TableBucketARN: aws.String(tableBucketARN),
		Namespace:      aws.String(namespace),
		Name:           aws.String(table),
		Format:         types.OpenTableFormatIceberg,
	})
	if err != nil {
		return WrapError("CreateTable", err)
	}

	result.TableCreated = true
	result.TableARN = aws.ToString(output.TableARN)
	result.Messages = append(result.Messages, fmt.Sprintf("Table '%s' created", table))
	return nil
}
