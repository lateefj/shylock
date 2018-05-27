package loopback

import (
	"bytes"
	"testing"
)

func TestMemoryLoopbackMQ(t *testing.T) {
	loopback := NewHeaderMemoryLoopbackMQ("foo", []byte("config_data"))
	mf, err := loopback.Open("bar")
	if err != nil {
		t.Fatalf("Should never error %s", err)
	}
	header := []byte("header")
	body := []byte("body")
	go func() {
		h, b, err := mf.Read()
		if err != nil {
			t.Fatalf("Failed to get with error %s", err)
		}
		if bytes.Compare(h, header) != 0 {
			t.Errorf("Expected header to be %s but was %s", string(header), string(h))
		}
		if bytes.Compare(b, body) != 0 {
			t.Errorf("Expected body to be %s but was %s", string(body), string(b))
		}

	}()
	written, err := mf.Write(0, header, body)
	if err != nil {
		t.Fatalf("Put failed %s", err)
	}
	if written != len(body)+len(header) {
		t.Fatalf("Expected written to be header + body %d but is %d", (len(header) + len(body)), written)
	}
	names, err := loopback.List("")
	if err != nil {
		t.Fatalf("Should not have na error %s", err)
	}
	if len(names) != 1 {
		t.Errorf("Expected a single file however got %d", len(names))
	}
}

func TestMemoryLoopbackKV(t *testing.T) {
	loopback := NewHeaderMemoryLoopbackKV("foo", []byte("config_data"))
	mf, err := loopback.Open("bar")
	if err != nil {
		t.Fatalf("Should never error %s", err)
	}
	err = loopback.Mount([]byte("Config Test Data"))
	if err != nil {
		t.Fatalf("Should never fail %s", err)
	}
	defer func() {
		err = loopback.Unmount()
		if err != nil {
			t.Fatalf("Should never fail %s", err)
		}
	}()
	header := []byte("header")
	body := []byte("body")
	go func() {
		h, b, err := mf.Read()
		if err != nil {
			t.Fatalf("Failed to get with error %s", err)
		}
		if bytes.Compare(h, header) != 0 {
			t.Errorf("Expected header to be %s but was %s", string(header), string(h))
		}
		if bytes.Compare(b, body) != 0 {
			t.Errorf("Expected body to be %s but was %s", string(body), string(b))
		}
	}()
	written, err := mf.Write(0, header, body)
	if err != nil {
		t.Fatalf("Put failed %s", err)
	}
	if written != len(body)+len(header) {
		t.Fatalf("Expected written to be header + body %d but is %d", (len(header) + len(body)), written)
	}
	names, err := loopback.List("")
	if err != nil {
		t.Fatalf("Should not have na error %s", err)
	}
	if len(names) != 1 {
		t.Errorf("Expected a single file however got %d", len(names))
	}
}
