package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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
	)
	return cmd
}

func urlJoin(base string, parts ...string) (*url.URL, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	for _, p := range parts {
		u, err = u.Parse(p)
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}

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
	flags.StringVar(&request.Title, "title", "", "The title to use for this drop")
	flags.StringVar(&request.URL, "url", "", "The URL to use for this drop")
	cmd.MarkFlagRequired("url")
	return cmd
}
