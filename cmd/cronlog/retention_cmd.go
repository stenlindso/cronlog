package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/example/cronlog/internal/store"
)

// runRetention handles the "retention" subcommand, which manages data retention
// policies for cron run logs. Policies can be scoped to a specific command or
// applied globally (empty command). Use --add, --delete, or --apply to modify
// policies; with no flag the current policies are listed.
func runRetention(args []string) error {
	fs := flag.NewFlagSet("retention", flag.ContinueOnError)
	add := fs.Bool("add", false, "add or update a retention policy")
	del := fs.Bool("delete", false, "delete a retention policy")
	apply := fs.Bool("apply", false, "apply all retention policies now")
	command := fs.String("command", "", "command to scope policy (empty = global)")
	maxAge := fs.Int("max-age-days", 0, "delete runs older than N days")
	maxCount := fs.Int("keep-last", 0, "keep only the last N runs per command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *add && *del {
		return fmt.Errorf("--add and --delete are mutually exclusive")
	}

	db, err := store.Open(defaultDBPath())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	switch {
	case *add:
		if *maxAge == 0 && *maxCount == 0 {
			return fmt.Errorf("--add requires --max-age-days and/or --keep-last")
		}
		err = store.UpsertRetentionPolicy(db, store.RetentionPolicy{
			Command:    *command,
			MaxAgeDays: *maxAge,
			MaxCount:   *maxCount,
		})
		if err != nil {
			return err
		}
		fmt.Println("retention policy saved")

	case *del:
		if err := store.DeleteRetentionPolicy(db, *command); err != nil {
			return err
		}
		fmt.Println("retention policy deleted")

	case *apply:
		n, err := store.ApplyRetentionPolicies(db)
		if err != nil {
			return err
		}
		fmt.Printf("pruned %d run(s)\n", n)

	default:
		policies, err := store.ListRetentionPolicies(db)
		if err != nil {
			return err
		}
		if len(policies) == 0 {
			fmt.Println("no retention policies configured")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tMAX AGE (days)\tKEEP LAST")
		for _, p := range policies {
			cmd := p.Command
			if cmd == "" {
				cmd = "(global)"
			}
			fmt.Fprintf(w, "%s\t%d\t%d\n", cmd, p.MaxAgeDays, p.MaxCount)
		}
		w.Flush()
	}
	return nil
}
