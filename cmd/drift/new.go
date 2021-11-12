package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type wholeNum int64

func (n *wholeNum) String() string {
	if n == nil {
		return ""
	}
	return strconv.Itoa(int(*n))
}

func (n *wholeNum) Set(s string) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("not a valid integer: %w", err)
	}
	if i < 0 {
		return fmt.Errorf("cannot be negative: %s", s)
	}
	*n = wholeNum(i)
	return nil
}

func (i *wholeNum) Type() string {
	return "positive_integer"
}

func newCmd() *cobra.Command {
	var (
		// Set default ID out of range in case someone wants to create
		// migration 0.
		id   wholeNum = -1
		slug string
	)

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new migration file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			path, err := newFile(viper.GetString("migrations-dir"), id, slug)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(path)
		},
	}
	flags := cmd.Flags()
	flags.Var(&id, "id", "Migration ID override")
	flags.StringVar(&slug, "slug", "", "Short text describing the migration")
	cmd.MarkFlagRequired("slug")
	return cmd
}

func newFile(migrationsDir string, id wholeNum, slug string) (string, error) {
	if id == -1 {
		ts := time.Now().Unix()
		if ts < 0 {
			return "", fmt.Errorf("migration ID cannot be negative: %d", ts)
		}
		id = wholeNum(ts)
	}

	files, err := available(migrationsDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.ID == int(id) {
			return "", fmt.Errorf("migration %d already exists: %s", id, f.Name)
		}
	}

	slug = slugify(slug)
	name := fmt.Sprintf("%d-%s.sql", id, slug)
	path := filepath.Join(migrationsDir, name)
	return path, os.WriteFile(path, []byte(template), 0o644)
}

var template = "-- TODO: write your migration here\n"

var reSeparator = regexp.MustCompile(`[\-\s._/]+`)

func slugify(s string) string {
	return reSeparator.ReplaceAllString(s, "_")
}
