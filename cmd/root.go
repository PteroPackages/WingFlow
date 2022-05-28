package cmd

import (
	"os"

	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:       "wflow",
	Short:     "automatic project deployment for pterodactyl",
	Long:      "a cli tool for automatically deploying projects to pterodactyl",
	ValidArgs: []string{"init", "check", "run"},
	Version:   Version,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "creates a new config file",
	Long:  "creates a new config file in the current workspace (or a specified one)",
	Run: func(cmd *cobra.Command, args []string) {
		nc := cmd.Flag("no-color").Value.String()
		dir := cmd.Flag("dir").Value.String()
		log := logger.New(nc, true)

		err := config.Create(dir)
		if err != nil {
			log.WithFatal(err)
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "runs validation checks on the config file",
	Long:  "runs validation checks on the config file",
	Run:   handleCheckCmd,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "fetches and deploys to pterodactyl",
	Long:  "fetches and deploys the configured project to the pterodactyl server",
	Run:   handleRunCmd,
}

func init() {
	dir, _ := os.Getwd()
	noColor := false

	if v, ok := os.LookupEnv("TERM"); ok {
		noColor = v == "dumb"
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		noColor = false
	}

	initCmd.Flags().String("dir", dir, "the directory to create the config in")
	initCmd.Flags().Bool("no-color", noColor, "disable color for the output")
	checkCmd.Flags().String("dir", dir, "the directory of the config file")
	checkCmd.Flags().Bool("no-color", noColor, "disable color for the output")
	runCmd.Flags().String("dir", dir, "the directory of the config file")
	runCmd.Flags().Bool("no-color", noColor, "disable color for the output")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
	rootCmd.Execute()
}
