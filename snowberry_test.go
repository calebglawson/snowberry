package snowberry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTree(t *testing.T) {
	words := []string{
		"An aardvark ate my apple.",
		"An apple is a fruit.",
		"My favorite fruit is mango.",
		"A mango is a nutritious snack.",
	}

	leafLimit := 2

	expectedTree := &branch{
		leafLimit: &leafLimit,
		position:  -1,
		branches: map[rune]*branch{
			'A': {
				leafLimit: &leafLimit,
				position:  0,
				branches: map[rune]*branch{
					'n': {
						leafLimit: &leafLimit,
						position:  1,
						leaves: []string{
							"An aardvark ate my apple.",
							"An apple is a fruit.",
						},
					},
					' ': {
						leafLimit: &leafLimit,
						position:  1,
						leaves: []string{
							"A mango is a nutritious snack.",
						},
					},
				},
			},
			'M': {
				leafLimit: &leafLimit,
				position:  0,
				branches:  nil,
				leaves:    []string{"My favorite fruit is mango."},
			},
		},
		leaves: nil,
	}

	tree := newTree(leafLimit)

	for _, word := range words {
		b := tree.descend(word)

		b.addLeaf(word)
	}

	assert.Equal(t, expectedTree, tree)
}
