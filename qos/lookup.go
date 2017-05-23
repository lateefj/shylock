package qos

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LoadIOCConfig ... Takes an io.Reader expecting csv file and returns a *IOMap
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
		if len(record) < 3 { // Skip empty or incomplete lines
			continue
		}
		fmt.Printf("Record is %v\n", record)

		path := strings.TrimSpace(record[0])
		readConf := record[1]
		writeConf := record[2]

		read, err := strconv.ParseUint(readConf, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing read limit %s", readConf)
		}
		write, err := strconv.ParseUint(writeConf, 10, 64)
		if err != nil {
			log.Fatalf("Error parsing write limit %s", writeConf)
		}

		mapping.Add(path, 1*time.Second, read, write)
		fmt.Printf("Path %s read %d write %d\n", path, read, write)
	}
	return mapping

}

// IOMap ... Mapping of key to IOC
type IOMap struct {
	Map   map[string]*IOC
	Mutex sync.RWMutex
}

// NewIOMap ... Creates a new IOMap with default params
func NewIOMap() *IOMap {
	return &IOMap{Map: make(map[string]*IOC), Mutex: sync.RWMutex{}}
}

// Add ... Add a IOC with a specific key to the map
func (iom *IOMap) Add(key string, duration time.Duration, read, write uint64) {
	c := NewIOC(duration, read, write)
	// Clock only around the map modification
	iom.Mutex.Lock()
	iom.Map[key] = c
	iom.Mutex.Unlock()
	go c.Start()
}

// Remove ... Remove a key
func (iom *IOMap) Remove(key string) {
	// Locking only around map modification
	iom.Mutex.Lock()
	c := iom.Map[key]
	delete(iom.Map, key)
	iom.Mutex.Unlock()
	// Stop the ioc
	c.Stop()
}

// Update ... Update existing entry
func (iom *IOMap) Update(key string, duration time.Duration, read, write uint64) {
	iom.Mutex.Lock()
	defer iom.Mutex.Unlock()

	c := iom.Map[key]
	c.Update(duration, read, write)
}

// Get ... Retrieve based on a key
func (iom *IOMap) Get(key string) (*IOC, bool) {
	iom.Mutex.RLock()
	defer iom.Mutex.RUnlock()

	c, exists := iom.Map[key]
	return c, exists
}

// FindPath ... Search the keys space for first path to match
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
