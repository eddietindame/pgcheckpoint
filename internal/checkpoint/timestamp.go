package checkpoint

import (
	"fmt"
	"strings"
	"time"
)

const timestampFormat = "2006-01-02_15-04-05"

// parseCheckpointTimestamp extracts the timestamp from a checkpoint filename.
func parseCheckpointTimestamp(filename string) (time.Time, error) {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint_"), ".sql")
	t, err := time.Parse(timestampFormat, suffix)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing checkpoint timestamp: %w", err)
	}
	return t, nil
}

// getLatestTimestampCheckpoint returns the checkpoint with the newest timestamp.
func getLatestTimestampCheckpoint(files []string) (string, time.Time, error) {
	var latest time.Time
	var latestFile string
	for _, file := range files {
		t, err := parseCheckpointTimestamp(file)
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

// getNextTimestampCheckpointFilePath generates a checkpoint path using the current time.
func getNextTimestampCheckpointFilePath(dir string) string {
	filename := fmt.Sprintf("checkpoint_%s.sql", time.Now().Format(timestampFormat))
	return getCheckpointFilePath(dir, filename)
}

// timestampCheckpointsToDelete returns all checkpoint files except the one matching latest.
func timestampCheckpointsToDelete(filenames []string, latest time.Time) ([]string, error) {
	var toDelete []string
	for _, file := range filenames {
		t, err := parseCheckpointTimestamp(file)
		if err != nil {
			return nil, err
		}
		if !t.Equal(latest) {
			toDelete = append(toDelete, file)
		}
	}
	return toDelete, nil
}
