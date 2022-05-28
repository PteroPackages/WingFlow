package cmd

import (
	"os"
	"os/exec"

	"github.com/pteropackages/wingflow/config"
	_ "github.com/pteropackages/wingflow/http"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

func contains(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}

	return false
}

// func inPattern(slice []string, item string) bool {
// 	for _, i := range slice {
// 		if ok, _ := filepath.Match(i, item); ok {
// 			return true
// 		}
// 	}

// 	return false
// }

func handleRunCmd(cmd *cobra.Command, args []string) {
	nc := cmd.Flag("no-color").Value.String()
	dir := cmd.Flag("dir").Value.String()
	log := logger.New(nc, true)

	cfg, err := config.Fetch(dir)
	if err != nil {
		log.WithFatal(err)
	}

	if _, err = exec.Command("git", "--version").Output(); err != nil {
		log.Fatal("git must be installed for this command")
	}

	temp, err := os.MkdirTemp("", "wflow-*")
	if err != nil {
		log.Fatal("the system temp directory is unavailable")
	}

	if _, err = exec.Command("git", "clone", cfg.Git.Address, temp).Output(); err != nil {
		log.Fatal("failed to clone repository into temp directory")
	}

	// safety check
	if !contains(cfg.Repository.Exclude, ".git") {
		cfg.Repository.Exclude = append(cfg.Repository.Exclude, ".git")
	}

	// client := http.New(cfg.Panel.URL, cfg.Panel.Key, cfg.Panel.ID)
	// if ok, code, err := client.Test(); !ok {
	// 	fmt.Fprintf(os.Stderr, "%s (code: %d)\n", err.Error(), code)
	// 	os.Exit(1)
	// }
	// fmt.Println("test request succeeded; fetching upload url...")

	// url, err := client.GetUploadURL()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err.Error())
	// 	os.Exit(1)
	// }

	// fmt.Println(url)
}
