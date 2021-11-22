package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Manage drops",
	}
	cmd.AddCommand(
		dropRandomCmd(),
		dropNewCmd(),
		dropGetCmd(),
		dropNextCmd(),
		dropEditCmd(),
		dropDeleteCmd(),
	)
	return cmd
}

func urlJoin(parts ...string) (*url.URL, error) {
	return url.Parse(strings.Join(parts, "/"))
}

// TODO: There's a _lot_ of deduplication to be done in here.

func dropRandomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "random",
		Short:        "Get a random drop",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			url, err := urlJoin(viper.GetString("url-base"), "drops/random")
			if err != nil {
				return err
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(viper.GetString("user-id"), viper.GetString("api-key"))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(body))
			return nil
		},
	}
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

			url, err := urlJoin(viper.GetString("url-base"), "drops/create")
			if err != nil {
				return err
			}

			var reqBody bytes.Buffer
			if err := json.NewEncoder(&reqBody).Encode(request); err != nil {
				return err
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), &reqBody)
			if err != nil {
				return err
			}
			req.SetBasicAuth(viper.GetString("user-id"), viper.GetString("api-key"))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(resBody))
			return nil
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

			url, err := urlJoin(viper.GetString("url-base"), "drops/update", id)
			if err != nil {
				return err
			}

			var reqBody bytes.Buffer
			if err := json.NewEncoder(&reqBody).Encode(request); err != nil {
				return err
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), &reqBody)
			if err != nil {
				return err
			}
			req.SetBasicAuth(viper.GetString("user-id"), viper.GetString("api-key"))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(resBody))
			return nil
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

			url, err := urlJoin(viper.GetString("url-base"), "drops/delete", id)
			if err != nil {
				return err
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(viper.GetString("user-id"), viper.GetString("api-key"))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(resBody))
			return nil
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

			url, err := urlJoin(viper.GetString("url-base"), "drops/get", id)
			if err != nil {
				return err
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(viper.GetString("user-id"), viper.GetString("api-key"))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(resBody))
			return nil
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

			url, err := urlJoin(viper.GetString("url-base"), "drops/next")
			if err != nil {
				return err
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(viper.GetString("user-id"), viper.GetString("api-key"))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(resBody))
			return nil
		},
	}
	return cmd
}
