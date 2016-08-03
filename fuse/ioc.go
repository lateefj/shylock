// IOC Input Output Constraint
// Controls the volume of input and output base

package main

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// IOC  Input / Ouput Constraint
type IOC struct {
	duration    time.Duration
	bytes       uint64
	limit       uint64
	lock        *sync.RWMutex
	notifier    *sync.Cond
	resetTicker *time.Ticker
}

func NewIOC(duration time.Duration, limit uint64) *IOC {
	ioc := &IOC{duration: duration, limit: limit, notifier: &sync.Cond{L: &sync.Mutex{}}, lock: &sync.RWMutex{}, resetTicker: time.NewTicker(duration)}
	ioc.reset()
	return ioc
}
func (ioc *IOC) reset() {
	defer ioc.notifier.Broadcast()
	defer ioc.lock.Unlock()
	ioc.lock.Lock()
	ioc.bytes = ioc.limit
}

func (ioc *IOC) Start() {
	for _ = range ioc.resetTicker.C {
		ioc.reset()
	}
}

func (ioc *IOC) Stop() {
	ioc.resetTicker.Stop()
}

func (ioc *IOC) Available() uint64 {
	defer ioc.lock.RUnlock()
	ioc.lock.RLock()
	return ioc.bytes

}

// BytesCheckout
func (ioc *IOC) BytesCheckout(requested uint64, stream chan uint64) error {
	for requested > 0 {
		log.Infof("Requested is %d current bytes %d", requested, ioc.Available())
		var out uint64
		ioc.lock.Lock()
		if ioc.bytes >= requested {
			out = requested
			ioc.bytes = ioc.bytes - requested
		} else {
			out = ioc.bytes
			ioc.bytes = 0
		}

		ioc.lock.Unlock()
		if out > 0 {
			stream <- out
		}
		requested = requested - out
		log.Infof("Requested %d after out is %d", requested, out)
		if requested > 0 {
			ioc.notifier.L.Lock()
			ioc.notifier.Wait()
			ioc.notifier.L.Unlock()
		}
	}
	close(stream)
	return nil
}
