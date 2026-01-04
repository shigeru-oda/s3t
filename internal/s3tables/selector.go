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

// promptRunner abstracts promptui.Select.Run for testing
type promptRunner interface {
	Run() (int, string, error)
}

// PromptSelector implements Selector using promptui
type PromptSelector struct {
	// runFunc allows overriding the prompt runner for testing
	runFunc func(prompt promptRunner) (int, string, error)
}

// NewPromptSelector creates a new PromptSelector
func NewPromptSelector() *PromptSelector {
	return &PromptSelector{
		runFunc: defaultPromptRun,
	}
}

// defaultPromptRun is the default implementation that calls prompt.Run()
func defaultPromptRun(prompt promptRunner) (int, string, error) {
	return prompt.Run()
}

// Select displays a selection prompt and returns the chosen item
func (s *PromptSelector) Select(label string, items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	prompt := &promptui.Select{
		Label: label,
		Items: items,
		Size:  10,
	}

	idx, _, err := s.runFunc(prompt)
	if err != nil {
		return "", fmt.Errorf("selection failed: %w", err)
	}

	return items[idx], nil
}

// BackOption is the special option for navigating back
const BackOption = ".. (Back)"

// FilterablePromptSelector implements InteractiveSelector with filtering
type FilterablePromptSelector struct {
	// runFunc allows overriding the prompt runner for testing
	runFunc func(prompt promptRunner) (int, string, error)
}

// NewFilterablePromptSelector creates a new FilterablePromptSelector
func NewFilterablePromptSelector() *FilterablePromptSelector {
	return &FilterablePromptSelector{
		runFunc: defaultPromptRun,
	}
}

// createSearcher creates a searcher function for substring matching
func createSearcher(items []string) func(string, int) bool {
	return func(input string, index int) bool {
		item := strings.ToLower(items[index])
		input = strings.ToLower(input)
		return strings.Contains(item, input)
	}
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

	prompt := &promptui.Select{
		Label:             label,
		Items:             displayItems,
		Size:              10,
		Searcher:          createSearcher(displayItems),
		StartInSearchMode: false,
	}

	idx, selected, err := s.runFunc(prompt)
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
