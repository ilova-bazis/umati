package cli

import (
	"fmt"
	"os"

	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/tui"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func runBoard(args []string) error {
	var agent string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			i++
			if i >= len(args) {
				return fmt.Errorf("--agent requires a value")
			}
			agent = args[i]
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	if agent == "" {
		return fmt.Errorf("--agent is required (e.g. umati board --agent human)")
	}

	actor := schema.Actor(agent)
	if !schema.IsValidActor(actor) {
		return fmt.Errorf("invalid agent: %s (valid: human, claude, opencode, codex)", agent)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := workspace.Discover(wd)
	if err != nil {
		return renderError(err)
	}

	cfg, err := workspace.LoadConfig(ctx)
	if err != nil {
		return renderError(err)
	}

	return tui.Run(ctx, cfg, actor)
}
