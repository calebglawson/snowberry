package snowberry

import (
	levenshtein "github.com/ka-weihe/fast-levenshtein"
	"regexp"
)

type fruit struct {
	original, masked string
}

type branch struct {
	start, step int
	branches    map[string]*branch
	fruit       []*fruit
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

func (e *branch) allDescendantFruit() []*fruit {
	f := make([]*fruit, len(e.fruit))
	copy(f, e.fruit)

	for _, b := range e.branches {
		f = append(f, b.allDescendantFruit()...)
	}

	return f
}

func (e *branch) addFruit(n *fruit) {
	e.fruit = append(e.fruit, n)

	var shriveledFruit []*fruit
	for _, f := range e.fruit {
		if len(f.masked) < e.start+e.step {
			shriveledFruit = append(shriveledFruit, f)
			continue
		}

		key := f.masked[e.start : e.start+e.step]
		if b, ok := e.branches[key]; ok {
			b.addFruit(f)
		} else {
			e.branches[key] = &branch{
				start:    e.start + e.step,
				step:     e.step,
				branches: make(map[string]*branch),
				fruit:    []*fruit{f},
			}
		}
	}

	e.fruit = shriveledFruit
}

type Counter struct {
	tree   *branch
	keys   []string
	counts map[string]int

	scoreThreshold           float32
	removeRegex, rejectRegex []*regexp.Regexp
	debugChannel             chan *AssignDebug
}

func NewCounter(
	step int,
	scoreThreshold float32,
	removeRegex, rejectRegex []*regexp.Regexp,
	debugChannel chan *AssignDebug,
) *Counter {
	return &Counter{
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

type AssignDebug struct {
	Input, MaskedInput         string
	Weight                     int
	BestMatch, BestMatchMasked string
	BestMatchScore             float32
	BestMatchAccepted          bool
}

func (c *Counter) WeightedAssign(input string, w int) {
	debug := &AssignDebug{Input: input}
	defer func() {
		if c.debugChannel != nil {
			c.debugChannel <- debug
		}
	}()

	maskedInput := input
	for _, r := range c.removeRegex {
		maskedInput = r.ReplaceAllString(maskedInput, "")
	}

	debug.MaskedInput = maskedInput

	for _, r := range c.rejectRegex {
		if r.MatchString(maskedInput) {
			return
		}
	}

	// Match the first part of the string until there's a mismatch
	b := c.tree.findTerminatingBranch(maskedInput)

	var bestMatch *fruit
	var bestScore float32 = 0
	for _, f := range b.allDescendantFruit() {
		if score := matchingIndex(
			maskedInput[b.start:],
			f.masked[b.start:],
			levenshtein.Distance(maskedInput[b.start:], f.masked[b.start:]),
		); score > bestScore {
			bestScore = score
			bestMatch = f

			// strings are perfectly equal and there is no point in continuing the search
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

	c.keys = append(c.keys, maskedInput)
	b.addFruit(&fruit{original: input, masked: maskedInput})
	c.counts[maskedInput] += w
}

func (c *Counter) Counts() map[string]int {
	ogCounts := make(map[string]int)
	for _, f := range c.tree.allDescendantFruit() {
		ogCounts[f.original] = c.counts[f.masked]
	}

	return ogCounts
}

// Close closes the debug channel
func (c *Counter) Close() {
	close(c.debugChannel)
}
