// IOC Input Output Constraint
// Controls the volume of input and output base

package main

import (
	"sync"
	"time"
)

type ByteLimit struct {
	Bytes    uint64
	Limit    uint64
	Mutex    *sync.RWMutex
	Notifier *sync.Cond
}

func (bl *ByteLimit) Reset() {
	defer bl.Notifier.Broadcast()
	defer bl.Mutex.Unlock()
	bl.Mutex.Lock()
	bl.Bytes = bl.Limit
}

func (bl *ByteLimit) Available() uint64 {
	defer bl.Mutex.RUnlock()
	bl.Mutex.RLock()
	return bl.Bytes
}

// IOC  Input / Ouput Constraint
type IOC struct {
	duration    time.Duration
	limit       uint64
	Mutex       *sync.RWMutex
	Notifier    *sync.Cond
	readLimit   *ByteLimit
	writeLimit  *ByteLimit
	resetTicker *time.Ticker
}

func NewIOC(duration time.Duration, rLimit, wLimit uint64) *IOC {
	readLimit := &ByteLimit{Limit: rLimit, Notifier: &sync.Cond{L: &sync.Mutex{}}, Mutex: &sync.RWMutex{}}
	writeLimit := &ByteLimit{Limit: wLimit, Notifier: &sync.Cond{L: &sync.Mutex{}}, Mutex: &sync.RWMutex{}}

	ioc := &IOC{duration: duration, readLimit: readLimit, writeLimit: writeLimit, resetTicker: time.NewTicker(duration)}
	ioc.reset()
	return ioc
}
func (ioc *IOC) reset() {
	ioc.readLimit.Reset()
	ioc.writeLimit.Reset()
}

func (ioc *IOC) Start() {
	for {
		select {
		case <-ioc.resetTicker.C:
			ioc.reset()
		}
	}
}

func (ioc *IOC) Stop() {
	ioc.resetTicker.Stop()
}

// Checkout
func (ioc *IOC) Checkout(bl *ByteLimit, requested uint64, stream chan uint64) error {
	for requested > 0 {
		//start := time.Now()
		var out uint64
		bl.Mutex.Lock()
		if bl.Bytes >= requested {
			out = requested
			bl.Bytes = bl.Bytes - requested
		} else {
			out = bl.Bytes
			bl.Bytes = 0
		}

		bl.Mutex.Unlock()
		//dl := time.Now().Sub(start)
		if out > 0 {
			stream <- out
		}
		requested = requested - out
		if requested > 0 {
			//s := time.Now()
			bl.Notifier.L.Lock()
			bl.Notifier.Wait()
			bl.Notifier.L.Unlock()
			//d := time.Now().Sub(s)
		}
		//end := time.Now()
		//diff := end.Sub(start)

	}
	close(stream)
	return nil
}

func (ioc *IOC) CheckoutRead(requested uint64, stream chan uint64) error {
	return ioc.Checkout(ioc.readLimit, requested, stream)
}
func (ioc *IOC) CheckoutWrite(requested uint64, stream chan uint64) error {
	return ioc.Checkout(ioc.writeLimit, requested, stream)
}
