package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	queueLength 		= 1 << 6
	queueChannelLength 	= 1 << 6
)

type considerValueQueueItem struct {
	metric *commonAggregative
	value  float64
}

type considerValueQueueT struct {
	writePos     uint32 // index of last added item
	queue        [queueLength]*considerValueQueueItem
}

func (s *considerValueQueueT) release() {
	s.writePos = 0
	considerValueQueuePool.Put(s)
}

var (
	considerValueQueue           *considerValueQueueT
	considerValueQueueChan       chan *considerValueQueueT
	considerValueQueuePool       = sync.Pool{
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
	go queueProcessor()
}

func queueProcessor() {
	considerValueQueueChan = make(chan *considerValueQueueT, queueChannelLength)
	for {
		// if we got a panic (inside queueProcessorLoop) then we need to restart
		queueProcessorLoop()
	}
}

func queueProcessorLoop() {
	defer recoverPanic()

	for {
		select {
		case queue := <-considerValueQueueChan:
			processQueue(queue)
		}
	}
}

func processQueue(queue *considerValueQueueT) {
	// we need to put queue back to the pool after work is done
	defer queue.release()
	for _, item := range queue.queue {
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
	// load current queue
	queue := loadConsiderValueQueue()

	// get write position in queue
	idx := atomic.AddUint32(&queue.writePos, 1)
	switch {
	case idx == queueLength:
		// this is the last position in queue, we will put data into it
		// since there is no space left for new items, we create new queue
		swapConsiderValueQueue()
	case idx > queueLength:
		runtime.Gosched()
		enqueueConsiderValue(metric, value)
		return
	default:
		// queue.writePointer < queueLength
		// normal operation, we write item to queue
	}

	item := queue.queue[idx-1]
	item.metric = metric
	item.value = value

	// if we just wrote last item in queue, we should schedule for processing
	if idx == queueLength {
		considerValueQueueChan <- queue
	}
}
