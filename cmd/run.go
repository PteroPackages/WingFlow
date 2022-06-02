package cmd

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pteropackages/wingflow/config"
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

func walkAll(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(root, entry.Name()) {
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

	// safety check
	if !contains(cfg.Repository.Exclude, ".git") {
		cfg.Repository.Exclude = append(cfg.Repository.Exclude, ".git")
	}

	files, err := filepath.Glob(filepath.Join(temp, "*"))
	if err != nil {
		log.WithFatal(err)
	}

	var includes []*regexp.Regexp
	replacer := strings.NewReplacer("*", ".*")
	for _, p := range cfg.Repository.Include {
		cleaned := replacer.Replace(filepath.Clean(p))
		includes = append(includes, regexp.MustCompile(cleaned))
	}

	var excludes []*regexp.Regexp
	for _, p := range cfg.Repository.Exclude {
		cleaned := replacer.Replace(filepath.Clean(p))
		excludes = append(excludes, regexp.MustCompile(cleaned))
	}

	match := func(f string) bool {
		for _, e := range excludes {
			if e.Match([]byte(f)) {
				return false
			}
		}

		for _, e := range includes {
			if e.Match([]byte(f)) {
				return true
			}
		}

		return false
	}

	var matched []string
	for _, f := range files {
		if match(f) {
			matched = append(matched, f)
		}
	}

	if len(matched) == 0 {
		log.Error("no files matched the configured patterns")
		log.Fatal("please ensure the patterns resolve to files in the repository")
	}

	var parsed []string
	for _, m := range matched {
		paths, err := walkAll(m)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		parsed = append(parsed, paths...)
	}

	if len(parsed) == 0 {
		log.Fatal("no files could be resolved to an absolute path")
	}
}
