package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"pgcheckpoint/internal/checkpoint"
)

var defaultValues = struct {
	port     int
	filename string
}{
	port:     5432,
	filename: "checkpoint_1.sql",
}

func checkDependencies() error {
	deps := []string{"pg_dump", "psql"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("%s not found in PATH", dep)
		}
	}
	return nil
}

func main() {
	port := flag.Int("port", defaultValues.port, "Postgres port")
	fileName := flag.String("filename", defaultValues.filename, "Checkpoint filename")
	list := flag.Bool("list", false, "List mode")
	prune := flag.Bool("prune", false, "Prune mode")
	flag.Parse()

	if *list {
		files, err := checkpoint.ListCheckpointFilenames()
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			fmt.Println(file)
		}

		return
	}

	if err := checkDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *prune {
		count, err := checkpoint.PruneCheckpoints()
		if err != nil {
			fmt.Printf("Error pruning checkpoints: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Checkpoints pruned: %d\n", count)
		return
	}

	fmt.Printf("Database url: %s\n", checkpoint.GetPgUrl(*port))

	out, err := checkpoint.CreateCheckpoint(*fileName, *port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(out)
}
