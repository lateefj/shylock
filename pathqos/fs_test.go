package pathqos

import (
	"sync"
	"testing"

	//"golang.org/x/net/context"

	"github.com/lateefj/shylock/qos"
)

func TestSFSRoot(t *testing.T) {
	iom := &qos.IOMap{Map: make(map[string]*qos.IOC), Mutex: sync.RWMutex{}}
	path := "/tmp/sfs"
	sfs := NewSFS(path, iom)
	_, err := sfs.Root()
	if err != nil {
		t.Error(err)
	}
}
