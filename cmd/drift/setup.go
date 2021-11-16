package main

import (
	"fmt"

	"github.com/metagram-net/firehose/drift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupCmd() *cobra.Command {
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
			fmt.Printf("Wrote %s\n", path)
			return nil
		},
	}
	return cmd
}
