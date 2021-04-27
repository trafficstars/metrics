package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	queueLength        = 1 << 6
	queueChannelLength = 1 << 6
)

type considerValueQueueItem struct {
	metric *commonAggregative
	value  float64
}

type considerValueQueueT struct {
	writePos   uint64 // index of next write
	wroteItems uint64
	queue      [queueLength]*considerValueQueueItem
}

func (s *considerValueQueueT) release() {
	s.writePos = 0
	s.wroteItems = 0
	considerValueQueuePool.Put(s)
}

var (
	considerValueQueue      *considerValueQueueT
	considerValueQueueChan  chan *considerValueQueueT
	considerValueQueueCount uint64
	considerValueQueuePool  = sync.Pool{
		New: func() interface{} {
			newConsiderValueQueue := &considerValueQueueT{}
			for idx := range considerValueQueue.queue {
				newConsiderValueQueue.queue[idx] = &considerValueQueueItem{}
			}
			return newConsiderValueQueue
		},
	}
)

func init() {
	// initialize first considerValueQueue from pool
	swapConsiderValueQueue()

	// To do not handle locks in aggregated statistics handlers we just process this handlers
	// in a single routine. And "queueProcessor" is the function for the routine.
	considerValueQueueChan = make(chan *considerValueQueueT, queueChannelLength)
	go queueProcessor()
}

func queueProcessor() {
	for {
		// if we got a panic (inside queueProcessorLoop) then we need to restart
		queueProcessorLoop()
	}
}

func queueProcessorLoop() {
	defer recoverPanic()

	for {
		processQueue(<-considerValueQueueChan)
		atomic.AddUint64(&considerValueQueueCount, 1)
	}
}

func processQueue(queue *considerValueQueueT) {
	// we need to put queue back to the pool after work is done
	defer queue.release()
	for _, item := range queue.queue[:queue.wroteItems] {
		item.metric.doConsiderValue(item.value)
	}
}

// loadConsiderValueQueue atomically loads current queue
func loadConsiderValueQueue() *considerValueQueueT {
	return (*considerValueQueueT)(atomic.LoadPointer(
		(*unsafe.Pointer)((unsafe.Pointer)(&considerValueQueue))),
	)
}

// swapConsiderValueQueue creates a new queue from pool and atomically replaces current queue with it
func swapConsiderValueQueue() *considerValueQueueT {
	newConsiderValueQueue := considerValueQueuePool.Get().(*considerValueQueueT)
	return (*considerValueQueueT)(
		atomic.SwapPointer(
			(*unsafe.Pointer)((unsafe.Pointer)(&considerValueQueue)),
			(unsafe.Pointer)(newConsiderValueQueue)),
	)
}

func enqueueConsiderValue(metric *commonAggregative, value float64) {
retry:
	// load current queue
	queue := loadConsiderValueQueue()

	// get write position in queue
	idx := atomic.AddUint64(&queue.writePos, 1)
	switch {
	case idx == queueLength:
		// this is the last position in queue, we will put data into it
		// since there is no space left for new items, we create new queue
		swapConsiderValueQueue()
	case idx > queueLength:
		runtime.Gosched()
		goto retry
	default:
		// queue.writePointer < queueLength
		// normal operation, we write item to queue
	}

	item := queue.queue[idx-1]
	item.metric = metric
	item.value = value

	// if we just wrote last item in queue, we should schedule for processing
	wIdx := atomic.AddUint64(&queue.wroteItems, 1)
	if wIdx == queueLength {
		considerValueQueueChan <- queue
	}
}
