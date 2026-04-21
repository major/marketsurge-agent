package queries

import (
	"fmt"
	"strings"
)

// Load returns the content of the named GraphQL query file.
// The name should include the .graphql extension.
func Load(name string) (string, error) {
	content, err := FS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("query file %q not found: %w", name, err)
	}
	return strings.TrimSpace(string(content)), nil
}
