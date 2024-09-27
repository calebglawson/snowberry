package snowberry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTree(t *testing.T) {
	fruit := []string{
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
								fruit:    []string{fruit[0]},
							},
							"pp": {
								start:    6,
								step:     2,
								branches: map[string]*branch{},
								fruit:    []string{fruit[1]},
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
				fruit:    []string{fruit[2]},
			},
			"A ": {
				start:    2,
				step:     2,
				branches: map[string]*branch{},
				fruit:    []string{fruit[3]},
			},
		},
		fruit: nil,
	}

	root := &branch{
		start:    0,
		step:     2,
		branches: map[string]*branch{},
	}

	for _, word := range fruit {
		b := root.findTerminatingBranch(word)

		b.addFruit(word)
	}

	assert.Equal(t, expectedTree, root)
}
