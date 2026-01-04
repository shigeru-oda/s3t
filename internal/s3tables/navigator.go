package s3tables

import (
	"context"
	"fmt"
)

// NavigationLevel represents the current navigation level
type NavigationLevel int

const (
	LevelTableBucket NavigationLevel = iota
	LevelNamespace
	LevelTable
)

// String returns the string representation of NavigationLevel
func (l NavigationLevel) String() string {
	switch l {
	case LevelTableBucket:
		return "TableBucket"
	case LevelNamespace:
		return "Namespace"
	case LevelTable:
		return "Table"
	default:
		return "Unknown"
	}
}

// NavigationAction represents the action taken by the user
type NavigationAction int

const (
	ActionSelect NavigationAction = iota // アイテムを選択
	ActionBack                           // ESC で戻る
	ActionExit                           // 終了
)

// String returns the string representation of NavigationAction
func (a NavigationAction) String() string {
	switch a {
	case ActionSelect:
		return "Select"
	case ActionBack:
		return "Back"
	case ActionExit:
		return "Exit"
	default:
		return "Unknown"
	}
}

// NavigationState holds the current navigation state and cached data
type NavigationState struct {
	Level             NavigationLevel
	TableBuckets      []TableBucketInfo // キャッシュされた Table Bucket 一覧
	Namespaces        []NamespaceInfo   // キャッシュされた Namespace 一覧
	Tables            []TableInfo       // キャッシュされた Table 一覧
	SelectedBucket    string            // 選択された Table Bucket 名
	SelectedBucketARN string            // 選択された Table Bucket ARN
	SelectedNamespace string            // 選択された Namespace 名
}

// NavigationController manages hierarchical navigation
type NavigationController struct {
	lister   *S3TablesLister
	selector InteractiveSelector
	state    *NavigationState
}

// NewNavigationController creates a new NavigationController
func NewNavigationController(lister *S3TablesLister, selector InteractiveSelector) *NavigationController {
	return &NavigationController{
		lister:   lister,
		selector: selector,
		state:    &NavigationState{},
	}
}

// GetState returns the current navigation state
func (c *NavigationController) GetState() *NavigationState {
	return c.state
}

// SetInitialState sets the initial state for navigation
func (c *NavigationController) SetInitialState(bucketName, bucketARN, namespace string) {
	c.state.SelectedBucket = bucketName
	c.state.SelectedBucketARN = bucketARN
	c.state.SelectedNamespace = namespace
}

// Navigate starts the navigation from the specified level
func (c *NavigationController) Navigate(ctx context.Context, startLevel NavigationLevel) error {
	c.state.Level = startLevel

	for {
		var action NavigationAction
		var err error

		switch c.state.Level {
		case LevelTableBucket:
			action, err = c.navigateTableBuckets(ctx)
			if err != nil {
				return err
			}
			if action == ActionBack || action == ActionExit {
				return nil // Exit application
			}
			if action == ActionSelect {
				c.state.Level = LevelNamespace
			}

		case LevelNamespace:
			action, err = c.navigateNamespaces(ctx)
			if err != nil {
				return err
			}
			if action == ActionExit {
				return nil // Exit application
			}
			if action == ActionBack {
				c.state.Level = LevelTableBucket
				continue
			}
			if action == ActionSelect {
				c.state.Level = LevelTable
			}

		case LevelTable:
			action, err = c.navigateTables(ctx)
			if err != nil {
				return err
			}
			if action == ActionExit {
				return nil // Exit application
			}
			if action == ActionBack {
				c.state.Level = LevelNamespace
				continue
			}
			// Table 選択後は詳細表示して終了
			return nil
		}
	}
}

