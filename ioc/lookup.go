package ioc

import (
	"os"
	"strings"
	"sync"
	"time"
)

var (
	configFile string
	mapping    IOMap
)

func init() {
	configFile = os.Getenv("IOC_FILE")
}

type IOMap struct {
	Map   map[string]*IOC
	Mutex sync.RWMutex
}

func (iom *IOMap) Add(key string, duration time.Duration, read, write uint64) {
	iom.Mutex.Lock()
	defer iom.Mutex.Unlock()

	iom.Map[key] = NewIOC(duration, read, write)
}

func (iom *IOMap) Remove(key string) {
	iom.Mutex.Lock()
	defer iom.Mutex.Unlock()

	delete(iom.Map, key)
}

func (iom *IOMap) Update(key string, duration time.Duration, read, write uint64) {
	iom.Mutex.Lock()
	defer iom.Mutex.Unlock()

	c := iom.Map[key]
	c.Update(duration, read, write)
}

func (iom *IOMap) Get(key string) (*IOC, bool) {
	iom.Mutex.RLock()
	defer iom.Mutex.RUnlock()

	c, exists := iom.Map[key]
	return c, exists
}

func (iom *IOMap) FindPath(p string) *IOC {
	iom.Mutex.RLock()
	defer iom.Mutex.RUnlock()

	for k := range iom.Map {

		if strings.HasPrefix(p, k) {
			return iom.Map[k]
		}
	}
	return nil
}
