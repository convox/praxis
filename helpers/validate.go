package helpers

import (
	"fmt"
	"regexp"
)

// ValidateAppName asserts an alphanumeric app name is provided
func ValidateAppName(name string) error {
	if name == "" {
		return fmt.Errorf("app name required")
	}

	r := regexp.MustCompile("^([a-z][a-z0-9-]+)$")
	if !r.Match([]byte(name)) {
		return fmt.Errorf("app name invalid")
	}

	return nil
}
