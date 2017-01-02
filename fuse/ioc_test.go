package main

import (
	"testing"
	"time"
)

func TestIOCCheckout(t *testing.T) {

	ioc := NewIOC(100*time.Microsecond, uint64(1), uint64(1))
	readStream := make(chan uint64, 1)
	writeStream := make(chan uint64, 1)
	go ioc.Start()
	go ioc.CheckoutRead(5, readStream)
	go ioc.CheckoutWrite(5, writeStream)
	for i := 0; i < 5; i++ {
		select {
		case readBits := <-readStream:
			if readBits != 1 {
				t.Fatalf("Expected 1 byte on readStream but got %d", readBits)
			}
			writeBits := <-writeStream
			if writeBits != 1 {
				t.Fatalf("Expected 1 byte on writeStream but got %d", writeBits)
			}
			break
		case <-time.After(2 * time.Millisecond):
			t.Fatal("Failed after 10 Microsecond")
		}
	}
}
