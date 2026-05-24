package store

import (
	"database/sql"
	"time"
)

// Streak holds consecutive success/failure run information for a command.
type Streak struct {
	Command        string
	CurrentStreak  int
	StreakType     string // "success" or "failure"
	LongestSuccess int
	LongestFailure int
	LastRunAt      time.Time
}

// GetStreak returns the current and longest streaks for a given command.
func GetStreak(db *sql.DB, command string) (*Streak, error) {
	rows, err := db.Query(`
		SELECT command, exit_code, started_at
		FROM runs
		WHERE command = ?
		ORDER BY started_at DESC
	`, command)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type entry struct {
		exitCode  int
		startedAt time.Time
	}

	var entries []entry
	for rows.Next() {
		var cmd string
		var e entry
		if err := rows.Scan(&cmd, &e.exitCode, &e.startedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, sql.ErrNoRows
	}

	s := &Streak{Command: command, LastRunAt: entries[0].startedAt}

	// Determine current streak type from the most recent run.
	if entries[0].exitCode == 0 {
		s.StreakType = "success"
	} else {
		s.StreakType = "failure"
	}

	curLen, longestS, longestF := 0, 0, 0
	segType := s.StreakType
	curLen = 0

	for _, e := range entries {
		runType := "failure"
		if e.exitCode == 0 {
			runType = "success"
		}
		if runType == segType {
			curLen++
		} else {
			if segType == "success" && curLen > longestS {
				longestS = curLen
			} else if segType == "failure" && curLen > longestF {
				longestF = curLen
			}
			segType = runType
			curLen = 1
		}
	}
	// flush last segment
	if segType == "success" && curLen > longestS {
		longestS = curLen
	} else if segType == "failure" && curLen > longestF {
		longestF = curLen
	}

	s.CurrentStreak = func() int {
		count := 0
		for _, e := range entries {
			rt := "failure"
			if e.exitCode == 0 {
				rt = "success"
			}
			if rt == s.StreakType {
				count++
			} else {
				break
			}
		}
		return count
	}()

	s.LongestSuccess = longestS
	s.LongestFailure = longestF
	return s, nil
}
