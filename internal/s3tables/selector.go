package s3tables

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// SelectionResult represents the result of a selection
type SelectionResult struct {
	Selected string           // 選択されたアイテム
	Action   NavigationAction // 実行されたアクション
}

// InteractiveSelector provides interactive selection with filtering
type InteractiveSelector interface {
	// SelectWithFilter displays items with real-time filtering
	// showBack adds a ".. (Back)" option at the top when true
	// Returns the selected item and the action taken
	SelectWithFilter(label string, items []string, showBack bool) (*SelectionResult, error)
}

// Selector provides interactive selection UI (legacy interface)
type Selector interface {
	// Select displays items and returns the selected item
	Select(label string, items []string) (string, error)
}

// PromptSelector implements Selector using promptui
type PromptSelector struct{}

// NewPromptSelector creates a new PromptSelector
func NewPromptSelector() *PromptSelector {
	return &PromptSelector{}
}

// Select displays a selection prompt and returns the chosen item
func (s *PromptSelector) Select(label string, items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	prompt := promptui.Select{
		Label: label,
		Items: items,
		Size:  10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("selection failed: %w", err)
	}

	return items[idx], nil
}

// BackOption is the special option for navigating back
const BackOption = ".. (Back)"

// FilterablePromptSelector implements InteractiveSelector with filtering
type FilterablePromptSelector struct{}

// NewFilterablePromptSelector creates a new FilterablePromptSelector
func NewFilterablePromptSelector() *FilterablePromptSelector {
	return &FilterablePromptSelector{}
}

// SelectWithFilter displays a selection prompt with real-time filtering
// Uses promptui's Searcher feature for case-insensitive substring matching
// Selecting ".. (Back)" returns ActionBack
func (s *FilterablePromptSelector) SelectWithFilter(label string, items []string, showBack bool) (*SelectionResult, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select")
	}

	// Prepend back option if enabled
	displayItems := items
	if showBack {
		displayItems = append([]string{BackOption}, items...)
	}

	// Searcher function for real-time filtering
	// Automatically performs substring matching (equivalent to *input* pattern)
	searcher := func(input string, index int) bool {
		item := strings.ToLower(displayItems[index])
		input = strings.ToLower(input)
		// Automatic substring search (case-insensitive)
		return strings.Contains(item, input)
	}

	prompt := promptui.Select{
		Label:             label,
		Items:             displayItems,
		Size:              10,
		Searcher:          searcher,
		StartInSearchMode: false,
	}

	idx, selected, err := prompt.Run()
	if err != nil {
		// Ctrl+C triggers ErrInterrupt - treat as exit
		if err == promptui.ErrInterrupt {
			return &SelectionResult{Action: ActionExit}, nil
		}
		return nil, fmt.Errorf("selection failed: %w", err)
	}

	// Check if back option was selected
	if showBack && idx == 0 {
		return &SelectionResult{Action: ActionBack}, nil
	}

	return &SelectionResult{
		Selected: selected,
		Action:   ActionSelect,
	}, nil
}
