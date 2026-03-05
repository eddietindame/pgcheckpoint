package checkpoint

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// sanitizeName normalises a human-readable checkpoint name for use in filenames.
// It lowercases, replaces spaces/underscores with hyphens, strips non-alphanumeric
// characters (except hyphens), collapses consecutive hyphens, and trims leading/trailing hyphens.
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.NewReplacer(" ", "-", "_", "-").Replace(name)
	name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}

// extractLeadingDigits returns the leading digit prefix of s.
func extractLeadingDigits(s string) string {
	for i, c := range s {
		if c < '0' || c > '9' {
			return s[:i]
		}
	}
	return s
}

// parseCheckpointNumber extracts the number from the end of a checkpoint file filename.
func parseCheckpointNumber(filename string) (int, error) {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint"), ".sql")
	intString := strings.TrimPrefix(suffix, "_")
	if intString == "" {
		intString = "0"
	}
	// Strip any trailing _name portion by taking only leading digits.
	intString = extractLeadingDigits(intString)
	if intString == "" {
		return 0, fmt.Errorf("error parsing checkpoint number: no digits found")
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
	latestFile := ""
	for _, file := range files {
		n, err := parseCheckpointNumber(file)
		if err != nil {
			return "", 0, fmt.Errorf("error finding latest checkpoint: %w", err)
		}
		if n > largest {
			largest = n
			latestFile = file
		}
	}
	if latestFile == "" && len(files) > 0 {
		latestFile = files[0]
	}

	return latestFile, largest, nil
}

// getCheckpointFilePath returns the file path for a checkpoint in a dir.
func getCheckpointFilePath(dir, filename string) string {
	return filepath.Join(dir, filename)
}

// getNextCheckpointFilePath returns the file path for the next checkpoint (latest + 1) in a dir.
// If name is non-empty it is appended as checkpoint_{N}_{name}.sql.
func getNextCheckpointFilePath(largest int, dir, name string) string {
	var filename string
	if name != "" {
		filename = fmt.Sprintf("checkpoint_%d_%s.sql", largest+1, sanitizeName(name))
	} else {
		filename = fmt.Sprintf("checkpoint_%d.sql", largest+1)
	}
	return getCheckpointFilePath(dir, filename)
}

// extractCheckpointIdentifier returns the identifier portion of a checkpoint filename
// (e.g. "3", "2026-03-02_15-30-45", "1740934245") based on the naming mode.
func extractCheckpointIdentifier(filename string, mode NamingMode) string {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint_"), ".sql")
	if mode == NamingModeTimestamp || mode == NamingModeCompact {
		format := timestampFormatForMode(mode)
		if len(suffix) > len(format) {
			return suffix[:len(format)]
		}
		return suffix
	}
	// Sequential and Unix: leading digits
	return extractLeadingDigits(suffix)
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
