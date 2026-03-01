package checkpoint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GetPgUrl builds a PostgreSQL connection string for the given port.
func GetPgUrl(port int) string {
	return fmt.Sprintf("postgresql://bertie_user_backend:bertie_password_backend@localhost:%d/bertie_db_backend?sslmode=disable", port)
}

// getOrCreateCheckpointDir returns the checkpoint directory path, creating it if it doesn't exist.
func getOrCreateCheckpointDir() (string, error) {
	path := filepath.Join(os.TempDir(), "pgcheckpoint")
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

// getNextCheckpointFilePath returns the file path for the next checkpoint (latest + 1) in a dir.
func getNextCheckpointFilePath(largest int, dir string) (string, error) {
	filename := fmt.Sprintf("checkpoint_%d", largest+1)
	return filepath.Join(dir, filename), nil
}

// checkpointsToDelete returns a list of files eligable to be deleted from a list of existing files.
func checkpointsToDelete(filenames []string, latest int) []string {
	var toDelete []string
	for i, file := range filenames {
		if i < latest-1 {
			toDelete = append(toDelete, file)
		}
	}
	return toDelete
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
func ListCheckpointFilenames() ([]string, error) {
	dir, err := getOrCreateCheckpointDir()
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
func CreateCheckpoint(filename string, port int) (string, error) {
	dir, err := getOrCreateCheckpointDir()
	if err != nil {
		return "", err
	}

	files, err := getCheckpointFilenames(dir)
	if err != nil {
		return "", err
	}

	_, largest, err := getLatestCheckpoint(files)
	if err != nil {
		return "", err
	}

	path, err := getNextCheckpointFilePath(largest, dir)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("pg_dump", "--dbname", GetPgUrl(port), "--file", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, out)
	}

	return string(out), nil
}

// PruneCheckpoints deletes all checkpoints except the latest one, returning the number deleted.
func PruneCheckpoints() (int, error) {
	dir, err := getOrCreateCheckpointDir()
	if err != nil {
		return 0, err
	}

	files, err := getCheckpointFilenames(dir)
	if err != nil {
		return 0, err
	}

	if len(files) == 1 {
		return 0, nil
	}

	_, latest, err := getLatestCheckpoint(files)
	if err != nil {
		return 0, err
	}

	toDelete := checkpointsToDelete(files, latest)

	count := 0
	for _, file := range toDelete {
		os.Remove(filepath.Join(dir, file))
		count++
	}

	return count, nil
}
