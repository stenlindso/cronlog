package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/user/cronlog/internal/store"
)

func runNotify(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("notify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	add := fs.String("add", "", "command to add a notification rule for")
	del := fs.String("delete", "", "rule ID to delete")
	list := fs.Bool("list", false, "list all notification rules")
	onFailure := fs.Bool("on-failure", false, "notify on failure (exit code != 0)")
	onSuccess := fs.Bool("on-success", false, "notify on success (exit code == 0)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	switch {
	case *add != "":
		if !*onFailure && !*onSuccess {
			return fmt.Errorf("specify --on-failure and/or --on-success")
		}
		id, err := store.AddNotifyRule(db, *add, *onFailure, *onSuccess)
		if err != nil {
			return fmt.Errorf("add notify rule: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Added notify rule id=%d for %q\n", id, *add)

	case *del != "":
		id, err := strconv.ParseInt(*del, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", *del, err)
		}
		if err := store.DeleteNotifyRule(db, id); err != nil {
			return fmt.Errorf("delete notify rule: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Deleted notify rule id=%d\n", id)

	case *list:
		rules, err := store.ListNotifyRules(db, "")
		if err != nil {
			return fmt.Errorf("list notify rules: %w", err)
		}
		if len(rules) == 0 {
			fmt.Fprintln(os.Stdout, "No notification rules configured.")
			return nil
		}
		fmt.Fprintf(os.Stdout, "%-6s %-40s %-10s %-10s\n", "ID", "COMMAND", "ON_FAILURE", "ON_SUCCESS")
		for _, r := range rules {
			fmt.Fprintf(os.Stdout, "%-6d %-40s %-10v %-10v\n",
				r.ID, r.Command, r.OnFailure, r.OnSuccess)
		}

	default:
		fs.Usage()
		return fmt.Errorf("specify --add, --delete, or --list")
	}
	return nil
}
