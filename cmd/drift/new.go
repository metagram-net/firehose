package main

import (
	"fmt"

	"github.com/metagram-net/firehose/drift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCmd() *cobra.Command {
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
			path, err := drift.NewFile(dir, id, slug, template)
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	}
	flags := cmd.Flags()
	flags.Var(&id, "id", "Migration ID override (default: Unix timestamp in seconds)")
	flags.StringVar(&slug, "slug", "", "Short text used to name the migration")
	cmd.MarkFlagRequired("slug")
	return cmd
}

// TODO: Load this template from a configurable file.
// TODO: Maybe allow real Go templating in it?
var template = "-- TODO: write your migration here\n"
