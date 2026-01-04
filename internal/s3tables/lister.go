package s3tables

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
)

// TableBucketInfo represents a table bucket with its metadata
type TableBucketInfo struct {
	Name      string
	ARN       string
	CreatedAt time.Time
}

// NamespaceInfo represents a namespace with its metadata
type NamespaceInfo struct {
	Name      string
	CreatedAt time.Time
}

// TableInfo represents a table with its metadata
type TableInfo struct {
	Name      string
	ARN       string
	Namespace string
	CreatedAt time.Time
	Type      string
}

// S3TablesLister manages S3 Tables resource listing
type S3TablesLister struct {
	client S3TablesAPI
}

// NewS3TablesLister creates a new S3TablesLister instance
func NewS3TablesLister(client S3TablesAPI) *S3TablesLister {
	return &S3TablesLister{client: client}
}

// ListTableBucketsAll retrieves all table buckets with pagination
func (l *S3TablesLister) ListTableBucketsAll(ctx context.Context, prefix string) ([]TableBucketInfo, error) {
	var buckets []TableBucketInfo
	var continuationToken *string

	for {
		input := &s3tables.ListTableBucketsInput{
			ContinuationToken: continuationToken,
		}
		if prefix != "" {
			input.Prefix = aws.String(prefix)
		}

		output, err := l.client.ListTableBuckets(ctx, input)
		if err != nil {
			return nil, WrapError("ListTableBuckets", err)
		}

		for _, bucket := range output.TableBuckets {
			buckets = append(buckets, TableBucketInfo{
				Name:      aws.ToString(bucket.Name),
				ARN:       aws.ToString(bucket.Arn),
				CreatedAt: aws.ToTime(bucket.CreatedAt),
			})
		}

		if output.ContinuationToken == nil || *output.ContinuationToken == "" {
			break
		}
		continuationToken = output.ContinuationToken
	}

	return buckets, nil
}

// ListNamespacesAll retrieves all namespaces in a table bucket with pagination
func (l *S3TablesLister) ListNamespacesAll(ctx context.Context, tableBucketARN, prefix string) ([]NamespaceInfo, error) {
	var namespaces []NamespaceInfo
	var continuationToken *string

	for {
		input := &s3tables.ListNamespacesInput{
			TableBucketARN:    aws.String(tableBucketARN),
			ContinuationToken: continuationToken,
		}
		if prefix != "" {
			input.Prefix = aws.String(prefix)
		}

		output, err := l.client.ListNamespaces(ctx, input)
		if err != nil {
			return nil, WrapError("ListNamespaces", err)
		}

		for _, ns := range output.Namespaces {
			var name string
			if len(ns.Namespace) > 0 {
				name = ns.Namespace[0]
			}
			namespaces = append(namespaces, NamespaceInfo{
				Name:      name,
				CreatedAt: aws.ToTime(ns.CreatedAt),
			})
		}

		if output.ContinuationToken == nil || *output.ContinuationToken == "" {
			break
		}
		continuationToken = output.ContinuationToken
	}

	return namespaces, nil
}

// ListTablesAll retrieves all tables in a namespace with pagination
func (l *S3TablesLister) ListTablesAll(ctx context.Context, tableBucketARN, namespace, prefix string) ([]TableInfo, error) {
	var tables []TableInfo
	var continuationToken *string

	for {
		input := &s3tables.ListTablesInput{
			TableBucketARN:    aws.String(tableBucketARN),
			Namespace:         aws.String(namespace),
			ContinuationToken: continuationToken,
		}
		if prefix != "" {
			input.Prefix = aws.String(prefix)
		}

		output, err := l.client.ListTables(ctx, input)
		if err != nil {
			return nil, WrapError("ListTables", err)
		}

		for _, tbl := range output.Tables {
			var ns string
			if len(tbl.Namespace) > 0 {
				ns = tbl.Namespace[0]
			}
			tables = append(tables, TableInfo{
				Name:      aws.ToString(tbl.Name),
				ARN:       aws.ToString(tbl.TableARN),
				Namespace: ns,
				CreatedAt: aws.ToTime(tbl.CreatedAt),
				Type:      string(tbl.Type),
			})
		}

		if output.ContinuationToken == nil || *output.ContinuationToken == "" {
			break
		}
		continuationToken = output.ContinuationToken
	}

	return tables, nil
}

// GetTableDetails retrieves detailed information about a specific table
func (l *S3TablesLister) GetTableDetails(ctx context.Context, tableBucketARN, namespace, table string) (*TableInfo, error) {
	input := &s3tables.GetTableInput{
		TableBucketARN: aws.String(tableBucketARN),
		Namespace:      aws.String(namespace),
		Name:           aws.String(table),
	}

	output, err := l.client.GetTable(ctx, input)
	if err != nil {
		return nil, WrapError("GetTable", err)
	}

	return &TableInfo{
		Name:      aws.ToString(output.Name),
		ARN:       aws.ToString(output.TableARN),
		Namespace: namespace,
		CreatedAt: aws.ToTime(output.CreatedAt),
		Type:      string(output.Type),
	}, nil
}

// GetTableBucketARN retrieves the ARN for a table bucket by name
func (l *S3TablesLister) GetTableBucketARN(ctx context.Context, tableBucketName string) (string, error) {
	buckets, err := l.ListTableBucketsAll(ctx, tableBucketName)
	if err != nil {
		return "", err
	}

	for _, bucket := range buckets {
		if bucket.Name == tableBucketName {
			return bucket.ARN, nil
		}
	}

	return "", &S3TablesError{
		Operation:  "GetTableBucketARN",
		Message:    fmt.Sprintf("table bucket '%s' not found", tableBucketName),
		Suggestion: "verify the table bucket name and try again",
		Type:       ErrorTypeNotFound,
	}
}
