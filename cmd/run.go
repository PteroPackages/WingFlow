package cmd

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/http"
	"github.com/pteropackages/wingflow/logger"
	ignore "github.com/sabhiram/go-gitignore"
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

func walkAll(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(root, entry.Name()) {
			if !entry.IsDir() {
				files = append(files, path)
			}

			return nil
		}

		if entry.IsDir() {
			other, err := walkAll(path)
			if err != nil {
				return err
			}

			files = append(files, other...)
		} else {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

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

	files, err := filepath.Glob(filepath.Join(temp, "*"))
	if err != nil {
		log.WithFatal(err)
	}

	// safety check
	if !contains(cfg.Repository.Exclude, ".git") {
		cfg.Repository.Exclude = append(cfg.Repository.Exclude, ".git")
	}
	sysIgnored := ignore.CompileIgnoreLines(cfg.Repository.Exclude...)

	var parsed []string
	for _, f := range files {
		paths, err := walkAll(f)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		for _, p := range paths {
			if sysIgnored.MatchesPath(p) {
				continue
			}

			parsed = append(parsed, p)
		}
	}

	if len(parsed) == 0 {
		log.Fatal("no files could be resolved to an absolute path")
	}

	http := http.New(cfg.Panel.URL, cfg.Panel.Key, cfg.Panel.ID)
	if ok, code, err := http.Test(); !ok {
		log.Fatal("%s (status: %d)", err.Error(), code)
	}

	state := "stop"
	if cfg.System.ForceKill {
		state = "kill"
	}

	if err = http.SetPower(state); err != nil {
		log.Warn("failed to update power state")
	}

	root, err := http.GetRootFiles()
	if err != nil {
		log.WithFatal(err)
	}

	if len(cfg.System.Ignore) == 0 {
		cfg.System.Ignore = append(cfg.System.Ignore, "*")
	}

	var remove []string
	if contains(cfg.System.Ignore, "*") {
		remove = root
	} else {
		panelIgnored := ignore.CompileIgnoreLines(cfg.System.Ignore...)

		for _, p := range root {
			if !panelIgnored.MatchesPath(p) {
				remove = append(remove, p)
			}
		}
	}

	if len(remove) != 0 {
		if err = http.DeleteFiles(remove); err != nil {
			log.WithFatal(err)
		}
	}

	wingsPath := func(p string) string {
		c := filepath.Join("/", strings.Split(p, temp)[1])
		if strings.Contains(c, "\\") {
			return strings.ReplaceAll(c, "\\", "/")
		}

		return c
	}

	for n, p := range parsed {
		log.Info("uploading file (%d/%d)", n+1, len(parsed))

		if err = http.WriteFile(p, wingsPath(p)); err != nil {
			log.Warn("failed to upload file:")
			log.Warn(p)
			log.Debug(err.Error())
		}
	}

	log.Info("uploads completed")
	if err = http.SetPower("start"); err != nil {
		log.Warn("failed to update power state")
	}
}
