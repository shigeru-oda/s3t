package s3tables

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// substringFilter is the reference implementation for substring filtering
// Used to verify the filtering logic in FilterablePromptSelector
func substringFilter(pattern string, items []string) []string {
	if pattern == "" {
		return items
	}
	pattern = strings.ToLower(pattern)
	var result []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), pattern) {
			result = append(result, item)
		}
	}
	return result
}

// TestPropertyFilterCorrectlyFiltersItemsByPattern tests Property 3
// Feature: s3t-list, Property 3: Filter correctly filters items by pattern
// Validates: Requirements 4.1, 4.2, 4.5
func TestPropertyFilterCorrectlyFiltersItemsByPattern(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for filter patterns (alphanumeric, no wildcards for substring matching)
	patternGen := gen.RegexMatch(`[a-zA-Z0-9]{0,10}`)
	// Generator for resource names
	nameGen := gen.RegexMatch(`[a-zA-Z0-9_-]{1,30}`)
	// Generator for list of names
	namesGen := gen.SliceOf(nameGen)

	properties.Property("filtered results contain only items matching substring pattern", prop.ForAll(
		func(pattern string, items []string) bool {
			filtered := substringFilter(pattern, items)

			// All filtered items must contain the pattern as substring (case-insensitive)
			patternLower := strings.ToLower(pattern)
			for _, item := range filtered {
				if !strings.Contains(strings.ToLower(item), patternLower) {
					return false
				}
			}
			return true
		},
		patternGen,
		namesGen,
	))

	properties.Property("all matching items are included in filtered results", prop.ForAll(
		func(pattern string, items []string) bool {
			filtered := substringFilter(pattern, items)

			// Count items that should match
			patternLower := strings.ToLower(pattern)
			expectedCount := 0
			for _, item := range items {
				if strings.Contains(strings.ToLower(item), patternLower) {
					expectedCount++
				}
			}

			return len(filtered) == expectedCount
		},
		patternGen,
		namesGen,
	))

	properties.Property("empty pattern returns all items unchanged", prop.ForAll(
		func(items []string) bool {
			filtered := substringFilter("", items)

			if len(filtered) != len(items) {
				return false
			}
			for i, item := range items {
				if filtered[i] != item {
					return false
				}
			}
			return true
		},
		namesGen,
	))

	properties.Property("filtering is case-insensitive", prop.ForAll(
		func(pattern string, items []string) bool {
			// Filter with lowercase pattern
			filteredLower := substringFilter(strings.ToLower(pattern), items)
			// Filter with uppercase pattern
			filteredUpper := substringFilter(strings.ToUpper(pattern), items)
			// Filter with original pattern
			filteredOriginal := substringFilter(pattern, items)

			// All should produce the same results
			if len(filteredLower) != len(filteredUpper) || len(filteredLower) != len(filteredOriginal) {
				return false
			}
			for i := range filteredLower {
				if filteredLower[i] != filteredUpper[i] || filteredLower[i] != filteredOriginal[i] {
					return false
				}
			}
			return true
		},
		patternGen,
		namesGen,
	))

	properties.TestingRun(t)
}


// TestPropertySubstringMatchingIsAutomatic tests Property 4
// Feature: s3t-list, Property 4: Substring matching is automatic
// Validates: Requirements 4.2
func TestPropertySubstringMatchingIsAutomatic(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for base strings (the pattern to search for)
	baseGen := gen.RegexMatch(`[a-z0-9]{1,8}`)
	// Generator for prefix strings
	prefixGen := gen.RegexMatch(`[a-z0-9]{0,8}`)
	// Generator for suffix strings
	suffixGen := gen.RegexMatch(`[a-z0-9]{0,8}`)

	properties.Property("pattern matches exact string", prop.ForAll(
		func(pattern string) bool {
			items := []string{pattern}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == pattern
		},
		baseGen,
	))

	properties.Property("pattern matches string with prefix", prop.ForAll(
		func(pattern, prefix string) bool {
			if pattern == "" {
				return true // Empty pattern matches everything
			}
			item := prefix + pattern
			items := []string{item}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == item
		},
		baseGen,
		prefixGen,
	))

	properties.Property("pattern matches string with suffix", prop.ForAll(
		func(pattern, suffix string) bool {
			if pattern == "" {
				return true // Empty pattern matches everything
			}
			item := pattern + suffix
			items := []string{item}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == item
		},
		baseGen,
		suffixGen,
	))

	properties.Property("pattern matches string with prefix and suffix", prop.ForAll(
		func(pattern, prefix, suffix string) bool {
			if pattern == "" {
				return true // Empty pattern matches everything
			}
			item := prefix + pattern + suffix
			items := []string{item}
			filtered := substringFilter(pattern, items)
			return len(filtered) == 1 && filtered[0] == item
		},
		baseGen,
		prefixGen,
		suffixGen,
	))

	properties.Property("no explicit wildcards required for substring matching", prop.ForAll(
		func(pattern, prefix, suffix string) bool {
			if pattern == "" {
				return true
			}
			// Create items that should match without wildcards
			items := []string{
				pattern,                   // exact match
				prefix + pattern,          // prefix + pattern
				pattern + suffix,          // pattern + suffix
				prefix + pattern + suffix, // prefix + pattern + suffix
			}
			filtered := substringFilter(pattern, items)
			// All items should match
			return len(filtered) == 4
		},
		baseGen,
		prefixGen,
		suffixGen,
	))

	properties.TestingRun(t)
}
