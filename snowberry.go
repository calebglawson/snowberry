package snowberry

import (
	"regexp"

	levenshtein "github.com/ka-weihe/fast-levenshtein"
)

type fruit struct {
	original, masked string
}

func newFruit(s string) *fruit {
	return &fruit{original: s, masked: s}
}

func (f *fruit) withIgnorePatterns(patterns []*regexp.Regexp) *fruit {
	for _, p := range patterns {
		f.masked = p.ReplaceAllString(f.masked, "")
	}

	return f
}

func (f *fruit) shouldReject(patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(f.masked) {
			return true
		}
	}

	return false
}

func (f *fruit) key(start, end int) string {
	return f.masked[start:end]
}

// compare returns matching index E [0..1] from two strings and an edit distance. 1 represents a perfect match.
func (f *fruit) compare(start int, other *fruit) float32 {
	t := f.masked[start:]
	o := other.masked[start:]

	// Calculate string distance
	distance := levenshtein.Distance(t, o)

	// Compare rune arrays length and make a matching percentage between them
	if len(t) >= len(o) {
		return float32(len(t)-distance) / float32(len(t))
	}
	return float32(len(o)-distance) / float32(len(o))
}

type branch struct {
	start, step int
	branches    map[string]*branch
	fruit       []*fruit
}

func (b *branch) end() int {
	return b.start + b.step
}

// findTerminatingBranch finds the deepest branch matching the provided masked string
func (b *branch) findTerminatingBranch(f *fruit) *branch {
	if len(f.masked) < b.end() {
		return b
	}

	if child, ok := b.branches[f.key(b.start, b.end())]; ok {
		return child.findTerminatingBranch(f)
	}

	return b
}

// allDescendantFruit returns all fruit on the current branch and all child branches
func (b *branch) allDescendantFruit() []*fruit {
	f := make([]*fruit, len(b.fruit))
	copy(f, b.fruit)

	for _, child := range b.branches {
		f = append(f, child.allDescendantFruit()...)
	}

	return f
}

// addFruit adds fruit to the tree structure at the deepest point possible
func (b *branch) addFruit(f *fruit) {
	b.fruit = append(b.fruit, f)

	var stuntedFruit []*fruit
	for _, fr := range b.fruit {
		if len(fr.masked) < b.end() {
			stuntedFruit = append(stuntedFruit, fr)
			continue
		}

		key := fr.key(b.start, b.end())
		if child, ok := b.branches[key]; ok {
			child.addFruit(fr)
		} else {
			b.branches[key] = &branch{
				start:    b.end(),
				step:     b.step,
				branches: make(map[string]*branch),
				fruit:    []*fruit{fr},
			}
		}
	}

	b.fruit = stuntedFruit
}

// Counter accepts strings and groups similar strings together, based on input parameters
type Counter struct {
	tree   *branch
	counts map[string]int

	scoreThreshold                 float32
	ignorePatterns, rejectPatterns []*regexp.Regexp
	debugChannel                   chan *AssignDebug
}

// NewCounter a new Counter. `step` represents the size of substrings used when building the tree-like index.
// `scoreThreshold` is a value between 0.0 and 1.0, where 1.0 represents a perfect match. A match must have a score
// above the threshold to be matched. The match with the highest score in the candidate set is always chosen.
func NewCounter(step int, scoreThreshold float32) *Counter {
	return &Counter{
		tree: &branch{
			step:     step,
			branches: make(map[string]*branch),
		},
		counts:         make(map[string]int),
		scoreThreshold: scoreThreshold,
	}
}

// WithIgnoreAssign returns a Counter which will ignore the targeted contents of assignments matching all regex
func (c *Counter) WithIgnoreAssign(r []*regexp.Regexp) *Counter {
	c.ignorePatterns = r

	return c
}

// WithRejectAssign returns a Counter which will reject assignments matching one or more regex
func (c *Counter) WithRejectAssign(r []*regexp.Regexp) *Counter {
	c.rejectPatterns = r

	return c
}

// WithDebugChannel returns a Counter which will pass AssignDebug to the passed in channel for debug/tuning purposes
func (c *Counter) WithDebugChannel(debugChannel chan *AssignDebug) *Counter {
	c.debugChannel = debugChannel

	return c
}

// AssignDebug contains details about every match processed
type AssignDebug struct {
	Input, MaskedInput         string
	Rejected                   bool
	BestMatch, BestMatchMasked string
	BestMatchScore             float32
	BestMatchAccepted          bool
}

// Assign assigns input to a category.
func (c *Counter) Assign(input string) {
	debug := &AssignDebug{Input: input}
	defer func() {
		if c.debugChannel != nil {
			c.debugChannel <- debug
		}
	}()

	n := newFruit(input).withIgnorePatterns(c.ignorePatterns)
	debug.MaskedInput = n.masked

	if n.shouldReject(c.rejectPatterns) {
		debug.Rejected = true

		return
	}

	// Match the first part of the masked string until there's a mismatch
	b := c.tree.findTerminatingBranch(n)

	var bestMatch *fruit
	var bestScore float32 = 0
	for _, f := range b.allDescendantFruit() {
		if score := n.compare(b.start, f); score > bestScore {
			bestScore = score
			bestMatch = f

			// Strings are perfectly equal and there is no point in continuing the search.
			if score == 1 {
				break
			}
		}
	}

	if bestMatch != nil {
		debug.BestMatch = bestMatch.original
		debug.BestMatchMasked = bestMatch.masked
		debug.BestMatchScore = bestScore
	}

	if bestScore > c.scoreThreshold && bestMatch != nil {
		c.counts[bestMatch.masked]++
		debug.BestMatchAccepted = true

		return
	}

	b.addFruit(n)
	c.counts[n.masked]++
}

// Counts returns the original, unmasked map of categories and counts
func (c *Counter) Counts() map[string]int {
	ogCounts := make(map[string]int)
	for _, f := range c.tree.allDescendantFruit() {
		ogCounts[f.original] = c.counts[f.masked]
	}

	return ogCounts
}

// Close closes the debug channel, no-op if not set
func (c *Counter) Close() {
	if c.debugChannel != nil {
		close(c.debugChannel)
	}
}
