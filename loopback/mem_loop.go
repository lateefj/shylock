package loopback

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lateefj/shylock/api"
)

var unknownError error

func init() {
	unknownError = errors.New("Unknown Error")
	api.RegisterHeaderDevice("memory_loopback_mq", NewHeaderMemoryLoopbackMQ)
	api.RegisterHeaderDevice("memory_loopback_kv", NewHeaderMemoryLoopbackKV)
}

type HeaderMemoryFileMQ struct {
	queue chan []*[]byte
}

func (mf *HeaderMemoryFileMQ) Read() (header, body *[]byte, err error) {
	m := <-mf.queue
	return m[0], m[1], nil

}
func (mf *HeaderMemoryFileMQ) Write(header, body *[]byte) error {
	mf.queue <- []*[]byte{header, body}
	return nil
}

func (mf *HeaderMemoryFileMQ) Close() {
	close(mf.queue)
}

type HeaderMemoryLoopbackMQ struct {
	queues map[string]*HeaderMemoryFileMQ
}

func NewHeaderMemoryLoopbackMQ(mountPoint string, config map[string]string) api.HeaderDevice {
	return &HeaderMemoryLoopbackMQ{queues: make(map[string]*HeaderMemoryFileMQ)}
}

// Mount ... Noop
func (mmq *HeaderMemoryLoopbackMQ) Mount(config map[string]string) error {
	return nil
}

// Unmount ... Close all open files and remove files from map
func (mmq *HeaderMemoryLoopbackMQ) Unmount() error {
	for k, q := range mmq.queues {
		q.Close()
		delete(mmq.queues, k)
	}
	return nil
}

func (mmq *HeaderMemoryLoopbackMQ) List(path string) ([]string, error) {
	files := make([]string, 0)
	for k, _ := range mmq.queues {
		files = append(files, k)
	}
	return files, nil
}

func (mmq *HeaderMemoryLoopbackMQ) Open(path string) (api.HeaderFile, error) {
	q, exists := mmq.queues[path]
	if !exists {
		q = &HeaderMemoryFileMQ{queue: make(chan []*[]byte)}
		mmq.queues[path] = q
	}
	return q, nil
}

type HeaderMemoryFileKV struct {
	header *[]byte
	body   *[]byte
}

func (mf *HeaderMemoryFileKV) Read() (header, body *[]byte, err error) {
	return mf.header, mf.body, nil

}
func (mf *HeaderMemoryFileKV) Write(header, body *[]byte) error {
	mf.header = header
	mf.body = body
	return nil
}

type HeaderMemoryLoopbackKV struct {
	db map[string]*HeaderMemoryFileKV
}

func NewHeaderMemoryLoopbackKV(mountPoint string, config map[string]string) api.HeaderDevice {
	return &HeaderMemoryLoopbackKV{db: make(map[string]*HeaderMemoryFileKV)}
}

func (mkv *HeaderMemoryLoopbackKV) Mount(config map[string]string) error {
	return nil
}
func (mkv *HeaderMemoryLoopbackKV) Unmount() error {
	return nil
}
func (mkv *HeaderMemoryLoopbackKV) List(path string) ([]string, error) {
	files := make([]string, 0)
	for k, _ := range mkv.db {
		if len(k) > len(path) {
			p := k[:len(path)]
			if strings.Compare(path, p) == 0 {
				s := k[len(path):]
				next := strings.Index(s, "/")
				if next > 0 {
					s = s[:next]
				}
				fmt.Printf("From List item is  %s\n", s)
				if len(s) > 0 {
					files = append(files, k)
				}
			}
		}
	}
	return files, nil
}

func (mkv *HeaderMemoryLoopbackKV) Open(path string) (api.HeaderFile, error) {
	f, exists := mkv.db[path]
	if !exists {
		f = &HeaderMemoryFileKV{}
		mkv.db[path] = f
	}
	return f, nil
}
