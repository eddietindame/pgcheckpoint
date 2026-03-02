package checkpoint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DefaultCheckpointDir returns the default checkpoint base directory (~/.pgcheckpoint/checkpoints).
func DefaultCheckpointDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".pgcheckpoint", "checkpoints")
}

// getOrCreateCheckpointDir returns the checkpoint directory path, creating it if it doesn't exist.
func getOrCreateCheckpointDir(baseDir, profile string) (string, error) {
	path := filepath.Join(baseDir, profile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Path %s does not exist, creating path...\n", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return "", fmt.Errorf("Error creating path %s: %w\n", path, err)
		}
		fmt.Println("Path created successfuly!")
	} else if err != nil {
		return "", fmt.Errorf("Error reading checkpoint dir")
	}

	return path, nil
}

// ParseCheckpointNumber extracts the number from the end of a checkpoint file filename.
func parseCheckpointNumber(filename string) (int, error) {
	suffix := strings.TrimSuffix(strings.TrimPrefix(filename, "checkpoint"), ".sql")
	intString := strings.TrimPrefix(suffix, "_")
	if intString == "" {
		intString = "0"
	}
	n, err := strconv.Atoi(intString)
	if err != nil {
		return 0, fmt.Errorf("Error parsing checkpoint number: %w", err)
	}
	return n, nil
}

// getLatestCheckpoint returns the latest checkpoint in a list of checkpoint file names.
func getLatestCheckpoint(files []string) (string, int, error) {
	largest := 0
	for _, file := range files {
		n, err := parseCheckpointNumber(file)
		if err != nil {
			return "", 0, fmt.Errorf("Error creating checkpoint: %w", err)
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

// checkpointsToDelete returns a list of files eligable to be deleted from a list of existing files.
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

// GetCheckpointFilenames returns a list of all checkpoint filenames in the provided dir.
func getCheckpointFilenames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, fmt.Errorf("Error reading checkpoint dir: %w", err)
	}

	var files []string
	for _, file := range entries {
		files = append(files, file.Name())
	}

	return files, nil
}

// ListCheckpointFilenames returns a list of all checkpoint filenames in the checkpoint directory.
func ListCheckpointFilenames(baseDir, profile string) ([]string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return []string{}, err
	}

	files, err := getCheckpointFilenames(dir)
	if err != nil {
		return []string{}, err
	}

	return files, nil
}

// CreateCheckpoint runs pg_dump to create a new checkpoint SQL file.
func CreateCheckpoint(filename, url, baseDir, profile, namingMode string) (string, string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return "", "", err
	}

	var path string
	if namingMode == "timestamp" {
		path = getNextTimestampCheckpointFilePath(dir)
	} else {
		files, err := getCheckpointFilenames(dir)
		if err != nil {
			return "", "", err
		}

		_, largest, err := getLatestCheckpoint(files)
		if err != nil {
			return "", "", err
		}

		path = getNextCheckpointFilePath(largest, dir)
	}
	cmd := exec.Command("pg_dump", "--dbname", url, "--file", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", err, out)
	}

	return string(out), path, nil
}

// PruneCheckpoints deletes all checkpoints except the latest one, returning the number deleted.
func PruneCheckpoints(baseDir, profile, namingMode string) (int, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return 0, err
	}

	files, err := getCheckpointFilenames(dir)
	if err != nil {
		return 0, err
	}

	if len(files) <= 1 {
		return 0, nil
	}

	var toDelete []string
	if namingMode == "timestamp" {
		_, latest, err := getLatestTimestampCheckpoint(files)
		if err != nil {
			return 0, err
		}
		toDelete, err = timestampCheckpointsToDelete(files, latest)
		if err != nil {
			return 0, err
		}
	} else {
		_, latest, err := getLatestCheckpoint(files)
		if err != nil {
			return 0, err
		}
		toDelete, err = checkpointsToDelete(files, latest)
		if err != nil {
			return 0, err
		}
	}

	count := 0
	for _, file := range toDelete {
		os.Remove(getCheckpointFilePath(dir, file))
		count++
	}

	return count, nil
}

// RestoreCheckpoint restores the configured database to the state stored in the provided target checkpoint. Defaults to latest checkpoint.
func RestoreCheckpoint(url, baseDir, profile, target, namingMode string) (string, string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return "", "", err
	}

	var filename string
	if target != "" {
		path := getCheckpointFilePath(dir, target)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", "", fmt.Errorf("checkpoint %q not found", target)
		}
		filename = target
	} else {
		files, err := getCheckpointFilenames(dir)
		if err != nil {
			return "", "", err
		}
		if namingMode == "timestamp" {
			filename, _, err = getLatestTimestampCheckpoint(files)
		} else {
			filename, _, err = getLatestCheckpoint(files)
		}
		if err != nil {
			return "", "", err
		}
	}

	cmd := exec.Command("psql", url, "-c", "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", "", fmt.Errorf("%w: %s", err, out)
	}

	cmd = exec.Command("psql", "--dbname", url, "--file", getCheckpointFilePath(dir, filename))
	out, err = cmd.CombinedOutput()

	if err != nil {
		return "", "", fmt.Errorf("%w: %s", err, out)
	}

	return string(out), filename, nil
}
