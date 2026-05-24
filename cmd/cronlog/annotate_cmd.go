package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/cronlog/internal/store"
)

// runAnnotate handles the `annotate` and `annotations` sub-commands.
//
//	cronlog annotate <id> <note>   — attach a note to a run
//	cronlog annotations <id>       — list notes for a run
func runAnnotate(db *sql.DB, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cronlog annotate <id> <note>\n       cronlog annotations <id>")
	}

	subcmd := args[0]

	switch subcmd {
	case "annotate":
		if len(args) < 3 {
			return fmt.Errorf("usage: cronlog annotate <id> <note>")
		}
		id, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid run id %q: %w", args[1], err)
		}
		note := strings.Join(args[2:], " ")
		if err := store.AddAnnotation(db, id, note); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "annotation added to run %d\n", id)
		return nil

	case "annotations":
		if len(args) < 2 {
			return fmt.Errorf("usage: cronlog annotations <id>")
		}
		id, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid run id %q: %w", args[1], err)
		}
		annotations, err := store.GetAnnotations(db, id)
		if err != nil {
			return err
		}
		if len(annotations) == 0 {
			fmt.Fprintln(os.Stdout, "no annotations for this run")
			return nil
		}
		for _, a := range annotations {
			fmt.Fprintf(os.Stdout, "[%s] %s\n", a.CreatedAt.Format("2006-01-02 15:04:05"), a.Note)
		}
		return nil

	default:
		return fmt.Errorf("unknown sub-command %q; expected 'annotate' or 'annotations'", subcmd)
	}
}
