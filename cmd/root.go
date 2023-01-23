package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/http"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

var (
	log      *logger.Logger
	useColor = false
)

var (
	formatAscii = strings.NewReplacer("$H", "\033[1m««« \033[36mSOAR\033[0m\033[1m »»»\033[0m", "$B", "\033[36m", "$S", "\033[1m", "$R", "\033[0m")
	formatNone  = strings.NewReplacer("$H", "««« SOAR »»»", "$B", "\033[0m", "$S", "\033[0m", "$R", "\033[0m")
)

var rootCmd = &cobra.Command{
	Use:     "wflow command",
	Short:   "A tool for automatically deploying projects to Pterodactyl",
	Long:    "A tool for automatically deploying projects to Pterodactyl.",
	Version: Version,
}

var initCmd = &cobra.Command{
	Use:   "init [-f | --force]",
	Short: "Creates a new config file in the current workspace",
	Long:  initCmdHelp,
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
	Long:  checkCmdHelp,
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
	Short: "Fetches and deploys the configured project to the Pterodactyl server",
	Long:  runCmdHelp,
	Run:   func(*cobra.Command, []string) {},
}

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); !ok {
		if t := os.Getenv("TERM"); t != "DUMB" {
			useColor = true
		}
	}

	log = logger.New(useColor, false)

	rootCmd.Flags().Bool("no-color", false, "disable ansi color codes")
	initCmd.Flags().BoolP("force", "f", false, "force overwrite the existing config")
	checkCmd.Flags().Bool("dry", false, "don't perform http checks")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(runCmd)

	cobra.AddTemplateFunc("splitLines", func(s string) []string {
		a := strings.Split(s, "\n")
		var r []string
		for _, i := range a {
			if len(i) != 0 {
				r = append(r, i)
			}
		}
		return r
	})
	rootCmd.SetHelpTemplate(color(`$H
{{.Short}}

$BUsage$R
» $S{{.UseLine}}$R
{{if gt (len .Commands) 0}}
$BCommands$R{{range .Commands}}
» $S{{rpad .Name .NamePadding}}$R {{.Short}}{{end}}
{{end}}
$BFlags$R{{range .LocalFlags.FlagUsages | splitLines}}
» {{.}}{{end}}

$BDescription$R
{{.Long}}
{{if gt (len .Commands) 0}}
Use '$Bwflow{{if (eq .Name "wflow" | not)}} {{.Name}}{{end}} --help$R' for more information about a command{{end}}
`))
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

func color(s string, data ...interface{}) string {
	s = fmt.Sprintf(s, data...)
	if useColor {
		return formatAscii.Replace(s)
	} else {
		return formatNone.Replace(s)
	}
}
