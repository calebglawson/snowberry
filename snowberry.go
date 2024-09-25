package snowberry

import (
	"github.com/hbollon/go-edlib"
)

type branch struct {
	leafLimit int
	chunkSize int
	position  int
	branches  map[string]*branch
	leaves    []string
}

func newTree(leafLimit, chunkSize int) *branch {
	return &branch{leafLimit: leafLimit, chunkSize: chunkSize, position: 0}
}

// descend recursively navigates to the deepest matching branch
func (b *branch) descend(word string) *branch {
	if len(word)-1 == b.position {
		return b
	}

	start := b.position
	if start < 0 {
		start = 0
	}

	end := start + b.chunkSize
	if end > len(word) {
		end = len(word)
	}

	if c, ok := b.branches[word[start:end]]; ok {
		return c.descend(word)
	}

	return b
}

// addLeaf adds a leaf to the current branch, if there are too many leaves, the branch grows
func (b *branch) addLeaf(leaf string) {
	b.leaves = append(b.leaves, leaf)

	if len(b.leaves) > b.leafLimit {
		b.grow()
	}
}

// grow reviews all leaves and adds them to child branches, until the end of the string or addLeaf is satisfied
func (b *branch) grow() {
	if len(b.branches) == 0 {
		b.branches = make(map[string]*branch)
	}

	var stuntedLeaves []string
	for _, l := range b.leaves {
		if len(l) > b.position+b.chunkSize {
			start := b.position
			if b.position < 0 {
				start = 0
			}

			end := start + b.chunkSize
			if end > len(l) {
				end = len(l)
			}

			name := l[start:end]
			if c, ok := b.branches[name]; ok {
				c.addLeaf(l)
			} else {
				b.branches[name] = &branch{
					leafLimit: b.leafLimit,
					chunkSize: b.chunkSize,
					position:  b.position + b.chunkSize,
					leaves:    []string{l},
				}
			}

			continue
		}

		stuntedLeaves = append(stuntedLeaves, l)
	}

	b.leaves = stuntedLeaves
}

// allDescendantLeaves returns all the leaves of the current branch and its children
func (b *branch) allDescendantLeaves() []string {
	var leaves []string
	leaves = append(leaves, b.leaves...)

	for _, c := range b.branches {
		leaves = append(leaves, c.allDescendantLeaves()...)
	}

	return leaves
}

type Counter struct {
	tree   *branch
	keys   []string
	counts map[string]int

	algorithm      edlib.Algorithm
	scoreThreshold float32
}

func NewCounter(leafLimit, chunkSize int, algorithm edlib.Algorithm, scoreThreshold float32) *Counter {
	return &Counter{
		tree:           newTree(leafLimit, chunkSize),
		keys:           make([]string, 0),
		counts:         make(map[string]int),
		algorithm:      algorithm,
		scoreThreshold: scoreThreshold,
	}
}

func (c *Counter) Assign(s string) {
	c.WeightedAssign(s, 1)
}

func (c *Counter) WeightedAssign(s string, w int) {
	// Match the first part of the string until there's a mismatch
	b := c.tree.descend(s)

	bestStr := ""
	var bestScore float32 = 0
	for _, l := range b.allDescendantLeaves() {
		// compare the ends of the strings, since everything up until the position is an exact-match
		score, err := edlib.StringsSimilarity(s[b.position:], l[b.position:], c.algorithm)
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
	b.addLeaf(s)
	c.counts[s] += w
}

func (c *Counter) Counts() map[string]int {
	return c.counts
}
