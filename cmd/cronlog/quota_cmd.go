package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/yourusername/cronlog/internal/store"
)

func runQuota(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("quota", flag.ContinueOnError)
	add := fs.Bool("add", false, "add or update a quota rule")
	del := fs.Bool("delete", false, "delete a quota rule")
	list := fs.Bool("list", false, "list quota rules")
	check := fs.Bool("check", false, "check if a command is within quota")
	command := fs.String("command", "", "command to apply quota to")
	maxRuns := fs.Int("max-runs", 0, "maximum runs allowed in the window")
	window := fs.Int("window", 0, "window size in seconds")

	if err := fs.Parse(args); err != nil {
		return err
	}

	switch {
	case *list:
		return quotaList(db)
	case *add:
		if *command == "" || *maxRuns == 0 || *window == 0 {
			return fmt.Errorf("--add requires --command, --max-runs, and --window")
		}
		if err := store.UpsertQuota(db, *command, *maxRuns, *window); err != nil {
			return err
		}
		fmt.Printf("quota set: %s — %d runs per %ds\n", *command, *maxRuns, *window)
		return nil
	case *del:
		if *command == "" {
			return fmt.Errorf("--delete requires --command")
		}
		if err := store.DeleteQuota(db, *command); err != nil {
			return err
		}
		fmt.Printf("quota deleted for %s\n", *command)
		return nil
	case *check:
		if *command == "" {
			return fmt.Errorf("--check requires --command")
		}
		v, err := store.CheckQuota(db, *command)
		if err != nil {
			return err
		}
		if v != nil {
			fmt.Fprintln(os.Stderr, "QUOTA EXCEEDED:", v.Error())
			os.Exit(1)
		}
		fmt.Printf("ok: %s is within quota\n", *command)
		return nil
	default:
		return fmt.Errorf("specify --list, --add, --delete, or --check")
	}
}

func quotaList(db *sql.DB) error {
	quotas, err := store.ListQuotas(db)
	if err != nil {
		return err
	}
	if len(quotas) == 0 {
		fmt.Println("no quota rules defined")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "COMMAND\tMAX RUNS\tWINDOW (s)")
	for _, q := range quotas {
		fmt.Fprintf(w, "%s\t%s\t%s\n", q.Command, strconv.Itoa(q.MaxRuns), strconv.Itoa(q.WindowSec))
	}
	return w.Flush()
}
