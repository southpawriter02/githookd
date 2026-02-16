package config

import (
	"fmt"
	"strings"
)

// StandardHooks is the list of all recognized Git hook names.
var StandardHooks = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
	"push-to-checkout",
	"pre-auto-gc",
	"post-rewrite",
	"sendemail-validate",
}

// ValidateHookName checks if the given name is a recognized Git hook name.
func ValidateHookName(name string) error {
	for _, h := range StandardHooks {
		if h == name {
			return nil
		}
	}
	return fmt.Errorf("unknown hook '%s'; must be one of: %s",
		name, strings.Join(StandardHooks, ", "))
}

// SuggestHookName returns the closest matching hook name if the Levenshtein
// distance is within a reasonable threshold, or an empty string if no close match.
func SuggestHookName(name string) string {
	bestMatch := ""
	bestDist := len(name) // worst case
	threshold := 3

	for _, h := range StandardHooks {
		d := levenshtein(name, h)
		if d < bestDist {
			bestDist = d
			bestMatch = h
		}
	}

	if bestDist <= threshold {
		return bestMatch
	}
	return ""
}

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use two rows for space efficiency
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

