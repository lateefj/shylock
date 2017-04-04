// IOC Input Output Constraint
// Controls the volume of input and output base

// XXX: This is Proof of concept code not optimized

package ioc

import (
	"errors"
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
	Mutex       sync.RWMutex
	Notifier    *sync.Cond
	readLimit   *ByteLimit
	writeLimit  *ByteLimit
	resetTicker *time.Ticker
	active      bool
	exit        chan bool
}

func NewIOC(duration time.Duration, rLimit, wLimit uint64) *IOC {
	readLimit := &ByteLimit{Limit: rLimit, Notifier: &sync.Cond{L: &sync.Mutex{}}, Mutex: &sync.RWMutex{}}
	writeLimit := &ByteLimit{Limit: wLimit, Notifier: &sync.Cond{L: &sync.Mutex{}}, Mutex: &sync.RWMutex{}}

	ioc := &IOC{duration: duration, Mutex: sync.RWMutex{}, readLimit: readLimit, writeLimit: writeLimit, resetTicker: time.NewTicker(duration), active: false, exit: make(chan bool)}
	ioc.reset()
	return ioc
}
func (ioc *IOC) reset() {
	ioc.readLimit.Reset()
	ioc.writeLimit.Reset()
}

func (ioc *IOC) Start() {
	ioc.Mutex.Lock()
	ioc.active = true
	ioc.Mutex.Unlock()
	for {
		select {
		case <-ioc.resetTicker.C:
			ioc.reset()
		case <-ioc.exit:
			return
		}
	}
}

func (ioc *IOC) Stop() {
	ioc.Mutex.Lock()
	ioc.active = false
	ioc.Mutex.Unlock()
	ioc.exit <- true
	ioc.resetTicker.Stop()
}

func (ioc *IOC) Active() bool {
	ioc.Mutex.RLock()
	defer ioc.Mutex.RUnlock()
	return ioc.active
}

func (ioc *IOC) Update(duration time.Duration, read, write uint64) {
	ioc.Mutex.Lock()
	defer ioc.Mutex.Unlock()
	ioc.duration = duration
	ioc.readLimit.Limit = read
	ioc.writeLimit.Limit = write
}

// Checkout
// TODO: Would like to eventually measure back pressure if possible
func (ioc *IOC) Checkout(bl *ByteLimit, requested uint64, stream chan uint64) error {
	defer close(stream)

	for requested > 0 {
		// If the io controller is still active
		if !ioc.Active() {
			return errors.New("IOC is not active")
		}
		var out uint64
		bl.Mutex.Lock()
		// If the amount requested is greater than what is avaialble
		if bl.Bytes >= requested {
			out = requested
			bl.Bytes = bl.Bytes - requested
		} else { // Less then what is available then give it all out
			out = bl.Bytes
			bl.Bytes = 0
		}

		bl.Mutex.Unlock()
		// If there are any bytes available
		if out > 0 {
			stream <- out
		}
		// Update the amount needed requested
		requested = requested - out
		// If there are more bytes requested then wait for reset
		if requested > 0 {
			bl.Notifier.L.Lock()
			bl.Notifier.Wait()
			bl.Notifier.L.Unlock()
		}
	}
	return nil
}

func (ioc *IOC) CheckoutRead(requested uint64, stream chan uint64) error {
	return ioc.Checkout(ioc.readLimit, requested, stream)
}
func (ioc *IOC) CheckoutWrite(requested uint64, stream chan uint64) error {
	return ioc.Checkout(ioc.writeLimit, requested, stream)
}
