package schema

import "fmt"

type Config struct {
	SchemaVersion int    `json:"schema_version"`
	IDPrefix      string `json:"id_prefix"`
	CreatedAt     string `json:"created_at"`
}

func ValidateConfig(cfg Config) error {
	if cfg.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema_version: %d", cfg.SchemaVersion)
	}
	if cfg.IDPrefix == "" {
		return fmt.Errorf("id_prefix is required")
	}
	if _, err := ParseTaskID(cfg.IDPrefix + "-1"); err != nil {
		return fmt.Errorf("invalid id_prefix %q", cfg.IDPrefix)
	}
	if err := ValidateTimestamp(cfg.CreatedAt); err != nil {
		return fmt.Errorf("created_at: %w", err)
	}
	return nil
}
