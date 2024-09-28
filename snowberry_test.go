package snowberry

import (
	"github.com/stretchr/testify/assert"
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
		b := root.findTerminatingBranch(word)

		b.addFruit(&fruit{original: word, masked: word})
	}

	assert.Equal(t, expectedTree, root)
}
