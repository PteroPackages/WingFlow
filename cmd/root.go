package cmd

import (
	"os"
	"runtime/debug"

	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "wflow",
	Example: "wflow [flags...] <command>",
	Short:   "automatic project deployment for pterodactyl",
	Long:    "A tool for automatically deploying projects to Pterodactyl.",
	Version: Version,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "creates a new config file",
	Long:  "Creates a new config file in the current workspace.",
	Run: func(*cobra.Command, []string) {
		log := logger.New(false, false)
		if err := config.Create(false); err != nil {
			log.WithFatal(err)
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "runs validation checks on the config file",
	Long:  "Runs validation checks on the config file.",
	Run:   func(*cobra.Command, []string) {},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "fetches and deploys to pterodactyl",
	Long:  "Fetches and deploys the configured project to the pterodactyl server.",
	Run:   func(*cobra.Command, []string) {},
}

func init() {
	// dir, _ := os.Getwd()
	// noColor := false

	// if v, ok := os.LookupEnv("TERM"); ok {
	// 	noColor = v == "dumb"
	// }
	// if _, ok := os.LookupEnv("NO_COLOR"); ok {
	// 	noColor = false
	// }

	// initCmd.Flags().String("dir", dir, "the directory to create the config in")
	// initCmd.Flags().BoolP("force", "f", false, "force overwrite the config file")
	// initCmd.Flags().Bool("no-color", noColor, "disable color for the output")
	// checkCmd.Flags().String("dir", dir, "the directory of the config file")
	// checkCmd.Flags().Bool("no-color", noColor, "disable color for the output")
	// runCmd.Flags().Bool("debug", false, "output debug logs")
	// runCmd.Flags().String("dir", dir, "the directory of the config file")
	// runCmd.Flags().Bool("no-color", noColor, "disable color for the output")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
	defer func() {
		if state := recover(); state != nil {
			stack := debug.Stack()

			log := logger.New(false, false)
			log.Error("%v", state)
			log.Fatal(string(stack))
		}

		os.Exit(0)
	}()

	rootCmd.Execute()
}
