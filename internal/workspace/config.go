package workspace

import (
	"encoding/json"
	"os"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
)

func LoadConfig(ctx Context) (schema.Config, error) {
	op := "workspace.LoadConfig"
	data, err := os.ReadFile(ctx.ConfigPath)
	if err != nil {
		return schema.Config{}, errs.E(errs.KindInvalidConfig, op, ctx.ConfigPath, err)
	}

	var cfg schema.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return schema.Config{}, errs.E(errs.KindInvalidConfig, op, ctx.ConfigPath, err)
	}
	if err := schema.ValidateConfig(cfg); err != nil {
		return schema.Config{}, errs.E(errs.KindInvalidConfig, op, ctx.ConfigPath, err)
	}

	return cfg, nil
}
