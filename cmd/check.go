package cmd

import (
	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

func handleCheckCmd(cmd *cobra.Command, args []string) {
	nc, _ := cmd.Flags().GetBool("no-color")
	dir := cmd.Flag("dir").Value.String()
	log := logger.New(nc, true)

	cfg, err := config.Fetch(dir)
	if err != nil {
		log.WithFatal(err)
	}

	var stack []string

	if cfg.Git.Address == "" {
		stack = append(stack, "'git.address' is required")
	}

	if cfg.Panel.URL == "" {
		stack = append(stack, "'panel.url' is required")
	}
	if cfg.Panel.Key == "" {
		stack = append(stack, "'panel.key' is required")
	}
	if len(cfg.Panel.Key) < 32 {
		stack = append(stack, "could not validate 'panel.key'")
	}
	if cfg.Panel.ID == "" {
		stack = append(stack, "'panel.id' is required")
	}

	if len(cfg.Repository.Include) == 0 {
		stack = append(stack, "at least 1 file path or pattern is required for 'repository.include'")
	}

	if len(stack) == 0 {
		log.Info("0 issues found")
	} else {
		log.Error("%d error(s) found:", len(stack))
		for _, e := range stack {
			log.Error("  - %s", e)
		}
	}
}
