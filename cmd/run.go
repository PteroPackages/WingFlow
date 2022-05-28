package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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

func recursiveWalk(path string) ([]string, error) {
	if filepath.IsAbs(path) {
		return []string{path}, nil
	}

	var parsed []string
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.Mode()&os.ModeSymlink != 0o0 {
		return nil, fmt.Errorf("symlinks are not followed")
	}

	if info.Mode()&fs.FileMode(os.O_RDONLY) == 0o0 {
		return nil, fmt.Errorf("cannot read directory")
	}

	entry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, e := range entry {
		if e.IsDir() {
			paths, err := recursiveWalk(filepath.Join(path, e.Name()))
			if err != nil {
				// cant append nil value, ignore
				continue
			}

			parsed = append(parsed, paths...)
		} else {
			parsed = append(parsed, filepath.Join(path, e.Name()))
		}
	}

	return parsed, nil
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
		cleaned := filepath.Clean(p)
		cleaned = replacer.Replace(cleaned)
		includes = append(includes, regexp.MustCompile(cleaned))
	}

	match := func(f string) bool {
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
		paths, err := recursiveWalk(m)
		if err != nil {
			continue
		}

		parsed = append(parsed, paths...)
	}

	if len(parsed) == 0 {
		log.Fatal("no files could be resolved to an absolute path")
	}

	fmt.Println(parsed)

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
