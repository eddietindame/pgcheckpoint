package checkpoint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
			return "", fmt.Errorf("error creating path %s: %w", path, err)
		}
		fmt.Println("Path created successfully!")
	} else if err != nil {
		return "", fmt.Errorf("error reading checkpoint dir: %w", err)
	}

	return path, nil
}

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

// getCheckpointFilenames returns a list of checkpoint filenames (checkpoint_*.sql) in the provided dir.
func getCheckpointFilenames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, fmt.Errorf("error reading checkpoint dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "checkpoint_") && strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
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

// CreateCheckpoint runs pg_dump to create a new checkpoint SQL file. The mode parameter
// controls the checkpoint filename format (sequential, timestamp, compact, or unix).
func CreateCheckpoint(url, baseDir, profile string, mode NamingMode) (string, string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return "", "", err
	}

	var path string
	if mode.IsTimestampBased() {
		path = getNextTimestampCheckpointFilePath(dir, mode)
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
// The mode parameter determines how checkpoint filenames are parsed to find the latest.
func PruneCheckpoints(baseDir, profile string, mode NamingMode) (int, error) {
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
	if mode.IsTimestampBased() {
		_, latest, err := getLatestTimestampCheckpoint(files, mode)
		if err != nil {
			return 0, err
		}
		toDelete, err = timestampCheckpointsToDelete(files, latest, mode)
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

// RestoreCheckpoint restores the configured database to the state stored in the provided target
// checkpoint. If no target is given, the latest checkpoint is used based on the mode.
func RestoreCheckpoint(url, baseDir, profile, target string, mode NamingMode) (string, string, error) {
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
		if mode.IsTimestampBased() {
			filename, _, err = getLatestTimestampCheckpoint(files, mode)
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

// TODO: add unit tests for CreateCheckpoint, PruneCheckpoints, and RestoreCheckpoint
// using temp directories and mocked exec.Command calls.
