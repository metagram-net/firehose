package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/moray"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Manage drops",
	}
	cmd.AddCommand(
		dropNewCmd(),
		dropGetCmd(),
		dropNextCmd(),
		dropListCmd(),
		dropEditCmd(),
		dropMoveCmd(),
		dropDeleteCmd(),
	)
	return cmd
}

func dropNewCmd() *cobra.Command {
	var body drop.CreateBody

	cmd := &cobra.Command{
		Use:          "new",
		Short:        "Create a new drop",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			d, err := c.Drops.Create(ctx, body)
			if err != nil {
				return err
			}
			// TODO: table encoder
			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&body.Title, "title", "", "Set the title")
	flags.StringVar(&body.URL, "url", "", "Set the URL")
	cmd.MarkFlagRequired("url")
	return cmd
}

func dropEditCmd() *cobra.Command {
	var body drop.UpdateBody

	cmd := &cobra.Command{
		Use:          "edit",
		Short:        "Edit a drop",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			d, err := c.Drops.Update(ctx, body)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	flags := cmd.Flags()
	flags.Var((*moray.UUID)(&body.ID), "id", "The drop ID")
	cmd.MarkFlagRequired("id")
	flags.Var(&body.Title, "title", "Set the title")
	flags.Var(&body.URL, "url", "Set the URL")
	return cmd
}

func dropMoveCmd() *cobra.Command {
	var body drop.MoveBody

	cmd := &cobra.Command{
		Use:          "move",
		Short:        "Move a drop to a different status",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			d, err := c.Drops.Move(ctx, body)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	flags := cmd.Flags()
	flags.Var((*moray.UUID)(&body.ID), "id", "The drop ID")
	cmd.MarkFlagRequired("id")
	flags.Var(&body.Status, "status", "Set the status")
	cmd.MarkFlagRequired("status")
	return cmd
}

func dropDeleteCmd() *cobra.Command {
	var body drop.DeleteBody

	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete a drop",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			d, err := c.Drops.Delete(ctx, body)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	flags := cmd.Flags()
	flags.Var((*moray.UUID)(&body.ID), "id", "The drop ID")
	cmd.MarkFlagRequired("id")
	return cmd
}

func dropGetCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:          "get",
		Short:        "Get a drop",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			res, err := c.Get(ctx, fmt.Sprintf("drops/get/%s", id))
			if err != nil {
				return err
			}
			defer res.Body.Close()

			_, err = io.Copy(os.Stdout, res.Body)
			return err
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&id, "id", "", "The drop ID")
	cmd.MarkFlagRequired("id")
	return cmd
}

func dropNextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "next",
		Short:        "Get the next unread drop",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			res, err := c.Get(ctx, "drops/next")
			if err != nil {
				return err
			}
			defer res.Body.Close()

			_, err = io.Copy(os.Stdout, res.Body)
			return err
		},
	}
	return cmd
}

func dropListCmd() *cobra.Command {
	var body drop.ListBody

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List drops",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			c, err := Client()
			if err != nil {
				return err
			}

			res, err := c.Drops.List(ctx, body)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(res)
		},
	}
	flags := cmd.Flags()
	body.Status = drop.StatusUnread
	flags.Var(&body.Status, "status", "The drop status")
	flags.Int32Var(&body.Limit, "limit", 5, "The number of drops to list")
	return cmd
}
