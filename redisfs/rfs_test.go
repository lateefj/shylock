package redisfs

import (
	"testing"
)

func TestParsePath(t *testing.T) {
	p := "/operation/topic/name"
	op, top, n, err := parsePath(p)
	if err != nil {
		t.Errorf("Failed to parse %s error: %s", p, err)
	}
	if op != "operation" {
		t.Errorf("Expect op 'operation' however is %s", op)
	}
	if top != "topic" {
		t.Errorf("Expect t 'topic' however is %s", top)
	}

	if n != "name" {
		t.Errorf("Expect n 'name' however is %s", n)
	}
}
