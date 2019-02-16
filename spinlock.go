package metrics

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type Spinlock int32

func (s *Spinlock) Lock() {
	for !atomic.CompareAndSwapInt32((*int32)(s), 0, 1) {
		time.Sleep(time.Nanosecond * 300 * time.Duration(rand.Intn(30)))
	}
}

func (s *Spinlock) Unlock() {
	for !atomic.CompareAndSwapInt32((*int32)(s), 1, 0) {
		time.Sleep(time.Nanosecond * 300 * time.Duration(rand.Intn(30)))
	}
}
