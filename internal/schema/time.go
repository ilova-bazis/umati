package schema

import (
	"fmt"
	"strings"
	"time"
)

func NowTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func ValidateTimestamp(value string) error {
	if value == "" {
		return fmt.Errorf("timestamp is required")
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return fmt.Errorf("invalid timestamp %q: %w", value, err)
	}
	if !strings.HasSuffix(value, "Z") {
		return fmt.Errorf("timestamp must be UTC: %q", value)
	}
	if t.UTC().Format(time.RFC3339) != value {
		return fmt.Errorf("timestamp must be normalized RFC3339 UTC: %q", value)
	}
	return nil
}
