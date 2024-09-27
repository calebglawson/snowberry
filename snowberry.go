package snowberry

import levenshtein "github.com/ka-weihe/fast-levenshtein"

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
		return b.findTerminatingBranch(f)
	}

	return e
}

func (e *branch) allDescendantFruit() []string {
	fruit := make([]string, len(e.fruit))
	copy(fruit, e.fruit)

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

		key := fruit[e.start : e.start+e.step]
		if b, ok := e.branches[key]; ok {
			b.addFruit(fruit)
		} else {
			e.branches[key] = &branch{
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

// Return matching index E [0..1] from two strings and an edit distance
func matchingIndex(str1 string, str2 string, distance int) float32 {
	// Convert strings to rune slices
	runeStr1 := []rune(str1)
	runeStr2 := []rune(str2)
	// Compare rune arrays length and make a matching percentage between them
	if len(runeStr1) >= len(runeStr2) {
		return float32(len(runeStr1)-distance) / float32(len(runeStr1))
	}
	return float32(len(runeStr2)-distance) / float32(len(runeStr2))
}

func (c *Counter) WeightedAssign(s string, w int) {
	// Match the first part of the string until there's a mismatch
	b := c.tree.findTerminatingBranch(s)

	bestStr := ""
	var bestScore float32 = 0
	for _, l := range b.allDescendantFruit() {
		if score := matchingIndex(
			s[b.start:],
			l[b.start:],
			levenshtein.Distance(s[b.start:], l[b.start:]),
		); score > bestScore {
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
