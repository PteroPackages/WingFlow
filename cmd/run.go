package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/http"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
)

func handleRunCmd(cmd *cobra.Command, _ []string) {
	log.Info("getting configuration file")
	cfg, err := config.Get(true)
	if err != nil {
		log.WithError(err)
		return
	}

	log.Info("checking for git")
	if exec.Command("git", "--version").Run() != nil {
		log.Error("git version tool not found")
		log.Error("git version tool is required for this command")
		log.Error("see https://git-scm.com for more information")
		return
	} else {
		log.Info("git checks passed")
	}

	log.Info("testing panel connection")
	client := http.New(cfg.Panel.URL, cfg.Panel.Key, cfg.Panel.ID)
	if st, err := client.TestConnection(); err != nil {
		log.Error("%s (status: %d)", err, st)
		log.Error("make sure that your credentials are correct")
		return
	}

	log.Info("setting up repository space")
	temp, err := os.MkdirTemp("", "wflow-*")
	if err != nil {
		log.Error("failed to create directory:")
		log.WithError(err)
		log.Error("")
		log.Error("make sure that wingflow has permission to create temp directories")
		return
	}

	defer func(p string) {
		if err := os.RemoveAll(p); err != nil {
			log.Warn("failed to clean up temp directory:")
			log.Warn(err.Error())
			log.Warn("")
			log.Warn("path: %s", p)
		}
	}(temp)

	log.Info("cloning repository: %s", cfg.Git.Address)
	if exec.Command("git", "clone", "--depth=1", cfg.Git.Address, temp).Run() != nil {
		log.Error("failed to clone git repository")
		log.Error("make sure your git credentials are correct")
		return
	}

	log.Info("collecting files")
	files, err := filepath.Glob(filepath.Join(temp, "*", "**"))
	if err != nil {
		log.Error("failed to collect files:")
		log.WithError(err)
		log.Error("make sure that wingflow has permission to manage files")
		return
	}

	var parsed []string
	i := ignore.CompileIgnoreLines(cfg.Git.Files.Exclude...)
	for _, f := range files {
		if !i.MatchesPath(f) {
			parsed = append(parsed, f)
		}
	}

	files = []string{}
	i = ignore.CompileIgnoreLines(cfg.Git.Files.Include...)
	for _, p := range parsed {
		if i.MatchesPath(p) {
			files = append(files, p)
		}
	}

	if len(files) == 0 {
		log.Error("no files could be resolved to be included/excluded")
		log.Error("make sure you have specified valid file paths in the 'Config.Git.Files' fields")
		return
	}
	log.Info("files collected; preparing server for upload")

	st := "stop"
	if cfg.Panel.Signal.Kill {
		st = "kill"
	}

	if err = client.SetPower(st); err != nil {
		log.Error("failed to set power state (state: %s):", st)
		log.WithError(err)
		// TODO: add Config.Panel.Signal.Force
		log.Error("cannot continue process")
		return
	}

	if t := cfg.Panel.Signal.Timeout; t != 0 {
		log.Info("waiting for signal timeout to end (%sms)", t)
		// TODO: spinner here?
		time.Sleep(time.Duration(t))
	}

	// TODO: create a BACKUP not an ARCHIVE!

	if cfg.Panel.Files.Truncate {
		log.Info("truncating server files")
		root, err := client.GetFiles()
		if err != nil {
			log.Error("failed to get server files:")
			log.WithError(err)
			// TODO: add force option here too
			log.Error("cannot continue process")
			return
		}

		if err = client.DeleteFiles(root); err != nil {
			log.Error("failed to truncate server files:")
			log.WithError(err)
			// TODO: here too...
			log.Error("cannot continue process")
			return
		}
	}

	log.Info("compressing files")
	buf, failed := compress(files)
	sep := len(files) - len(failed)
	if len(failed) != 0 {
		log.Warn("%d file(s) could not be written to the archive", len(failed))
	}

	if sep == 0 {
		log.Error("no files could be written to the archive")
		log.Error("this could be due to an unsupported file types or corrupted data")
		log.Error("if this error repeats, contact PteroPackages support")
		return
	}

	log.Info("wrote %d files to the archive", sep)
	log.Info("uploading archive to the server")

	// TODO: retry for this?
	if err = client.UploadFile(*buf); err != nil {
		log.Error("failed to upload archive to the server:")
		log.WithError(err)
		return
	}

	log.Info("upload completed!")
	log.Info("cleaning up leftover processes...")
}

func compress(f []string) (*bytes.Buffer, []string) {
	buf := bytes.Buffer{}
	var failed []string

	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	defer tw.Close()
	defer gw.Close()

	for _, p := range f {
		st, err := os.Stat(p)
		if err != nil {
			failed = append(failed, p)
			continue
		}

		b, err := os.ReadFile(p)
		if err != nil {
			failed = append(failed, p)
			continue
		}

		tw.WriteHeader(&tar.Header{
			Name: st.Name(),
			Size: st.Size(),
			Mode: int64(st.Mode()),
		})
		tw.Write(b)
	}

	return &buf, failed
}
