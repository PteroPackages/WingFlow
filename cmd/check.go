package cmd

import (
	"github.com/go-playground/validator/v10"
	"github.com/pteropackages/wingflow/config"
	"github.com/pteropackages/wingflow/logger"
	"github.com/spf13/cobra"
)

func handleCheckCmd(cmd *cobra.Command, _ []string) {
	nc, _ := cmd.Flags().GetBool("no-color")
	dir := cmd.Flag("dir").Value.String()
	log := logger.New(nc, false)

	cfg, err := config.Fetch(dir)
	if err != nil {
		log.WithFatal(err)
	}

	validate := validator.New()
	err = validate.Struct(cfg)
	if err == nil {
		return
	}

	errs := err.(validator.ValidationErrors)
	for _, e := range errs {
		log.Error(e.Error())
	}

	log.Fatal("%d total errors", len(errs))
}
