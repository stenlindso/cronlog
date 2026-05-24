package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/example/cronlog/internal/store"
)

// runAlert handles the `cronlog alert` subcommand.
// Usage:
//
//	cronlog alert list [--command <cmd>]
//	cronlog alert delete <id>
func runAlert(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("alert subcommand required: list, delete")
	}
	switch args[0] {
	case "list":
		return runAlertList(db, args[1:])
	case "delete":
		return runAlertDelete(db, args[1:])
	default:
		return fmt.Errorf("unknown alert subcommand: %s", args[0])
	}
}

func runAlertList(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("alert list", flag.ContinueOnError)
	command := fs.String("command", "", "filter by command")
	limit := fs.Int("limit", 50, "max results")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	alerts, err := store.ListAlerts(db, *command, *limit)
	if err != nil {
		return fmt.Errorf("list alerts: %w", err)
	}
	if len(alerts) == 0 {
		fmt.Println("no alerts found")
		return nil
	}
	fmt.Printf("%-6s %-8s %-30s %-14s %-10s %s\n",
		"ID", "RUN_ID", "COMMAND", "CONDITION", "CHANNEL", "SENT_AT")
	for _, a := range alerts {
		cmd := a.Command
		if len(cmd) > 28 {
			cmd = cmd[:25] + "..."
		}
		fmt.Printf("%-6d %-8d %-30s %-14s %-10s %s\n",
			a.ID, a.RunID, cmd, a.Condition, a.Channel,
			a.SentAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func runAlertDelete(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("alert delete requires an <id> argument")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", args[0], err)
	}
	if err := store.DeleteAlert(db, id); err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}
	fmt.Printf("alert %d deleted\n", id)
	return nil
}
