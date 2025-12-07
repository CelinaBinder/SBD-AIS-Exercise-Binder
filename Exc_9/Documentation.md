# 📄 **SBD Exercise 9 - MapReduce Word Frequency in Go**


## 📌 Overview

In this assignment, the MapReduce programming model was implemented in Go to compute word frequencies from *Marcus Aurelius — Meditations* (Project Gutenberg edition).
The task included:

* Reading the text file into memory
* Implementing a concurrent MapReduce pipeline
* Cleaning the text using regex (remove digits, punctuation, symbols)
* Making all operations concurrent (goroutines, channels, sync)
* Passing all provided unit tests
* Computing the final word-frequency distribution

This document summarizes the design, implementation, and results.

---

# 🧩 Project Structure

```
Exc_9/
 ├── go.mod
 ├── main.go
 ├── meditations.txt
 └── mapred/
       ├── interface.go
       ├── map_reduce.go
       ├── map_reduce_test.go
```

---

# ⚙️ Implementation Details

## 1. MapReduce Design

The MapReduce implementation follows the classical 3-stage architecture:

### **Mapper**

* Accepts a string (line of text)
* Removes non-letter characters using a regex:

  ```
  [^a-zA-Z]+ → " "
  ```
* Lowercases all words
* Emits one `KeyValue{word, 1}` per occurrence

### **Shuffle / Grouping**

* All mapper outputs are collected via a channel
* Key-value pairs are grouped into:

  ```
  map[string][]int
  ```
* This prepares the input for reducers

### **Reducer**

* One goroutine per key
* Sums up all values belonging to a word
* Returns a single `KeyValue{word, totalCount}`

### **Concurrency**

* Mapper stage uses one goroutine per input line
* Reducers run concurrently as well
* Synchronization handled using:

    * `sync.WaitGroup`
    * `sync.Mutex`
    * channels

---

# 📚 2. Source Code

## `mapred/map_reduce.go`

```go
package mapred

import (
	"regexp"
	"strings"
	"sync"
)

// MapReduce implements MapReduceInterface
type MapReduce struct{}

// wordCountMapper turns an input text into a slice of KeyValue pairs.
// It lowercases words and strips out non-letter characters.
func (mr MapReduce) wordCountMapper(text string) []KeyValue {
	// Replace any sequence of non-letters with a single space
	re := regexp.MustCompile(`[^a-zA-Z]+`)
	clean := re.ReplaceAllString(text, " ")
	fields := strings.Fields(clean)

	kvs := make([]KeyValue, 0, len(fields))
	for _, f := range fields {
		w := strings.ToLower(f)
		if w == "" {
			continue
		}
		kvs = append(kvs, KeyValue{Key: w, Value: 1})
	}
	return kvs
}

// wordCountReducer takes a key and a slice of ints (values) and sums them.
func (mr MapReduce) wordCountReducer(key string, values []int) KeyValue {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return KeyValue{Key: key, Value: sum}
}

// Run executes the MapReduce over the input slice of strings and returns final counts.
// Mappers and reducers run concurrently.
func (mr MapReduce) Run(input []string) map[string]int {
	// Channel to collect emitted key-value pairs from mappers
	kvCh := make(chan KeyValue)

	var mapWg sync.WaitGroup
	// Start mapper goroutines (one per input string)
	for _, txt := range input {
		mapWg.Add(1)
		go func(t string) {
			defer mapWg.Done()
			kvs := mr.wordCountMapper(t)
			for _, kv := range kvs {
				kvCh <- kv
			}
		}(txt)
	}

	// Close kvCh once all mappers are done
	go func() {
		mapWg.Wait()
		close(kvCh)
	}()

	// Shuffle: collect values per key
	groups := make(map[string][]int)
	for kv := range kvCh {
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}

	// Reducer stage: run reducer per key concurrently
	var reduceWg sync.WaitGroup
	result := make(map[string]int)
	var mu sync.Mutex

	for key, vals := range groups {
		reduceWg.Add(1)
		go func(k string, v []int) {
			defer reduceWg.Done()
			out := mr.wordCountReducer(k, v)
			mu.Lock()
			result[out.Key] = out.Value
			mu.Unlock()
		}(key, vals)
	}

	reduceWg.Wait()
	return result
}
```

---

## `main.go`

```go
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
	f, err := os.Open("meditations.txt")
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
```

---

# 🧪 3. Test Results

Running:

```
go test ./mapred
```

Output:

```
ok      exc9/mapred     0.668s
```

All tests passed successfully.

---

# 📊 4. Final Word Frequency Output

Top 40 most frequent words in *Meditations*:

```
  1. and             3334
  2. the             2808
  3. of              2530
  4. to              2078
  5. that            1957
  6. is              1458
  7. in              1195
  8. it              1192
  9. a               1150
 10. be              1004
 11. as              945
 12. or              850
 13. thou            821
 14. for             821
 15. not             730
 16. all             686
 17. which           612
 18. but             575
 19. he              571
 20. things          554
 21. with            522
 22. i               494
 23. by              479
 24. so              470
 25. this            457
 26. are             447
 27. his             447
 28. unto            439
 29. they            417
 30. what            380
 31. man             368
 32. if              366
 33. from            351
 34. thy             347
 35. any             346
 36. one             338
 37. thee            312
 38. them            308
 39. have            296
 40. nature          273
```


---

# 🏁 Conclusion

This exercise demonstrates:

* Successful implementation of the MapReduce paradigm in Go
* Correct use of goroutines, channels, and sync primitives
* Proper text preprocessing using regex
* Fully passing unit tests
* Real-world application: computing word frequencies from a classical text


