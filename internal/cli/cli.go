package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/apeople/internal/config"
	"github.com/mph-llm-experiments/apeople/internal/ui"
)

// Run executes the CLI with the given config and arguments.
func Run(cfg *config.Config, args []string) error {
	remaining, err := ParseGlobalFlags(args)
	if err != nil {
		return err
	}

	// Reload config if --config flag was provided
	if globalFlags.Config != "" {
		newCfg, err := config.Load(globalFlags.Config)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		cfg = newCfg
	}

	// Override contacts directory if --dir flag was provided
	if globalFlags.Dir != "" {
		cfg.ContactsDirectory = globalFlags.Dir
	}

	// Also check APEOPLE_DIR env var
	if envDir := os.Getenv("APEOPLE_DIR"); envDir != "" && globalFlags.Dir == "" {
		cfg.ContactsDirectory = envDir
	}

	// If no arguments, launch TUI
	if len(remaining) == 0 {
		m := ui.NewModel(cfg.ContactsDirectory)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	}

	// Create root command
	root := &Command{
		Name:  "apeople",
		Usage: "apeople <command> [options]",
		Description: `Agent-first contacts management using Denote file conventions.

Commands:
  list       List contacts
  show       Show contact details
  new        Create a new contact
  update     Update contact fields
  log        Log an interaction
  bump       Bump a contact (review without contacting)
  delete     Delete a contact

Global Options:
  --config PATH  Use specific config file
  --dir PATH     Override contacts directory
  --json         Output in JSON format
  --no-color     Disable color output
  --quiet, -q    Minimal output`,
	}

	root.Subcommands = append(root.Subcommands,
		listCommand(cfg),
		showCommand(cfg),
		newCommand(cfg),
		updateCommand(cfg),
		logCommand(cfg),
		bumpCommand(cfg),
		deleteCommand(cfg),
	)

	return root.Execute(remaining)
}
