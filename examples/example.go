package main

import (
	"context"
	"log"
	"os"
	"sort"
	"time"

	"github.com/calebglawson/snowberry"
	"github.com/fhalim/csvreader"
)

func main() {
	step := 10
	var scoreThreshold float32 = 0.70

	f, err := os.Open("examples/sample_data/one.csv")
	if err != nil {
		log.Fatal(err)
	}

	r, err := csvreader.NewReader(context.Background(), f)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	c := snowberry.NewCounter(step, scoreThreshold)

	for row := range r.Data {
		c.Assign(row["sentence"])
	}

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

	log.Println("Time elapsed: ", time.Since(start).Seconds())
}
