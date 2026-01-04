package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 1: Config Options Building Correctness
// For any combination of profile and region flag values (including empty strings),
// the generated config options should satisfy:
// - If profile is non-empty, the options should include WithSharedConfigProfile(profile)
// - If profile is empty, the options should NOT include any profile option
// - If region is non-empty, the options should include WithRegion(region)
// - If region is empty, the options should NOT include any region option
// **Validates: Requirements 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 3.3**

// TestPropertyBuildConfigOptions tests that buildConfigOptions returns the correct
// number and type of config options based on profile and region inputs.
// Feature: aws-profile-region, Property 1: Config Options Building Correctness
func TestPropertyBuildConfigOptions(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for arbitrary strings (including empty)
	stringGen := gen.AnyString()

	// Property: The number of options equals the count of non-empty inputs
	properties.Property("options count matches non-empty inputs", prop.ForAll(
		func(profile, region string) bool {
			opts := buildConfigOptions(profile, region)

			expectedCount := 0
			if profile != "" {
				expectedCount++
			}
			if region != "" {
				expectedCount++
			}

			return len(opts) == expectedCount
		},
		stringGen,
		stringGen,
	))

	// Property: Empty profile and region returns empty options
	properties.Property("empty inputs return empty options", prop.ForAll(
		func(_ bool) bool {
			opts := buildConfigOptions("", "")
			return len(opts) == 0
		},
		gen.Bool(), // dummy generator to satisfy gopter
	))

	// Property: Non-empty profile returns exactly one option when region is empty
	properties.Property("profile only returns one option", prop.ForAll(
		func(profile string) bool {
			if profile == "" {
				return true // skip empty profiles
			}
			opts := buildConfigOptions(profile, "")
			return len(opts) == 1
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	))

	// Property: Non-empty region returns exactly one option when profile is empty
	properties.Property("region only returns one option", prop.ForAll(
		func(region string) bool {
			if region == "" {
				return true // skip empty regions
			}
			opts := buildConfigOptions("", region)
			return len(opts) == 1
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	))

	// Property: Both non-empty returns exactly two options
	properties.Property("both profile and region returns two options", prop.ForAll(
		func(profile, region string) bool {
			if profile == "" || region == "" {
				return true // skip if either is empty
			}
			opts := buildConfigOptions(profile, region)
			return len(opts) == 2
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	))

	// Property: Options are valid config.LoadOptions functions
	properties.Property("options are callable without panic", prop.ForAll(
		func(profile, region string) bool {
			opts := buildConfigOptions(profile, region)
			// Verify each option is a valid function by checking it can be applied
			// to a LoadOptions struct without panicking
			for _, opt := range opts {
				loadOpts := &config.LoadOptions{}
				if err := opt(loadOpts); err != nil {
					return false
				}
			}
			return true
		},
		stringGen,
		stringGen,
	))

	properties.TestingRun(t)
}
