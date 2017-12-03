package loopback

import (
	"bytes"
	"testing"
)

func TestMemoryLoopbackMQ(t *testing.T) {
	loopback := NewHeaderMemoryLoopbackMQ("foo", map[string]string{})
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
		if bytes.Compare(*h, header) != 0 {
			t.Errorf("Expected header to be %s but was %s", string(header), string(*h))
		}
		if bytes.Compare(*b, body) != 0 {
			t.Errorf("Expected body to be %s but was %s", string(body), string(*b))
		}

	}()
	err = mf.Write(&header, &body)
	if err != nil {
		t.Fatalf("Put failed %s", err)
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
	loopback := NewHeaderMemoryLoopbackKV("foo", map[string]string{})
	mf, err := loopback.Open("bar")
	if err != nil {
		t.Fatalf("Should never error %s", err)
	}
	var config map[string]string
	err = loopback.Mount(config)
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
		if bytes.Compare(*h, header) != 0 {
			t.Errorf("Expected header to be %s but was %s", string(header), string(*h))
		}
		if bytes.Compare(*b, body) != 0 {
			t.Errorf("Expected body to be %s but was %s", string(body), string(*b))
		}
	}()
	err = mf.Write(&header, &body)
	if err != nil {
		t.Fatalf("Put failed %s", err)
	}
	names, err := loopback.List("")
	if err != nil {
		t.Fatalf("Should not have na error %s", err)
	}
	if len(names) != 1 {
		t.Errorf("Expected a single file however got %d", len(names))
	}
}
