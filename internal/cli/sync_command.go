package cli

import (
	"flag"
	"fmt"
	"log"

	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/apeople/internal/config"
)

func syncCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	push := fs.Bool("push", false, "Push local changes to R2 (default)")
	pull := fs.Bool("pull", false, "Pull remote changes from R2")

	return &Command{
		Name:        "sync",
		Usage:       "apeople sync [--push|--pull]",
		Description: "Sync contact files with Cloudflare R2",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			direction := "push"
			if *pull {
				direction = "pull"
			}
			_ = push // push is the default

			acoreCfg, err := acore.LoadConfig()
			if err != nil {
				return fmt.Errorf("loading acore config: %w", err)
			}
			if !acoreCfg.R2.Enabled() {
				return fmt.Errorf("R2 not configured — add [r2] section to ~/.config/acore/config.toml")
			}

			local := acore.NewLocalStore(cfg.ContactsDirectory)
			remote, err := acoreCfg.R2StoreFor("apeople")
			if err != nil {
				return fmt.Errorf("creating R2 store: %w", err)
			}

			result, err := acore.SyncApp(local, remote, direction, acore.SyncOpts{Delete: true})
			if err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}

			if !globalFlags.Quiet {
				printSyncResult(result, direction)
			}
			return nil
		},
	}
}

func printSyncResult(result *acore.SyncResult, direction string) {
	if len(result.Pushed) == 0 && len(result.Deleted) == 0 && len(result.Errors) == 0 {
		fmt.Println("Already in sync.")
		return
	}

	verb := "pushed"
	if direction == "pull" {
		verb = "pulled"
	}

	if len(result.Pushed) > 0 {
		fmt.Printf("%d files %s\n", len(result.Pushed), verb)
	}
	if len(result.Deleted) > 0 {
		fmt.Printf("%d files deleted from target\n", len(result.Deleted))
	}
	for _, err := range result.Errors {
		fmt.Printf("  error: %v\n", err)
	}
}

// SyncOnStartup pulls from R2 if configured. Errors are logged, not fatal.
func SyncOnStartup(cfg *config.Config) {
	acoreCfg, err := acore.LoadConfig()
	if err != nil {
		return
	}
	if !acoreCfg.R2.Enabled() {
		return
	}

	local := acore.NewLocalStore(cfg.ContactsDirectory)
	remote, err := acoreCfg.R2StoreFor("apeople")
	if err != nil {
		return
	}

	if _, err := acore.SyncApp(local, remote, "pull", acore.SyncOpts{Delete: false}); err != nil {
		log.Printf("sync pull: %v", err)
	}
}

// SyncOnShutdown pushes to R2 if configured. Errors are logged, not fatal.
func SyncOnShutdown(cfg *config.Config) {
	acoreCfg, err := acore.LoadConfig()
	if err != nil {
		return
	}
	if !acoreCfg.R2.Enabled() {
		return
	}

	local := acore.NewLocalStore(cfg.ContactsDirectory)
	remote, err := acoreCfg.R2StoreFor("apeople")
	if err != nil {
		return
	}

	if _, err := acore.SyncApp(local, remote, "push", acore.SyncOpts{Delete: false}); err != nil {
		log.Printf("sync push: %v", err)
	}
}
