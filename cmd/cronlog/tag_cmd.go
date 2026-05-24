package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/cronlog/internal/store"
)

// runTag handles the `cronlog tag` sub-command.
//
// Usage:
//
//	cronlog tag <run-id> <tag1>[,tag2,...]
//	cronlog tag --list <run-id>
func runTag(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("tag", flag.ContinueOnError)
	list := fs.Bool("list", false, "list tags for a run instead of adding")
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *list {
		if fs.NArg() < 1 {
			return fmt.Errorf("tag --list requires a run ID")
		}
		id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid run ID %q: %w", fs.Arg(0), err)
		}
		tags, err := store.ListTags(db, id)
		if err != nil {
			return fmt.Errorf("list tags: %w", err)
		}
		if len(tags) == 0 {
			fmt.Println("(no tags)")
			return nil
		}
		for _, t := range tags {
			fmt.Println(t)
		}
		return nil
	}

	if fs.NArg() < 2 {
		return fmt.Errorf("usage: cronlog tag <run-id> <tag1>[,tag2,...]")
	}

	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid run ID %q: %w", fs.Arg(0), err)
	}

	tags := strings.Split(fs.Arg(1), ",")
	if err := store.AddTags(db, id, tags); err != nil {
		return fmt.Errorf("add tags: %w", err)
	}

	fmt.Printf("Tagged run %d with: %s\n", id, strings.Join(tags, ", "))
	return nil
}
