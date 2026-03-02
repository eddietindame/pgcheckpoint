package checkpoint

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// timestamp naming mode: checkpoint_2026-03-02_15-30-45.sql
	timestampFormat = "2006-01-02_15-04-05"
	// compact naming mode: checkpoint_20260302T153045.sql
	compactFormat = "20060102T150405"
)

// timestampFormatForMode returns the time format string for a given naming mode.
func timestampFormatForMode(mode NamingMode) string {
	if mode == NamingModeCompact {
		return compactFormat
	}
	return timestampFormat
}

// parseCheckpointTimestamp extracts the timestamp from a checkpoint filename using the given format.
func parseCheckpointTimestamp(filename, format string) (time.Time, error) {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint_"), ".sql")
	t, err := time.Parse(format, suffix)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing checkpoint timestamp: %w", err)
	}
	return t, nil
}

// parseCheckpointUnix extracts the unix timestamp from a checkpoint filename.
func parseCheckpointUnix(filename string) (time.Time, error) {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint_"), ".sql")
	n, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing checkpoint unix timestamp: %w", err)
	}
	return time.Unix(n, 0), nil
}

// parseCheckpointTime parses a checkpoint filename into a time.Time for the given naming mode.
func parseCheckpointTime(filename string, mode NamingMode) (time.Time, error) {
	if mode == NamingModeUnix {
		return parseCheckpointUnix(filename)
	}
	return parseCheckpointTimestamp(filename, timestampFormatForMode(mode))
}

// getLatestTimestampCheckpoint returns the checkpoint with the newest time for the given naming mode.
func getLatestTimestampCheckpoint(files []string, mode NamingMode) (string, time.Time, error) {
	var latest time.Time
	var latestFile string
	for _, file := range files {
		t, err := parseCheckpointTime(file, mode)
		if err != nil {
			return "", time.Time{}, err
		}
		if t.After(latest) {
			latest = t
			latestFile = file
		}
	}
	return latestFile, latest, nil
}

// getNextTimestampCheckpointFilePath generates a checkpoint path using the current time and naming mode.
func getNextTimestampCheckpointFilePath(dir string, mode NamingMode) string {
	now := time.Now()
	var suffix string
	if mode == NamingModeUnix {
		suffix = strconv.FormatInt(now.Unix(), 10)
	} else {
		suffix = now.Format(timestampFormatForMode(mode))
	}
	filename := fmt.Sprintf("checkpoint_%s.sql", suffix)
	return getCheckpointFilePath(dir, filename)
}

// timestampCheckpointsToDelete returns all checkpoint files except the one matching latest,
// parsing filenames according to the given naming mode.
func timestampCheckpointsToDelete(filenames []string, latest time.Time, mode NamingMode) ([]string, error) {
	var toDelete []string
	for _, file := range filenames {
		t, err := parseCheckpointTime(file, mode)
		if err != nil {
			return nil, err
		}
		if !t.Equal(latest) {
			toDelete = append(toDelete, file)
		}
	}
	return toDelete, nil
}
