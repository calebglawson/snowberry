package snowberry

import (
	"regexp"

	"github.com/google/uuid"
	levenshtein "github.com/ka-weihe/fast-levenshtein"
)

type fruit struct {
	original, masked string
}

func newFruit(s string, removeRegex []*regexp.Regexp) *fruit {
	maskedInput := s
	for _, r := range removeRegex {
		maskedInput = r.ReplaceAllString(maskedInput, "")
	}

	return &fruit{original: s, masked: maskedInput}
}

func (f *fruit) key(start, step int) string {
	return f.masked[start : start+step]
}

// compare returns matching index E [0..1] from two strings and an edit distance. 1 represents a perfect match.
func (f *fruit) compare(start int, other *fruit) float32 {
	str1 := f.masked[start:]
	str2 := other.masked[start:]

	// Calculate string distance
	distance := levenshtein.Distance(str1, str2)

	// Convert strings to rune slices
	runeStr1 := []rune(str1)
	runeStr2 := []rune(str2)

	// Compare rune arrays length and make a matching percentage between them
	if len(runeStr1) >= len(runeStr2) {
		return float32(len(runeStr1)-distance) / float32(len(runeStr1))
	}
	return float32(len(runeStr2)-distance) / float32(len(runeStr2))
}

type branch struct {
	start, step int
	branches    map[string]*branch
	fruit       []*fruit
}

// findTerminatingBranch finds the deepest branch matching the provided masked string
func (e *branch) findTerminatingBranch(f *fruit) *branch {
	if e.start+e.step > len(f.masked) {
		return e
	}

	if b, ok := e.branches[f.key(e.start, e.step)]; ok {
		return b.findTerminatingBranch(f)
	}

	return e
}

// allDescendantFruit returns all fruit on the current branch and all child branches
func (e *branch) allDescendantFruit() []*fruit {
	f := make([]*fruit, len(e.fruit))
	copy(f, e.fruit)

	for _, b := range e.branches {
		f = append(f, b.allDescendantFruit()...)
	}

	return f
}

// addFruit adds fruit to the tree structure at the deepest point possible
func (e *branch) addFruit(f *fruit) {
	e.fruit = append(e.fruit, f)

	var stuntedFruit []*fruit
	for _, fr := range e.fruit {
		if len(fr.masked) < e.start+e.step {
			stuntedFruit = append(e.fruit, fr)
			return
		}

		key := fr.key(e.start, e.step)
		if b, ok := e.branches[key]; ok {
			b.addFruit(fr)
		} else {
			e.branches[key] = &branch{
				start:    e.start + e.step,
				step:     e.step,
				branches: make(map[string]*branch),
				fruit:    []*fruit{fr},
			}
		}
	}

	e.fruit = stuntedFruit
}

// Counter accepts strings and groups similar strings together, based on input parameters
type Counter struct {
	id     string
	tree   *branch
	keys   []string
	counts map[string]int

	scoreThreshold           float32
	removeRegex, rejectRegex []*regexp.Regexp
	debugChannel             chan *AssignDebug
}

// NewCounter a new Counter. `step` represents the size of substrings used when building the tree-like index.
// `scoreThreshold` is a value between 0.0 and 1.0, where 1.0 represents a perfect match. A match must have a score
// above the threshold to be matched. The match with the highest score in the candidate set is always chosen.
// `removeRegex` are optional regex that are intended to cleanse data of elements that may introduce irrelevant sources
// of uniqueness. `rejectRegex` are optional regex that ignore lines if a match is detected. They are applied after
// `removeRegex`. `debugChannel` is an optional channel that will output details about every assignment processed, use
// sparingly.
func NewCounter(
	step int,
	scoreThreshold float32,
	removeRegex, rejectRegex []*regexp.Regexp,
	debugChannel chan *AssignDebug,
) *Counter {
	return &Counter{
		id: uuid.New().String(),
		tree: &branch{
			step:     step,
			branches: make(map[string]*branch),
		},
		keys:           make([]string, 0),
		counts:         make(map[string]int),
		scoreThreshold: scoreThreshold,
		removeRegex:    removeRegex,
		rejectRegex:    rejectRegex,
		debugChannel:   debugChannel,
	}
}

// Assign submits a string for categorization with a weight of 1.
func (c *Counter) Assign(s string) {
	c.WeightedAssign(s, 1)
}

// AssignDebug contains details about every match processed
type AssignDebug struct {
	CounterID                  string
	Input, MaskedInput         string
	Weight                     int
	BestMatch, BestMatchMasked string
	BestMatchScore             float32
	BestMatchAccepted          bool
}

// WeightedAssign assigns input to a category with the given weight.
func (c *Counter) WeightedAssign(input string, w int) {
	debug := &AssignDebug{CounterID: c.id, Input: input}
	defer func() {
		if c.debugChannel != nil {
			c.debugChannel <- debug
		}
	}()

	n := newFruit(input, c.removeRegex)
	debug.MaskedInput = n.masked

	for _, r := range c.rejectRegex {
		if r.MatchString(n.masked) {
			return
		}
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
		c.counts[bestMatch.masked] += w
		debug.BestMatchAccepted = true

		return
	}

	c.keys = append(c.keys, n.masked)
	b.addFruit(n)
	c.counts[n.masked] += w
}

// Counts returns the original, unmasked map of categories and the counts
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
