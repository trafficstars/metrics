package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/xaionaro-go/atomicmap"
)

type sender struct {
	nextFuncID uint64
	interval   time.Duration
	funcs      atomicmap.Map
	stopChan   chan bool
}

type sendersRegistry struct {
	sync.Mutex
	m map[uint64]*sender
}

var (
	sendersRegistryInstance sendersRegistry
)

func init() {
	sendersRegistryInstance.m = map[uint64]*sender{}
}

func workerSenderLoop(data *sender) {
	ticker := time.NewTicker(data.interval)
	for {
		select {
		case <-data.stopChan:
			ticker.Stop()
			return
		case <-ticker.C:
		}
		for _, funcID := range data.funcs.Keys() {
			fn, _ := data.funcs.Get(funcID)
			if fn != nil {
				fn.(func())()
			}
		}
	}
}

func (data *sender) GetNextFuncID() uint64 {
	funcID := atomic.AddUint64(&data.nextFuncID, 1)
	var fn interface{}
	for ; fn != nil; fn, _ = data.funcs.Get(funcID) {
		funcID = atomic.AddUint64(&data.nextFuncID, 1)
	}

	return funcID
}

func appendToSenderLoop(interval time.Duration, senderFunc func()) uint64 {
	sendersRegistryInstance.Lock()
	defer sendersRegistryInstance.Unlock()

	intervalInt := uint64(interval.Nanoseconds())
	senderData := sendersRegistryInstance.m[intervalInt]
	if senderData == nil {
		senderData = &sender{
			interval: interval,
			funcs:    atomicmap.New(),
			stopChan: make(chan bool),
		}
		go workerSenderLoop(senderData)
		sendersRegistryInstance.m[intervalInt] = senderData
	}

	funcID := senderData.GetNextFuncID()

	senderData.funcs.Set(funcID, senderFunc)
	return funcID
}

func removeFromSenderLoop(interval time.Duration, funcID uint64) {
	sendersRegistryInstance.Lock()
	defer sendersRegistryInstance.Unlock()

	intervalInt := uint64(interval.Nanoseconds())
	senderData := sendersRegistryInstance.m[intervalInt]
	funcs := senderData.funcs
	funcs.(interface{ LockUnset(key atomicmap.Key) error }).LockUnset(funcID)
	if funcs.Len() == 0 {
		senderData.stopChan <- true
		delete(sendersRegistryInstance.m, intervalInt)
	}
}
