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
		log := logger.New(true)
		dir := cmd.Flag("dir").Value.String()
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
	Use: "run",
	Run: handleRunCmd,
}

func init() {
	dir, _ := os.Getwd()

	initCmd.Flags().String("dir", dir, "the directory to create the config in")
	checkCmd.Flags().String("dir", dir, "the directory of the config file")
	runCmd.Flags().String("dir", dir, "the directory of the config file")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
	rootCmd.Execute()
}
