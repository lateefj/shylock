package ioc

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

func NewIOMap() *IOMap {
	return &IOMap{Map: make(map[string]*IOC), Mutex: sync.RWMutex{}}
}

func LoadIOCConfig(f io.Reader) *IOMap {
	reader := csv.NewReader(f)
	mapping := NewIOMap()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error parsing io.Reader with error: %s", err)
		}

		path := strings.TrimSpace(record[0])
		durationConf := record[1]
		readConf := record[2]
		writeConf := record[3]

		duration, err := strconv.Atoi(durationConf)
		if err != nil {
			log.Fatalf("Error parsing duration %s", durationConf)
		}
		read, err := strconv.ParseUint(readConf, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing read limit %s", readConf)
		}
		write, err := strconv.ParseUint(writeConf, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing write limit %s", writeConf)
		}

		mapping.Add(path, time.Duration(duration)*time.Millisecond, read, write)
	}
	return mapping

}

type IOMap struct {
	Map   map[string]*IOC
	Mutex sync.RWMutex
}

func (iom *IOMap) Add(key string, duration time.Duration, read, write uint64) {
	c := NewIOC(duration, read, write)
	// Clock only around the map modification
	iom.Mutex.Lock()
	iom.Map[key] = c
	iom.Mutex.Unlock()
	go c.Start()
}

func (iom *IOMap) Remove(key string) {
	// Locking only around map modification
	iom.Mutex.Lock()
	c := iom.Map[key]
	delete(iom.Map, key)
	iom.Mutex.Unlock()
	// Stop the ioc
	c.Stop()
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
