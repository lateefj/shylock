package ioc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/lateefj/mctest"
)

func TestHandleGetDeletePutPost(t *testing.T) {
	k := "test"
	duration := 1
	read := uint64(1)
	write := uint64(1)
	iom := NewIOMap()

	jsonc := &jsonIOC{Key: k, Duration: duration, ReadLimit: read, WriteLimit: write}
	bits, err := json.Marshal(jsonc)
	if err != nil {
		t.Fatalf("Failed to marshal %s", err)
	}
	rest := NewRest(iom)
	body := ioutil.NopCloser(bytes.NewReader(bits))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/key/%s", k), body)
	resp := mctest.NewMockTestResponse(t)

	rest.Default(resp, req)
	_, exists := iom.Get(k)
	if !exists {
		t.Fatalf("Expected ioc to be added via rest URL")
	}
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/key/%s", k), nil)
	resp = mctest.NewMockTestResponse(t)
	rest.Default(resp, req)
	if !resp.AssertCode(http.StatusOK) {
		t.Fatal("Status code was not OK")
	}
	if !resp.AssertJson(&jsonIOC{}, jsonc) {
		t.Fatalf("Expected equal response however got %s and expected %v", resp.String(), jsonc)
	}

	writeUpdate := uint64(2)
	jsonc.WriteLimit = writeUpdate
	bits, err = json.Marshal(jsonc)
	if err != nil {
		t.Fatalf("Failed to marshal %s", err)
	}
	body = ioutil.NopCloser(bytes.NewReader(bits))
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/key/%s", k), body)
	resp = mctest.NewMockTestResponse(t)

	rest.Default(resp, req)
	ioc, _ := iom.Get(k)
	if ioc.writeLimit.Limit != writeUpdate {
		t.Fatalf("Expected update of %d but got %d", writeUpdate, ioc.writeLimit.Limit)
	}

}
