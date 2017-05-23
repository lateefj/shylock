package qos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type jsonIOC struct {
	Key        string `json:"key"`
	ReadLimit  uint64 `json:"read_limit"`
	WriteLimit uint64 `json:"write_limit"`
}

func unmarshalIOC(req *http.Request) (*jsonIOC, error) {

	bits, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	ioc := &jsonIOC{}
	err = json.Unmarshal(bits, ioc)
	return ioc, err
}

// Rest ... Structure that should make it easier to modify an IOMap mount at runtime
type Rest struct {
	iom *IOMap
}

// NewRest ... Create an new instance of the Rest struct
func NewRest(m *IOMap) *Rest {
	return &Rest{iom: m}
}

func (r *Rest) handleGet(key string, ioc *IOC, w http.ResponseWriter) {
	jioc := &jsonIOC{Key: key, ReadLimit: ioc.readLimit.Limit, WriteLimit: ioc.writeLimit.Limit}
	bits, err := json.Marshal(jioc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error %s", err.Error())
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bits)

}

func (r *Rest) updateIOC(key string, ioc *IOC, req *http.Request, w http.ResponseWriter) {
	tmp, err := unmarshalIOC(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Unmarshal error: %s", err.Error())
	}
	r.iom.Update(key, 1*time.Second, tmp.ReadLimit, tmp.WriteLimit)
	w.WriteHeader(http.StatusOK)
}
func (r *Rest) addIOC(key string, req *http.Request, w http.ResponseWriter) {
	tmp, err := unmarshalIOC(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Unmarshal error: %s", err.Error())
	}
	r.iom.Add(key, 1*time.Second, tmp.ReadLimit, tmp.WriteLimit)
	w.WriteHeader(http.StatusOK)
}

// Default ... Default handler for a mounted system
func (r *Rest) Default(w http.ResponseWriter, req *http.Request) {

	key := req.URL.Path[len("/key/"):]
	var ioc *IOC
	exists := false
	if req.Method != http.MethodPost {
		ioc, exists = r.iom.Get(key)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "could not find key %s", key)
		}
	}

	switch req.Method {
	case http.MethodGet:
		r.handleGet(key, ioc, w)
	case http.MethodDelete:
		r.iom.Remove(key)
	case http.MethodPut:
		r.updateIOC(key, ioc, req, w)
	case http.MethodPost:
		r.addIOC(key, req, w)
	}
}

// Setup ... This associates the IOMap with rest endpoints
func Setup(iom *IOMap) {
	rest := NewRest(iom)
	http.HandleFunc("/key/", rest.Default)
	//router.HandleFunc("/path/", rest.FindPath)
}
