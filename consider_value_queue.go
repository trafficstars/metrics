package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	queueLength = 64
)

type considerValueQueueItem struct {
	metric *commonAggregative
	value  float64
}

type considerValueQueueT struct {
	writePointer uint32
	wrotePointer uint32
	queue        [queueLength]*considerValueQueueItem
}

var (
	considerValueQueue     = &considerValueQueueT{}
	considerValueQueuePool = sync.Pool{
		New: func() interface{} {
			newConsiderValueQueue := &considerValueQueueT{}
			for idx := range considerValueQueue.queue {
				newConsiderValueQueue.queue[idx] = &considerValueQueueItem{}
			}
			return newConsiderValueQueue
		},
	}
	considerValueQueueChan chan *considerValueQueueT
)

func init() {
	swapConsiderValueQueue()

	// To do not handle locks in aggregative statistics handlers we just process this handlers
	// in a single routine. And "queueProcessor" is the function for the routine.
	go queueProcessor()
}

func queueProcessor() {
	considerValueQueueChan = make(chan *considerValueQueueT, 16)
	for {
		// if we got a panic (inside queueProcessorLoop) then we need to restart
		queueProcessorLoop()
	}
}

func queueProcessorLoop() {
	defer recoverPanic()

	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		var queue *considerValueQueueT
		select {
		case <-ticker.C:
			queue = swapConsiderValueQueue()
			atomic.AddUint32(&considerValueQueueNumber, 1)
		case newQueue := <-considerValueQueueChan:
			queue = newQueue
		}

		processQueue(queue)
	}
}

func processQueue(queue *considerValueQueueT) {
	count := 0
	var wrotes uint32
	for {
		writes := atomic.LoadUint32(&queue.writePointer)
		if writes > queueLength {
			writes = queueLength
		}
		wrotes = atomic.LoadUint32(&queue.wrotePointer)
		if writes == wrotes {
			break
		}
		runtime.Gosched()
		count++
		if count > 10 {
			time.Sleep(time.Nanosecond * time.Duration(count))
		}
		if count > 10000 {
			panic(fmt.Errorf(`an infinite loop :( : %v %v`, writes, wrotes))
		}
	}

	for _, item := range queue.queue[:wrotes] {
		item.metric.doConsiderValue(item.value)
	}

	atomic.AddUint32(&considerValueQueuesProcessed, 1)

	queue.writePointer = 0
	queue.wrotePointer = 0
	considerValueQueuePool.Put(queue)
}

var (
	considerValueQueueNumber     uint32
	considerValueQueuesProcessed uint32
)

func submitConsiderValueQueue(queue *considerValueQueueT) {
	atomic.AddUint32(&considerValueQueueNumber, 1)
	considerValueQueueChan <- queue
}

func waitUntilAllSubmittedConsiderValueQueuesProcessed() {
	for {
		if atomic.LoadUint32(&considerValueQueueNumber) == atomic.LoadUint32(&considerValueQueuesProcessed) {
			break
		}
		//fmt.Println(atomic.LoadUint32(&considerValueQueueNumber), atomic.LoadUint32(&considerValueQueuesProcessed))
		runtime.Gosched()
	}
}

func loadConsiderValueQueue() *considerValueQueueT {
	return (*considerValueQueueT)(atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&considerValueQueue))))
}

func swapConsiderValueQueue() *considerValueQueueT {
	newConsiderValueQueue := considerValueQueuePool.Get().(*considerValueQueueT)
	return (*considerValueQueueT)(atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&considerValueQueue)), (unsafe.Pointer)(newConsiderValueQueue)))
}

func enqueueConsiderValue(metric *commonAggregative, value float64) {
	queue := loadConsiderValueQueue()
	idx := atomic.AddUint32(&queue.writePointer, 1) - 1
	if idx == queueLength-1 {
		anotherQueue := swapConsiderValueQueue()
		if anotherQueue != queue {
			submitConsiderValueQueue(anotherQueue)
		}
	} else if idx == queueLength {
		runtime.Gosched()
		enqueueConsiderValue(metric, value)
		return
	} else if idx > queueLength {
		time.Sleep(time.Microsecond)
		enqueueConsiderValue(metric, value)
		return
	}
	item := queue.queue[idx]
	item.metric = metric
	item.value = value
	wrote := atomic.AddUint32(&queue.wrotePointer, 1)
	if wrote == queueLength {
		submitConsiderValueQueue(queue)
	}
}
