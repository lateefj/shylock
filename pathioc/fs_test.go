package pathioc

import (
	"sync"
	"testing"

	//"golang.org/x/net/context"

	"github.com/lateefj/shylock/ioc"
)

func TestSFSRoot(t *testing.T) {
	iom := &ioc.IOMap{Map: make(map[string]*ioc.IOC), Mutex: sync.RWMutex{}}
	path := "/tmp/sfs"
	sfs := NewSFS(path, iom)
	_, err := sfs.Root()
	if err != nil {
		t.Error(err)
	}
}
