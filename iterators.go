package metrics

// Metric iterators allows us not to create a seperate goroutine for every metric. It collects all metrics and Iterate()-s them in the specified interval

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/xaionaro-go/atomicmap"
)

type iterator interface {
	EqualsTo(iterator) bool
	IsRunning() bool
	Iterate()
	GetInterval() time.Duration
}

type iterationHandler struct {
	sync.RWMutex

	iterateInterval time.Duration
	iterators       []iterator
	stopChan        chan struct{}
}

type iterationHandlersT struct {
	sync.Mutex

	m             atomicmap.Map
	routinesCount int64
	//iterators []*metricIterator
}

var (
	iterationHandlers = iterationHandlersT{
		m: atomicmap.New(),
	}
)

func (iterationHandler *iterationHandler) loop() {
	ticker := time.NewTicker(iterationHandler.iterateInterval)
	for {
		select {
		case <-iterationHandler.stopChan:
			ticker.Stop()
			atomic.AddInt64(&iterationHandlers.routinesCount, -1)
			return
		case <-ticker.C:
		}
		iterationHandler.RLock()
		iterators := iterationHandler.iterators
		iterationHandler.RUnlock()

		for _, iterator := range iterators {
			if !iterator.IsRunning() {
				continue
			}
			iterator.Iterate()
		}
	}
}

func (iterationHandler *iterationHandler) start() {
	go func() {
		atomic.AddInt64(&iterationHandlers.routinesCount, 1)
		iterationHandler.loop()
	}()
}

func (iterationHandler *iterationHandler) stop() {
	go func() {
		iterationHandler.stopChan <- struct{}{}
	}()
}

// Add add a metric to the iterationHandler. It will periodically call method Iterate() of the metric
func (iterationHandler *iterationHandler) Add(iterator iterator) {
	iterationHandler.RLock()
	iterators := iterationHandler.iterators
	iterationHandler.RUnlock()
	for _, curIterator := range iterators {
		if curIterator.EqualsTo(iterator) {
			return
		}
	}

	// RLock is prefered over Lock and a real adding is a rare event, so…

	iterationHandler.Lock()
	defer iterationHandler.Unlock()
	for _, curIterator := range iterationHandler.iterators {
		if curIterator.EqualsTo(iterator) {
			return
		}
	}
	iterationHandler.iterators = append(iterationHandler.iterators, iterator)
}

// Remove removed a metric from the iterationHandler.
// It it was the last metric it will stop the iterationHandler and remove it from the iterationHandlers registry
func (iterationHandler *iterationHandler) Remove(removeIterator iterator) (result bool) {
	iterationHandler.Lock()
	defer iterationHandler.Unlock()

	if len(iterationHandler.iterators) == 1 {
		if iterationHandler.iterators[0] == removeIterator {
			iterationHandler.iterators = nil
			//iterationHandler.stop()
			//mapKey := uint64(iterationHandler.iterateInterval.Nanoseconds())
			//iterationHandlers.m.(interface{ Unset(atomicmap.Key) error }).Unset(mapKey)
			//iterationHandler.Release()
			return true
		}
		return false
	}

	// len(iterationHandler.iterators) > 1

	leftIterators := make([]iterator, 0, len(iterationHandler.iterators)-1)
	for _, curIterator := range iterationHandler.iterators {
		if curIterator == removeIterator {
			continue
		}
		leftIterators = append(leftIterators, curIterator)
	}

	result = false
	if len(iterationHandler.iterators) != len(leftIterators) {
		result = true
		iterationHandler.iterators = leftIterators
	}

	return
}

func (iterationHandlers *iterationHandlersT) getIterationHandler(iterator iterator) *iterationHandler {
	if iterator == nil {
		return nil
	}

	// Lock() requires ~50ns of time out my CPU, while atomicmap.Map.GetByUint64() requires only ~30ns
	iterationHandlerKey := iterator.GetInterval().Nanoseconds()
	iterationHandlerI, _ := iterationHandlers.m.GetByUint64(uint64(iterationHandlerKey))
	if iterationHandlerI == nil {
		return nil
	}
	return iterationHandlerI.(*iterationHandler)
}

func (iterationHandlers *iterationHandlersT) getOrCreateIterationHandler(iterator iterator) *iterationHandler {
	iterationHandler := iterationHandlers.getIterationHandler(iterator)
	if iterationHandler != nil {
		return iterationHandler
	}

	iterationHandlers.Lock()
	iterationHandler = iterationHandlers.getIterationHandler(iterator)
	if iterationHandler != nil {
		iterationHandlers.Unlock()
		return iterationHandler
	}

	iterationHandler = newIterationHandler()
	iterationHandler.iterateInterval = iterator.GetInterval()
	if iterationHandler.iterateInterval == time.Duration(0) {
		iterationHandlers.Unlock()
		return nil
	}
	iterationHandler.start()
	iterationHandlers.m.Set(uint64(iterationHandler.iterateInterval.Nanoseconds()), iterationHandler)
	iterationHandlers.Unlock()
	return iterationHandler
}

var c int32

// Add adds a metric to the iterators registry. So it will be called method Iterate() of the metric in the interval returned by metric.GetInterval()
func (iterators *iterationHandlersT) Add(iterator iterator) {
	iterationHandler := iterators.getOrCreateIterationHandler(iterator)
	iterationHandler.Add(iterator)
}

// Remove removes a metric from the iterators registry (see `(*metricIterators).Add()`).
func (iterators *iterationHandlersT) Remove(iterator iterator) {
	iterationHandler := iterators.getIterationHandler(iterator)
	iterationHandler.Remove(iterator)
}
