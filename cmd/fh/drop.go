package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/metagram-net/firehose/drop"
	"github.com/metagram-net/firehose/moray"
	"github.com/metagram-net/firehose/null"
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

// TODO: extract the pattern for codegen
// res, err := fn(cmd.Context(), Client(), args)
// moray.Exit(io, res, err)

func dropEditCmd() *cobra.Command {
	var args struct {
		ID    null.UUID   `flag:"id,required" usage:"The Drop ID"`
		Title null.String `flag:"title" usage:"Set the title"`
		URL   null.String `flag:"url" usage:"Set the URL"`
		Tags  moray.UUIDs `flag:"tags" usage:"Set the tags"`
	}

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

			d, err := c.Drops.Update(ctx, drop.UpdateBody{
				ID:    args.ID.Value,
				Title: args.Title.Ptr(),
				URL:   args.URL.Ptr(),
				Tags:  args.Tags.Slice(),
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	moray.BindFlags(cmd, &args)
	return cmd
}

func dropMoveCmd() *cobra.Command {
	var args struct {
		ID     null.UUID   `flag:"id,required" usage:"The Drop ID"`
		Status drop.Status `flag:"status" usage:"Set the status"`
	}

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

			d, err := c.Drops.Move(ctx, drop.MoveBody{
				ID:     args.ID.Value,
				Status: args.Status,
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	moray.BindFlags(cmd, &args)
	return cmd
}

func dropDeleteCmd() *cobra.Command {
	var args struct {
		ID null.UUID `flag:"id,required" usage:"The Drop ID"`
	}

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

			d, err := c.Drops.Delete(ctx, drop.DeleteBody{
				ID: args.ID.Value,
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	moray.BindFlags(cmd, &args)
	return cmd
}

func dropGetCmd() *cobra.Command {
	var args struct {
		ID null.UUID `flag:"id,required" usage:"The Drop ID"`
	}

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

			d, err := c.Drops.Get(ctx, drop.GetParams{
				ID: drop.ID(args.ID.Value),
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	moray.BindFlags(cmd, &args)
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

			d, err := c.Drops.Next(ctx)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	return cmd
}

func dropListCmd() *cobra.Command {
	var args struct {
		Status drop.Status `flag:"status" usage:"The drop status"`
		Limit  null.Int32  `flag:"limit" usage:"The maximum number of drops"`
		Tags   moray.UUIDs `flag:"tags" usage:"List drops with these tags"`
	}

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

			if args.Status == drop.StatusUnknown {
				args.Status = drop.StatusUnread
			}
			res, err := c.Drops.List(ctx, drop.ListBody{
				Status: args.Status,
				Limit:  args.Limit.Ptr(),
				Tags:   args.Tags.Slice(),
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(res)
		},
	}
	moray.BindFlags(cmd, &args)
	return cmd
}
