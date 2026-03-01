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

// getCheckpointFilePath returns the full path for a checkpoint file.
func getCheckpointFilePath(filename string) (string, error) {
	dir, err := getOrCreateCheckpointDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
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

// getLatestCheckpoint finds the checkpoint with the highest number and returns its name and number.
func getLatestCheckpoint() (string, int, error) {
	files, err := GetCheckpointFilenames()
	if err != nil {
		return "", 0, err
	}

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

	return fmt.Sprintf("checkpoint_%d", largest), largest, nil
}

// getNextCheckpointFilePath returns the file path for the next checkpoint (latest + 1).
func getNextCheckpointFilePath() (string, error) {
	_, largest, err := getLatestCheckpoint()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("checkpoint_%d", largest+1)
	return getCheckpointFilePath(filename)
}

// GetCheckpointFilenames returns a list of all checkpoint filenames in the checkpoint directory.
func GetCheckpointFilenames() ([]string, error) {
	path, err := getOrCreateCheckpointDir()
	if err != nil {
		return []string{}, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return []string{}, fmt.Errorf("Error reading checkpoint dir: %w", err)
	}

	var files []string
	for _, file := range entries {
		files = append(files, file.Name())
	}

	return files, nil
}

// CreateCheckpoint runs pg_dump to create a new checkpoint SQL file.
func CreateCheckpoint(filename string, port int) (string, error) {
	path, err := getNextCheckpointFilePath()
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
	files, err := GetCheckpointFilenames()
	if err != nil {
		return 0, err
	}

	if len(files) == 1 {
		return 0, nil
	}

	_, n, err := getLatestCheckpoint()
	if err != nil {
		return 0, err
	}

	dir, err := getOrCreateCheckpointDir()
	if err != nil {
		return 0, err
	}

	count := 0
	for m, file := range files {
		if m < n {
			os.Remove(filepath.Join(dir, file))
			count++
		}
	}

	return count, nil
}
