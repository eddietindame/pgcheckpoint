package main

import (
	"flag"
	"fmt"
	"pgcheckpoint/internal/checkpoint"
)

type defaults struct {
	port     int
	filename string
}

var defaultValues = defaults{
	port:     5432,
	filename: "checkpoint_1.sql",
}

func main() {
	port := flag.Int("port", defaultValues.port, "Postgres port")
	fileName := flag.String("filename", defaultValues.filename, "Checkpoint filename")
	list := flag.Bool("list", false, "List mode")
	prune := flag.Bool("prune", false, "Prune mode")
	flag.Parse()

	if *list {
		files, err := checkpoint.GetCheckpointFilenames()
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			fmt.Println(file)
		}

		return
	}

	if *prune {
		count, err := checkpoint.PruneCheckpoints()
		if err != nil {
			fmt.Printf("Error pruning checkpoints: %v\n", err)
		}
		fmt.Printf("Checkpoints pruned: %d\n", count)
		return
	}

	fmt.Printf("Database url: %s\n", checkpoint.GetPgUrl(*port))

	out, err := checkpoint.CreateCheckpoint(*fileName, *port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(out)
}
