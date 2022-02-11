package main

import (
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/metagram-net/firehose/clio"
	"github.com/metagram-net/firehose/drift"
)

func newCmd(io *clio.IO) *cobra.Command {
	var (
		// Set the default ID out of range to distinguish explicit zero.
		id   drift.MigrationID = -1
		slug string
	)

	cmd := &cobra.Command{
		Use:          "new",
		Short:        "Create a new migration file",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := viper.GetString("migrations-dir")
			templateFile := viper.GetString("template-file")

			tmpl, err := migrationTemplate(templateFile)
			if err != nil {
				return err
			}

			path, err := drift.NewFile(io, dir, id, slug, tmpl)
			if err != nil {
				return err
			}
			io.Logf("Created new migration file: %s", path)
			io.Printf(path)
			return nil
		},
	}
	flags := cmd.Flags()
	flags.Var(&id, "id", "Migration ID override (default: Unix timestamp in seconds)")
	flags.StringVar(&slug, "slug", "", "Short text used to name the migration")
	cmd.MarkFlagRequired("slug")
	flags.String("template", "", "Template file for the migration")
	viper.BindPFlag("template-file", flags.Lookup("template"))
	return cmd
}

func migrationTemplate(path string) (*template.Template, error) {
	if path == "" {
		// Drift uses a sensible default template in case of nil.
		return nil, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return template.New("migration").Parse(string(b))
}
