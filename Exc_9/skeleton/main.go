package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"

	"exc9/mapred"
)

func main() {
	// Open the Meditations text file
	f, err := os.Open("res/meditations.txt")
	if err != nil {
		log.Fatalf("failed to open meditations.txt: %v", err)
	}
	defer f.Close()

	// Read file into []string, one line per entry
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	// Run MapReduce
	var mr mapred.MapReduce
	results := mr.Run(lines)

	// Convert results to sortable slice
	type kv struct {
		k string
		v int
	}
	var list []kv
	for k, v := range results {
		list = append(list, kv{k, v})
	}

	// Sort by frequency descending
	sort.Slice(list, func(i, j int) bool {
		return list[i].v > list[j].v
	})

	// Print top N words
	N := 40
	if len(list) < N {
		N = len(list)
	}

	fmt.Printf("Top %d most frequent words in Meditations:\n", N)
	for i := 0; i < N; i++ {
		fmt.Printf("%3d. %-15s %d\n", i+1, list[i].k, list[i].v)
	}
}
