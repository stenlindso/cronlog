package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/cronlog/internal/format"
	"github.com/user/cronlog/internal/store"
)

func runSearch(db *store.DB, args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	command := fs.String("command", "", "substring to match against command")
	exitCodeFlag := fs.Int("exit-code", -1, "filter by exit code (-1 = any)")
	limit := fs.Int("limit", 50, "maximum number of results")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *command == "" && *exitCodeFlag == -1 {
		fmt.Fprintln(os.Stderr, "cronlog search: provide at least --command or --exit-code")
		fs.Usage()
		return fmt.Errorf("no search criteria provided")
	}

	opts := store.SearchOptions{
		Command: *command,
		Limit:   *limit,
	}
	if *exitCodeFlag >= 0 {
		v := *exitCodeFlag
		opts.ExitCode = &v
	}

	runs, err := store.Search(db.DB, opts)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("no runs matched")
		return nil
	}

	format.PrintTable(os.Stdout, runs)
	return nil
}
