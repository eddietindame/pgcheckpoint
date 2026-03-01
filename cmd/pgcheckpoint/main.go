package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type defaults struct {
	port     int
	filename string
}

var defaultValues = defaults{
	port:     5432,
	filename: "checkpoint_1.sql",
}

func getPgUrl(port int) string {
	return fmt.Sprintf("postgresql://bertie_user_backend:bertie_password_backend@localhost:%d/bertie_db_backend?sslmode=disable", port)
}

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

func getCheckpointFilePath(filename string) (string, error) {
	dir, err := getOrCreateCheckpointDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}

func getNextCheckpointFilePath() (string, error) {
	files, err := getCheckpointFilenames()
	if err != nil {
		return "", err
	}

	largest := 0
	for _, file := range files {
		suffix := strings.TrimSuffix(strings.TrimPrefix(file, "checkpoint"), ".sql")
		intString := strings.TrimPrefix(suffix, "_")
		if intString == "" {
			intString = "0"
		}
		n, err := strconv.Atoi(intString)
		if err != nil {
			return "", fmt.Errorf("Error creating checkpoint: %w", err)
		}
		if n > largest {
			largest = n
		}
	}

	filename := fmt.Sprintf("checkpoint_%d", largest+1)
	return getCheckpointFilePath(filename)
}

func getCheckpointFilenames() ([]string, error) {
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

func createCheckpoint(filename string, port int) (string, error) {
	path, err := getNextCheckpointFilePath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("pg_dump", "--dbname", getPgUrl(port), "--file", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, out)
	}

	return string(out), nil
}

func main() {
	port := flag.Int("port", defaultValues.port, "Postgres port")
	fileName := flag.String("filename", defaultValues.filename, "Checkpoint filename")
	list := flag.Bool("list", false, "List mode")
	flag.Parse()

	if *list {
		files, err := getCheckpointFilenames()
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			fmt.Println(file)
		}

		return
	}

	fmt.Printf("Database url: %s\n", getPgUrl(*port))
	fmt.Println(os.TempDir())

	out, err := createCheckpoint(*fileName, *port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(out)
}
