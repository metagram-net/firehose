package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/spf13/cobra"
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
		dropDeleteCmd(),
	)
	return cmd
}

func dropNewCmd() *cobra.Command {
	var request struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}

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

			d, err := c.Drops.Create(ctx, request.Title, request.URL)
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(d)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&request.Title, "title", "", "Set the title")
	flags.StringVar(&request.URL, "url", "", "Set the URL")
	cmd.MarkFlagRequired("url")
	return cmd
}

func dropEditCmd() *cobra.Command {
	var id string
	var request struct {
		Title  string `json:"title,omitempty"`
		URL    string `json:"url,omitempty"`
		Status string `json:"status,omitempty"`
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

			var reqBody bytes.Buffer
			if err := json.NewEncoder(&reqBody).Encode(request); err != nil {
				return err
			}

			res, err := c.Post(ctx, fmt.Sprintf("drops/update/%s", id), &reqBody)
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
	flags.StringVar(&request.Title, "title", "", "Set the title")
	flags.StringVar(&request.URL, "url", "", "Set the URL")
	flags.StringVar(&request.Status, "status", "", "Set the status")
	return cmd
}

func dropDeleteCmd() *cobra.Command {
	var id string

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

			res, err := c.Post(ctx, fmt.Sprintf("drops/delete/%s", id), nil)
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
	var request struct {
		// TODO: Status enum as a CLI arg?
		Status string `json:"status"`
		Limit  int    `json:"limit"`
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

			var reqBody bytes.Buffer
			if err := json.NewEncoder(&reqBody).Encode(request); err != nil {
				return err
			}

			res, err := c.Post(ctx, "drops/list", &reqBody)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			_, err = io.Copy(os.Stdout, res.Body)
			return err
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&request.Status, "status", "unread", "The drop status")
	flags.IntVar(&request.Limit, "limit", 5, "The number of drops to list")
	return cmd
}
