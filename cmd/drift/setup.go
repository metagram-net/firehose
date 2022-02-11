package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/metagram-net/firehose/clio"
	"github.com/metagram-net/firehose/drift"
)

func setupCmd(io *clio.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "setup",
		Aliases:      []string{"init"},
		Short:        "Set up the migrations directory",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			dir := viper.GetString("migrations-dir")
			path, err := drift.Setup(dir)
			if err != nil {
				return err
			}
			io.Logf("Created the first migration file: %s", path)
			io.Logf("Run the migrate command to apply it.")
			return nil
		},
	}
	return cmd
}
