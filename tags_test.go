package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagsSetGet(t *testing.T) {
	tags := NewTags()
	tags.Set("k", "v")
	assert.Equal(t, tags.Get("k"), "v")
}
