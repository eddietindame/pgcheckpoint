package checkpoint

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DefaultCheckpointDir returns the default checkpoint base directory (~/.pgcheckpoint/checkpoints).
func DefaultCheckpointDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return getCheckpointFilePath(home, ".pgcheckpoint/checkpoints")
}

// getOrCreateCheckpointDir returns the checkpoint directory path, creating it if it doesn't exist.
func getOrCreateCheckpointDir(baseDir, profile string) (string, error) {
	path := getCheckpointFilePath(baseDir, profile)
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
// If name is non-empty, it is sanitised and appended to the filename.
func CreateCheckpoint(url, baseDir, profile string, mode NamingMode, name string) (string, string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return "", "", err
	}

	var path string
	if mode.IsTimestampBased() {
		path = getNextTimestampCheckpointFilePath(dir, mode, name)
	} else {
		files, err := getCheckpointFilenames(dir)
		if err != nil {
			return "", "", err
		}

		_, largest, err := getLatestCheckpoint(files)
		if err != nil {
			return "", "", err
		}

		path = getNextCheckpointFilePath(largest, dir, name)
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

// DeleteCheckpoint removes a specific checkpoint file by name or short name.
func DeleteCheckpoint(baseDir, profile, target string, mode NamingMode) (string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return "", err
	}

	files, err := getCheckpointFilenames(dir)
	if err != nil {
		return "", err
	}

	filename, err := resolveCheckpointTarget(files, target, mode)
	if err != nil {
		return "", err
	}

	if err := os.Remove(getCheckpointFilePath(dir, filename)); err != nil {
		return "", fmt.Errorf("error deleting checkpoint: %w", err)
	}

	return filename, nil
}

// RenameCheckpoint changes the name portion of an existing checkpoint file.
// If newName is empty, the name is removed. Returns the new filename.
func RenameCheckpoint(baseDir, profile, target, newName string, mode NamingMode) (string, error) {
	dir, err := getOrCreateCheckpointDir(baseDir, profile)
	if err != nil {
		return "", err
	}

	files, err := getCheckpointFilenames(dir)
	if err != nil {
		return "", err
	}

	filename, err := resolveCheckpointTarget(files, target, mode)
	if err != nil {
		return "", err
	}

	oldPath := getCheckpointFilePath(dir, filename)
	id := extractCheckpointIdentifier(filename, mode)

	var newFilename string
	if newName != "" {
		newFilename = fmt.Sprintf("checkpoint_%s_%s.sql", id, sanitizeName(newName))
	} else {
		newFilename = fmt.Sprintf("checkpoint_%s.sql", id)
	}

	newPath := getCheckpointFilePath(dir, newFilename)
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("error renaming checkpoint: %w", err)
	}

	return newFilename, nil
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
		files, err := getCheckpointFilenames(dir)
		if err != nil {
			return "", "", err
		}
		filename, err = resolveCheckpointTarget(files, target, mode)
		if err != nil {
			return "", "", err
		}
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
