package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFastTagDuplicatesBug(t *testing.T) {
	Reset()

	SetDefaultTags(Tags{
		"tag0": true,
		"tag1": true,
		"tag2": true,
	})

	SetHiddenTags(HiddenTags{
		{Key: "thirdTag"},
	})

	tags := NewFastTags().
		Set("tag2", false).
		Set("tag3", false)
	Count("someKey", tags).Increment()

	list := List()
	assert.Equal(t, `someKey,tag3=false`, string((*list)[0].GetKey()))
}