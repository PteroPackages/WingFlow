package cmd

import (
	"fmt"
	"os"

	"github.com/pteropackages/wingflow/config"
	"github.com/spf13/cobra"
)

func handleCheckCmd(cmd *cobra.Command, args []string) {
	dir := cmd.Flag("dir").Value.String()
	cfg, err := config.Fetch(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var stack []string

	if cfg.Git.Address == "" {
		stack = append(stack, "git.address is required")
	}

	if cfg.Panel.URL == "" {
		stack = append(stack, "panel.url is required")
	}
	if cfg.Panel.Key == "" {
		stack = append(stack, "panel.key is required")
	}
	if len(cfg.Panel.Key) < 32 {
		stack = append(stack, "could not validate panel.key")
	}
	if cfg.Panel.ID == "" {
		stack = append(stack, "panel.id is required")
	}

	if len(cfg.Repository.Include) == 0 {
		stack = append(stack, "at least 1 file path or pattern is required for repository.include")
	}

	if len(stack) == 0 {
		fmt.Println("0 issues found")
	} else {
		fmt.Fprintf(os.Stderr, "%d error(s) found:\n", len(stack))
		for i, e := range stack {
			fmt.Fprintf(os.Stderr, "%d: %s\n", i, e)
		}
	}
}
