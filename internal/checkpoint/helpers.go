package checkpoint

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// parseCheckpointNumber extracts the number from the end of a checkpoint file filename.
func parseCheckpointNumber(filename string) (int, error) {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint"), ".sql")
	intString := strings.TrimPrefix(suffix, "_")
	if intString == "" {
		intString = "0"
	}
	n, err := strconv.Atoi(intString)
	if err != nil {
		return 0, fmt.Errorf("error parsing checkpoint number: %w", err)
	}
	return n, nil
}

// getLatestCheckpoint returns the latest checkpoint in a list of checkpoint file names.
func getLatestCheckpoint(files []string) (string, int, error) {
	largest := 0
	for _, file := range files {
		n, err := parseCheckpointNumber(file)
		if err != nil {
			return "", 0, fmt.Errorf("error finding latest checkpoint: %w", err)
		}
		if n > largest {
			largest = n
		}
	}

	return fmt.Sprintf("checkpoint_%d.sql", largest), largest, nil
}

// getCheckpointFilePath returns the file path for a checkpoint in a dir.
func getCheckpointFilePath(dir, filename string) string {
	return filepath.Join(dir, filename)
}

// getNextCheckpointFilePath returns the file path for the next checkpoint (latest + 1) in a dir.
func getNextCheckpointFilePath(largest int, dir string) string {
	filename := fmt.Sprintf("checkpoint_%d.sql", largest+1)
	return getCheckpointFilePath(dir, filename)
}

// checkpointsToDelete returns a list of files eligible to be deleted from a list of existing files.
func checkpointsToDelete(filenames []string, latest int) ([]string, error) {
	var toDelete []string
	for _, file := range filenames {
		n, err := parseCheckpointNumber(file)
		if err == nil && n < latest {
			toDelete = append(toDelete, file)
		} else if err != nil {
			return []string{}, err
		}
	}

	return toDelete, nil
}
