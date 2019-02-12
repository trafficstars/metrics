package metrics

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type Spinlock int32

func (s *Spinlock) Lock() {
	for atomic.AddInt32((*int32)(s), 1) != 1 {
		atomic.AddInt32((*int32)(s), -1)
		time.Sleep(time.Nanosecond * 300 * time.Duration(rand.Intn(30)))
	}
}

func (s *Spinlock) Unlock() {
	atomic.AddInt32((*int32)(s), -1)
}
