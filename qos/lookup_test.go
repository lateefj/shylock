package qos

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIOMapCrud(t *testing.T) {
	iom := NewIOMap()

	key := "foo"
	duration := 1 * time.Millisecond
	read := uint64(1)
	write := uint64(1)
	iom.Add(key, duration, read, write)
	c, exists := iom.Get(key)
	if !exists {
		t.Fatalf("Failed to find key just added %s", key)
	}
	if c.duration != duration {
		t.Errorf("Duration %v did not match %v", duration, c.duration)
	}
	if c.readLimit.Limit != read {
		t.Errorf("Read %d did not match %d", read, c.readLimit.Limit)
	}
	if c.writeLimit.Limit != write {
		t.Errorf("Write %d did not match %d", write, c.writeLimit.Limit)
	}

	readUpdate := uint64(2)
	writeUpdate := uint64(2)
	durationUpdate := 2 * time.Millisecond
	iom.Update(key, durationUpdate, readUpdate, writeUpdate)

	c, exists = iom.Get(key)
	if !exists {
		t.Fatalf("Failed to find key just added %s", key)
	}
	if c.duration != durationUpdate {
		t.Errorf("Duration %v did not match %v", durationUpdate, c.duration)
	}
	if c.readLimit.Limit != readUpdate {
		t.Errorf("Read %d did not match %d", readUpdate, c.readLimit.Limit)
	}
	if c.writeLimit.Limit != writeUpdate {
		t.Errorf("Write %d did not match %d", writeUpdate, c.writeLimit.Limit)
	}

	iom.Remove(key)
	_, exists = iom.Get(key)
	if exists {
		t.Fatalf("After delete found key %s which should not exits", key)
	}

}
func TestIOMapFindPath(t *testing.T) {

	iom := IOMap{Map: make(map[string]*IOC), Mutex: sync.RWMutex{}}

	keyFoo := "/foo/foo"
	duration := 1 * time.Millisecond
	read := uint64(1)
	write := uint64(1)
	iom.Add(keyFoo, duration, read, write)
	keyBar := "/foo/bar"
	read2 := uint64(2)
	write2 := uint64(2)
	duration2 := 2 * time.Millisecond
	iom.Add(keyBar, duration2, read2, write2)

	fooTest := "/foo/foo/test"
	c := iom.FindPath(fooTest)
	if c == nil {
		t.Fatalf("Failed to path find %s", fooTest)
	}
	if c.duration != duration {
		t.Errorf("Miss match path find %s", fooTest)
	}

	barTest := "/foo/bar/test"
	c = iom.FindPath(barTest)
	if c == nil {
		t.Fatalf("Fail to find path %s", barTest)
	}
	if c.duration != duration2 {
		t.Errorf("IOC duration %v does not match expected %v\n", c.duration, duration2)
	}
}

func TestLoadIOCConfig(t *testing.T) {
	txt := `/foo/bar/,1,1,1
	/foo/foo/,2,2,2
	/bar/foo/,3,3,3
	/bar/bar/,4,4,4`

	ioMap := LoadIOCConfig(strings.NewReader(txt))
	first := ioMap.FindPath("/foo/bar/bat")
	if first == nil {
		t.Fatalf("expected to find path '/foo/bar/bat' but got nil")
	}
	if first.readLimit.Limit != 1 {
		t.Errorf("Expected 1 but read limit is %d", first.readLimit.Limit)
	}

	last, exists := ioMap.Get("/bar/bar/")
	if !exists {
		t.Fatalf("Last entry in csv file could not find key '/bar/bar/'")
	}

	if last.writeLimit.Limit != 4 {
		t.Errorf("4 is write field however %d is the limit retrieved", last.writeLimit.Limit)
	}
}
