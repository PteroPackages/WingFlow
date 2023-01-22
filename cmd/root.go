package cmd

import (
	"os"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/http"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

var log *logger.Logger

var rootCmd = &cobra.Command{
	Use:     "wflow",
	Example: "wflow command [flags...]",
	Short:   "automatic project deployment for pterodactyl",
	Long:    "A tool for automatically deploying projects to Pterodactyl.",
	Version: Version,
}

var initCmd = &cobra.Command{
	Use:   "init [-f | --force]",
	Short: "creates a new config file",
	Long:  "Creates a new config file in the current workspace.",
	Run: func(cmd *cobra.Command, _ []string) {
		force, _ := cmd.Flags().GetBool("force")

		if err := config.Create(force); err != nil {
			if err.Error() == "exists" {
				log.Error("config file already exists in this directory")
				log.Error("re-run this command with '--force' to overwrite")
			} else {
				log.WithError(err)
			}
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check [--dry]",
	Short: "runs validation checks on the config file",
	Long:  "Runs validation checks on the config file.",
	Run: func(cmd *cobra.Command, _ []string) {
		cfg, err := config.Get(true)
		if err != nil {
			errs, ok := err.(validator.ValidationErrors)
			if !ok {
				log.WithError(err)
				return
			}

			log.Error("%d error(s) found", len(errs))
			log.Error("")
			for i, e := range errs {
				log.Error("%d: %s rule failed for the '%s' field", i+1, e.Tag(), e.StructNamespace())
			}
			return
		}

		log.Info("config checks passed")
		if dry, _ := cmd.Flags().GetBool("dry"); dry {
			return
		}

		client := http.New(cfg.Panel.URL, cfg.Panel.Key, cfg.Panel.ID)
		if st, err := client.TestConnection(); err != nil {
			log.Error("%s (status: %d)", err, st)
			return
		}

		log.Info("http checks passed")
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "fetches and deploys to pterodactyl",
	Long:  "Fetches and deploys the configured project to the pterodactyl server.",
	Run:   func(*cobra.Command, []string) {},
}

func init() {
	c := false
	if _, ok := os.LookupEnv("NO_COLOR"); !ok {
		if t := os.Getenv("TERM"); t != "DUMB" {
			c = true
		}
	}

	log = logger.New(c, false)

	initCmd.Flags().BoolP("force", "f", false, "force overwrite the existing config")
	checkCmd.Flags().Bool("dry", false, "don't perform http checks")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
	defer func() {
		if state := recover(); state != nil {
			stack := debug.Stack()

			log.Error("%v", state)
			log.Error(string(stack))
			os.Exit(1)
		}

		os.Exit(0)
	}()

	rootCmd.Execute()
}
