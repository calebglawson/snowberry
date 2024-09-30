package snowberry

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestTree(t *testing.T) {
	f := []string{
		"An aardvark ate my apple.",
		"An apple is a fruit.",
		"My favorite fruit is mango.",
		"A mango is a nutritious snack.",
	}

	expectedTree := &branch{
		start: 0,
		step:  2,
		branches: map[string]*branch{
			"An": {
				start: 2,
				step:  2,
				branches: map[string]*branch{
					" a": {
						start: 4,
						step:  2,
						branches: map[string]*branch{
							"ar": {
								start:    6,
								step:     2,
								branches: map[string]*branch{},
								fruit:    []*fruit{{original: f[0], masked: f[0]}},
							},
							"pp": {
								start:    6,
								step:     2,
								branches: map[string]*branch{},
								fruit:    []*fruit{{original: f[1], masked: f[1]}},
							},
						},
						fruit: nil,
					},
				},
				fruit: nil,
			},
			"My": {
				start:    2,
				step:     2,
				branches: map[string]*branch{},
				fruit:    []*fruit{{original: f[2], masked: f[2]}},
			},
			"A ": {
				start:    2,
				step:     2,
				branches: map[string]*branch{},
				fruit:    []*fruit{{original: f[3], masked: f[3]}},
			},
		},
		fruit: nil,
	}

	root := &branch{
		start:    0,
		step:     2,
		branches: map[string]*branch{},
	}

	for _, word := range f {
		b := root.findTerminatingBranch(newFruit(word))

		b.addFruit(&fruit{original: word, masked: word})
	}

	assert.Equal(t, expectedTree, root)
}

func TestCounter(t *testing.T) {
	f := []string{
		"An aardvark ate my apple.",
		"An apple is a fruit.",
		"A mango is a fruit.",
		"My favorite fruit is mango.",
		"A mango is a nutritious snack.",
		"There's a snake in my boot.",
		"There's a snail in my boot.",
		"There's a boot in my boot.",
		"My name is Talky Tina, and you'd better be nice to me.",
		"My name is Chalky Tina, and you'd better be nice to me!",
		"To infinity and beyond!",
		"To Nanaimo and beyond!",
		"You've got a friend in me.",

		// rejected
		"2024-09-08T23:30:03.333",
	}

	c := NewCounter(2, 0.70).
		WithIgnoreAssign([]*regexp.Regexp{regexp.MustCompile("[.!]$"), regexp.MustCompile("[,']")}).
		WithRejectAssign([]*regexp.Regexp{regexp.MustCompile("\\d{4}")})

	for _, word := range f {
		c.Assign(word)
	}

	assert.Equal(t, map[string]int{
		"An aardvark ate my apple.":                              1,
		"An apple is a fruit.":                                   1,
		"A mango is a fruit.":                                    1,
		"My favorite fruit is mango.":                            1,
		"A mango is a nutritious snack.":                         1,
		"There's a snake in my boot.":                            3,
		"My name is Talky Tina, and you'd better be nice to me.": 2,
		"To infinity and beyond!":                                1,
		"To Nanaimo and beyond!":                                 1,
		"You've got a friend in me.":                             1,
	}, c.Counts())
}
