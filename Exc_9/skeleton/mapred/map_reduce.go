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
