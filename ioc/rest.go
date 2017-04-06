package ioc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type jsonIOC struct {
	Key        string `json:"key"`
	Duration   uint64 `json:"duration"`
	ReadLimit  uint64 `json:"read_limit"`
	WriteLimit uint64 `json:"write_limit"`
}

// Rest ... Structure that should make it easier to modify an IOMap mount at runtime
type Rest struct {
	iom *IOMap
}

// NewRest ... Create an new instance of the Rest struct
func NewRest(m *IOMap) *Rest {
	return &Rest{iom: m}
}

func (r *rest) handleGet(key string, w http.ResponseWriter) {
	ioc := r.iom.Get(key)
	if ioc == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "could not find key %s", key)
	}

	jioc := jsonIOC{Key: key, Duration: ioc.Duratio, ReadLimit: ioc.readLimit.Limit, WriteLimit: ioc.writeLimit.Limit}
	bits, err := json.Marshal(jioc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error %s", err)
	}
	w.Write(bits)

}

// Default ... Default handler for a mounted system
func (r *Rest) Default(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	key, exists := vars["key"]
	switch req.Method {
	case http.MethodGet:
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Could not find 'key' field in list of vars got %s", key)
			return
		}
		handleGet(key, w)

	case http.MethodDelete:
		if exists {
			r.iom.Remove(key)
		}
	case http.MethodPost || http.MethodPut:
		r.updateIOC(key, req)

	}

}
