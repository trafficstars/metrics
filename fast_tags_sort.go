package metrics

import (
	"math/rand"
	"sort"
)

// this code is mostly copied from https://github.com/demdxx/sort-algorithms/blob/master/algorithms.go

func (tags *FastTags) shuffle() {
	// This function is used only for as a benchmark helper, so there's no need to shuffle correctly.
	newIdx := rand.Perm(tags.Len())
	for oldIdx, newIdx := range newIdx {
		tags.Swap(oldIdx, newIdx)
	}
}

func (tags *FastTags) sortBuiltin() {
	sort.Sort(tags)
}

func (tags *FastTags) sortBubble() {
	n := tags.Len() - 1
	b := false
	for i := 0; i < n; i++ {
		for j := 0; j < n-i; j++ {
			if tags.Less(j+1, j) {
				tags.Swap(j+1, j)
				b = true
			}
		}
		if !b {
			break
		}
		b = false
	}
}

func (tags FastTags) sortQuick_partition(p int, r int) int {
	x := tags.Slice[r]
	i := p - 1
	for j := p; j < r; j++ {
		if tags.Slice[j].Key <= x.Key {
			i++
			tags.Swap(i, j)
		}
	}
	i++
	tags.Swap(i, r)
	return i
}

func (tags FastTags) sortQuick_r(p int, r int) {
	var q int
	if p < r {
		q = tags.sortQuick_partition(p, r)
		tags.sortQuick_r(p, q-1)
		tags.sortQuick_r(q+1, r)
	}
}

func (tags FastTags) sortQuick() {
	tags.sortQuick_r(0, len(tags.Slice)-1)
}
