package cmd

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/http"
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

func addToArchive(writer *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, _ := file.Stat()
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	if err = writer.WriteHeader(header); err != nil {
		return err
	}

	if _, err = io.Copy(writer, file); err != nil {
		return err
	}

	return nil
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

	tarfile, err := os.CreateTemp("", "export-*.tar.gz")
	// testing
	// tarfile, err := os.Create("export.tar.gz")
	if err != nil {
		log.Fatal("the system temp directory is unavailable")
	}
	tstat, _ := tarfile.Stat()
	defer tarfile.Close()

	gz := gzip.NewWriter(tarfile)
	tz := tar.NewWriter(gz)
	defer gz.Close()
	defer tz.Close()

	for _, p := range parsed {
		log.Debug(p)
		if err = addToArchive(tz, p); err != nil {
			log.Warn("failed to archive file:")
			log.Warn(p)
		}
	}

	log.Info("completed archive; uploading to server")

	client := http.New(cfg.Panel.URL, cfg.Panel.Key, cfg.Panel.ID)
	if ok, code, err := client.Test(); !ok {
		log.Fatal("%s (code: %d)", code, err.Error())
	}

	if err = client.UploadFile(tstat.Name(), tarfile); err != nil {
		log.Error("failed to upload tarball:")
		log.WithFatal(err)
	}

	log.Info("tarball upload complete")
}
