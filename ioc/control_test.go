package ioc

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
		case readBits, closed := <-readStream:
			if readBits != 1 {
				t.Fatalf("Expected 1 byte on readStream but got %d with closed channel %v", readBits, closed)
			}
			writeBits, closed := <-writeStream
			if writeBits != 1 {
				t.Fatalf("Expected 1 byte on writeStream but got %d with closed channel %v", writeBits, closed)
			}
			break
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Failed after 10 Microsecond")
		}
	}
}

// Make sure the stop gets called we
func TestIOCCheckoutStop(t *testing.T) {

	ioc := NewIOC(100*time.Microsecond, uint64(1), uint64(1))
	readStream := make(chan uint64, 1)
	writeStream := make(chan uint64, 1)
	go ioc.Start()
	go func() {
		err := ioc.CheckoutRead(5, readStream)
		if err == nil {
			t.Errorf("Exepected to fail before read finished")
		}
	}()

	go func() {
		err := ioc.CheckoutWrite(5, writeStream)
		if err == nil {
			t.Errorf("Expected write to fail before finished")
		}
	}()
	go ioc.Stop()
	for i := 0; i < 5; i++ {
		select {
		case _, ok := <-readStream:
			if ok {
				if i > 1 {
					t.Fatalf("Should not be able to read bytes stopped\n")
				}
			}
		case _, ok := <-writeStream:
			if ok {
				if i > 1 {
					t.Errorf("Should not be able to write bytes\n")
				}
			}
		case <-time.After(10 * time.Millisecond):
			break
		}
	}
	readStream = make(chan uint64, 1)
	writeStream = make(chan uint64, 1)
	go func() {
		err := ioc.CheckoutRead(5, readStream)
		if err == nil {
			t.Fatalf("Expected inactive state of IOC")
		}
		err = ioc.CheckoutWrite(5, writeStream)
		if err == nil {
			t.Fatalf("Expected inactive state of IOC")
		}
	}()
	select {
	case <-readStream:
		t.Fatalf("Should not be able to read bytes stopped\n")
	case <-writeStream:
		t.Fatalf("Should not be able to write bytes\n")
	case <-time.After(10 * time.Millisecond):
		break
	}
}