// navigateTableBuckets handles Table Bucket level navigation
func (c *NavigationController) navigateTableBuckets(ctx context.Context) (NavigationAction, error) {
	// Fetch table buckets if not cached
	if c.state.TableBuckets == nil {
		buckets, err := c.lister.ListTableBucketsAll(ctx, "")
		if err != nil {
			return ActionExit, err
		}
		c.state.TableBuckets = buckets
	}

	if len(c.state.TableBuckets) == 0 {
		fmt.Println("No table buckets found")
		return ActionExit, nil
	}

	// Extract names for selection
	names := make([]string, len(c.state.TableBuckets))
	for i, bucket := range c.state.TableBuckets {
		names[i] = bucket.Name
	}

	// No back option at top level
	result, err := c.selector.SelectWithFilter("Select Table Bucket", names, false)
	if err != nil {
		return ActionExit, err
	}

	if result.Action == ActionBack || result.Action == ActionExit {
		return result.Action, nil
	}

	// Find selected bucket and store ARN
	for _, bucket := range c.state.TableBuckets {
		if bucket.Name == result.Selected {
			c.state.SelectedBucket = bucket.Name
			c.state.SelectedBucketARN = bucket.ARN
			break
		}
	}

	// Clear namespace cache when bucket changes
	c.state.Namespaces = nil
	c.state.Tables = nil

	return ActionSelect, nil
}

// navigateNamespaces handles Namespace level navigation
func (c *NavigationController) navigateNamespaces(ctx context.Context) (NavigationAction, error) {
	// Fetch namespaces if not cached
	if c.state.Namespaces == nil {
		namespaces, err := c.lister.ListNamespacesAll(ctx, c.state.SelectedBucketARN, "")
		if err != nil {
			return ActionExit, err
		}
		c.state.Namespaces = namespaces
	}

	if len(c.state.Namespaces) == 0 {
		fmt.Printf("No namespaces found in table bucket '%s'\n", c.state.SelectedBucket)
		return ActionBack, nil
	}

	// Extract names for selection
	names := make([]string, len(c.state.Namespaces))
	for i, ns := range c.state.Namespaces {
		names[i] = ns.Name
	}

	// Show back option to return to table bucket selection
	result, err := c.selector.SelectWithFilter("Select Namespace", names, true)
	if err != nil {
		return ActionExit, err
	}

	if result.Action == ActionBack {
		return ActionBack, nil
	}
	if result.Action == ActionExit {
		return ActionExit, nil
	}

	c.state.SelectedNamespace = result.Selected

	// Clear tables cache when namespace changes
	c.state.Tables = nil

	return ActionSelect, nil
}

// navigateTables handles Table level navigation
func (c *NavigationController) navigateTables(ctx context.Context) (NavigationAction, error) {
	// Fetch tables if not cached
	if c.state.Tables == nil {
		tables, err := c.lister.ListTablesAll(ctx, c.state.SelectedBucketARN, c.state.SelectedNamespace, "")
		if err != nil {
			return ActionExit, err
		}
		c.state.Tables = tables
	}

	if len(c.state.Tables) == 0 {
		fmt.Printf("No tables found in namespace '%s'\n", c.state.SelectedNamespace)
		return ActionBack, nil
	}

	// Extract names for selection
	names := make([]string, len(c.state.Tables))
	for i, tbl := range c.state.Tables {
		names[i] = tbl.Name
	}

	// Show back option to return to namespace selection
	result, err := c.selector.SelectWithFilter("Select Table", names, true)
	if err != nil {
		return ActionExit, err
	}

	if result.Action == ActionBack {
		return ActionBack, nil
	}
	if result.Action == ActionExit {
		return ActionExit, nil
	}

	// Display table details
	for _, tbl := range c.state.Tables {
		if tbl.Name == result.Selected {
			c.displayTableDetails(&tbl)
			break
		}
	}

	return ActionSelect, nil
}

// displayTableDetails prints the details of a table
func (c *NavigationController) displayTableDetails(tbl *TableInfo) {
	fmt.Printf("\nTable Details:\n")
	fmt.Printf("  Name:      %s\n", tbl.Name)
	fmt.Printf("  ARN:       %s\n", tbl.ARN)
	fmt.Printf("  Namespace: %s\n", tbl.Namespace)
	fmt.Printf("  Type:      %s\n", tbl.Type)
	fmt.Printf("  Created:   %s\n", tbl.CreatedAt.Format("2006-01-02 15:04:05"))
}
