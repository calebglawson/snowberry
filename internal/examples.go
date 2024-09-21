package main

import (
	"context"
	"log"
	"os"
	"snowberry"
	"sort"
	"sync"
	"time"

	"github.com/fhalim/csvreader"
	"github.com/hbollon/go-edlib"
)

func main() {
	leafLimit := 10
	algorithm := edlib.Levenshtein
	var scoreThreshold float32 = 0.70

	in := make(chan string)
	out := make(chan map[string]int)
	var wg sync.WaitGroup
	for i := 0; i <= 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := snowberry.NewCounter(leafLimit, algorithm, scoreThreshold)

			for w := range in {
				c.Assign(w)
			}

			out <- c.Counts()
		}()
	}

	f, err := os.Open("internal/test_data/one.csv")
	if err != nil {
		log.Fatal(err)
	}

	r, err := csvreader.NewReader(context.Background(), f)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	for row := range r.Data {
		in <- row["sentence"]
	}
	close(in)

	// join thread results
	go func() {
		wg.Wait()
		close(out)
	}()

	c := snowberry.NewCounter(leafLimit, algorithm, scoreThreshold)
	for counts := range out {
		for word, count := range counts {
			c.WeightedAssign(word, count)
		}
	}

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

	for _, key := range keys {
		log.Println(key, counts[key])
	}
}
