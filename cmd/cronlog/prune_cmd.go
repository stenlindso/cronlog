package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/example/cronlog/internal/store"
)

// runPrune handles the `cronlog prune` sub-command.
// Usage:
//
//	cronlog prune --keep-last N
//	cronlog prune --older-than-days N
func runPrune(db *sql.DB, args []string) error {
	opts := store.PruneOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--keep-last":
			if i+1 >= len(args) {
				return fmt.Errorf("--keep-last requires a value")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				return fmt.Errorf("--keep-last must be a positive integer")
			}
			opts.KeepLast = n
		case "--older-than-days":
			if i+1 >= len(args) {
				return fmt.Errorf("--older-than-days requires a value")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				return fmt.Errorf("--older-than-days must be a positive integer")
			}
			opts.OlderThanDays = n
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	if opts.KeepLast == 0 && opts.OlderThanDays == 0 {
		fmt.Fprintln(os.Stderr, "prune: nothing to do — specify --keep-last or --older-than-days")
		return nil
	}

	deleted, err := store.Prune(db, opts)
	if err != nil {
		return fmt.Errorf("prune: %w", err)
	}

	fmt.Printf("Pruned %d run(s).\n", deleted)
	return nil
}
