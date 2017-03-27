package pathioc

import (
	"testing"
	//"golang.org/x/net/context"
)

func TestSFSRoot(t *testing.T) {
	path := "/tmp/sfs"
	sfs := NewSFS(path)
	_, err := sfs.Root()
	if err != nil {
		t.Error(err)
	}
}
