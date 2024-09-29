package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/calebglawson/snowberry"
	"github.com/fhalim/csvreader"
)

func main() {
	step := 10
	var scoreThreshold float32 = 0.70

	in := make(chan string)
	out := make(chan map[string]int)

	/* Uncomment and pass `debug` into `Counter` to print debug lines

	debug := make(chan *snowberry.AssignDebug, 1000)
	go func() {
		for d := range debug {
			log.Printf(
				"DEBUG - Counter ID: %s, Input: %s, Masked Input: %s, Best Match: %s, Best Match Masked: %s, Best Match Score: %2f, Match Accepted: %v",
				d.CounterID,
				d.Input,
				d.MaskedInput,
				d.BestMatch,
				d.BestMatchMasked,
				d.BestMatchScore,
				d.BestMatchAccepted,
			)
		}
	}()
	*/

	var wg sync.WaitGroup
	for i := 0; i <= 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := snowberry.NewCounter(step, scoreThreshold, nil, nil, nil)

			for w := range in {
				c.Assign(w)
			}

			out <- c.Counts()
		}()
	}

	f, err := os.Open("examples/sample_data/one.csv")
	if err != nil {
		log.Fatal(err)
	}

	r, err := csvreader.NewReader(context.Background(), f)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	dataCount := 0
	for row := range r.Data {
		dataCount++
		in <- row["sentence"]
	}
	close(in)

	// join thread results
	go func() {
		wg.Wait()
		close(out)
	}()

	c := snowberry.NewCounter(step, scoreThreshold, nil, nil, nil)
	for counts := range out {
		for word, count := range counts {
			c.WeightedAssign(word, count)
		}
	}

	c.Close()
	log.Println("Time elapsed: ", time.Since(start).Seconds())

	counts := c.Counts()
	var keys []string
	for key := range counts {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i] != keys[j] {
			return counts[keys[i]] > counts[keys[j]]
		}

		return keys[i] < keys[j]
	})

	sum := 0
	for _, key := range keys {
		sum += counts[key]
		log.Println(key, counts[key])
	}

	log.Println("Data Count: ", dataCount)
	log.Println("Result Count: ", sum)
}
