package snowberry

import (
	"github.com/hbollon/go-edlib"
)

type branch struct {
	start, step int
	branches    map[string]*branch
	fruit       []string
}

func (e *branch) findTerminatingBranch(f string) *branch {
	if e.start+e.step > len(f) {
		return e
	}

	if b, ok := e.branches[f[e.start:e.start+e.step]]; ok {
		b.findTerminatingBranch(f)
	}

	return e
}

func (e *branch) allDescendantFruit() []string {
	var fruit []string
	fruit = append(fruit, e.fruit...)

	for _, b := range e.branches {
		fruit = append(fruit, b.allDescendantFruit()...)
	}

	return fruit
}

func (e *branch) addFruit(f string) {
	e.fruit = append(e.fruit, f)

	var shriveledFruit []string
	for _, fruit := range e.fruit {
		if len(fruit) < e.start+e.step {
			shriveledFruit = append(shriveledFruit, fruit)
			continue
		}

		if b, ok := e.branches[fruit[e.start:e.start+e.step]]; ok {
			b.addFruit(fruit)
		} else {
			e.branches[fruit[e.start:e.start+e.step]] = &branch{
				start:    e.start + e.step,
				step:     e.step,
				branches: make(map[string]*branch),
				fruit:    []string{fruit},
			}
		}
	}

	e.fruit = shriveledFruit
}

type Counter struct {
	tree   *branch
	keys   []string
	counts map[string]int

	scoreThreshold float32
}

func NewCounter(step int, scoreThreshold float32) *Counter {
	return &Counter{
		tree: &branch{
			step:     step,
			branches: make(map[string]*branch),
		},
		keys:           make([]string, 0),
		counts:         make(map[string]int),
		scoreThreshold: scoreThreshold,
	}
}

func (c *Counter) Assign(s string) {
	c.WeightedAssign(s, 1)
}

func (c *Counter) WeightedAssign(s string, w int) {
	// Match the first part of the string until there's a mismatch
	b := c.tree.findTerminatingBranch(s)

	bestStr := ""
	var bestScore float32 = 0
	for _, l := range b.allDescendantFruit() {
		// strings have been exact matched up until this point, so compare only the remainder
		score, err := edlib.StringsSimilarity(s[b.start:], l[b.start:], edlib.Levenshtein)
		if err != nil {
			panic(err)
		}

		if score > bestScore {
			bestScore = score
			bestStr = l

			// strings are perfectly equal and there is no point in continuing the search
			if score == 1 {
				break
			}
		}
	}

	if bestScore > c.scoreThreshold {
		c.counts[bestStr] += w

		return
	}

	c.keys = append(c.keys, s)
	b.addFruit(s)
	c.counts[s] += w
}

func (c *Counter) Counts() map[string]int {
	return c.counts
}
