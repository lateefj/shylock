package loopback

import (
	"errors"
	"strings"

	"github.com/lateefj/shylock/api"
)

var unknownError error

const (
	FSMemoryLoopbackHeaderMQ = "LOOPBACK_HEADER_MQ"
	FSMemoryLoopbacHeaderKV  = "LOOPBACK_HEADER_KV"
	FSMemoryLoopbacKV        = "LOOPBACK_KV"
)

func init() {
	unknownError = errors.New("Unknown Error")
	/*api.RegisterHeaderDevice(FSMemoryLoopbackHeaderMQ, NewHeaderMemoryLoopbackMQ)
	api.RegisterHeaderDevice(FSMemoryLoopbacHeaderKV, NewHeaderMemoryLoopbackKV)*/
	api.RegisterSimpleDevice(FSMemoryLoopbacKV, NewMemoryLoopbackKV)
}

type HeaderMemoryFileMQ struct {
	queue chan [][]byte
}

func (mf *HeaderMemoryFileMQ) Read() (header, body []byte, err error) {
	m := <-mf.queue
	return m[0], m[1], nil

}
func (mf *HeaderMemoryFileMQ) Write(offset int, header, body []byte) (int, error) {
	mf.queue <- [][]byte{header, body}
	return len(header) + len(body), nil
}

func (mf *HeaderMemoryFileMQ) Close() {
	close(mf.queue)
}

type HeaderMemoryLoopbackMQ struct {
	queues map[string]*HeaderMemoryFileMQ
}

func NewHeaderMemoryLoopbackMQ(mountPoint string, config []byte) api.HeaderDevice {
	return &HeaderMemoryLoopbackMQ{queues: make(map[string]*HeaderMemoryFileMQ)}
}

// Mount ... Noop
func (mmq *HeaderMemoryLoopbackMQ) Mount(config []byte) error {
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
		q = &HeaderMemoryFileMQ{queue: make(chan [][]byte)}
		mmq.queues[path] = q
	}
	return q, nil
}

type HeaderMemoryFileKV struct {
	header []byte
	body   []byte
}

func (mf *HeaderMemoryFileKV) Read() (header, body []byte, err error) {
	return mf.header, mf.body, nil

}
func (mf *HeaderMemoryFileKV) Write(offset int, header, body []byte) (int, error) {
	mf.header = header
	for i := 0; i < len(body); i++ {
		if len(mf.body)+offset < len(body) {
			mf.body = append(mf.body, body[i])
		} else {
			mf.body[offset+i] = body[i]
		}
	}
	return len(header) + len(body), nil
}

type HeaderMemoryLoopbackKV struct {
	db map[string]*HeaderMemoryFileKV
}

func NewHeaderMemoryLoopbackKV(mountPoint string, config []byte) api.HeaderDevice {
	return &HeaderMemoryLoopbackKV{db: make(map[string]*HeaderMemoryFileKV)}
}

func (mkv *HeaderMemoryLoopbackKV) Mount(config []byte) error {
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

type MemoryFileKV struct {
	body []byte
}

func (mf *MemoryFileKV) Read() (body []byte, err error) {
	return mf.body, nil

}
func (mf *MemoryFileKV) Write(body []byte) error {
	mf.body = body
	return nil
}
func (mf *MemoryFileKV) Close() error {
	return nil
}

type MemoryLoopbackKV struct {
	db map[string]*MemoryFileKV
}

func NewMemoryLoopbackKV(mountPoint string, config []byte) api.SimpleDevice {
	return &MemoryLoopbackKV{db: make(map[string]*MemoryFileKV)}
}

func (mkv *MemoryLoopbackKV) Mount(config []byte) error {
	return nil
}
func (mkv *MemoryLoopbackKV) Unmount() error {
	return nil
}
func (mkv *MemoryLoopbackKV) List(path string) ([]string, error) {
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
				if len(s) > 0 {
					files = append(files, k)
				}
			}
		}
	}
	return files, nil
}

func (mkv *MemoryLoopbackKV) Open(path string) (api.SimpleFile, error) {
	f, exists := mkv.db[path]
	if !exists {
		f = &MemoryFileKV{}
		mkv.db[path] = f
	}
	return f, nil
}

func (mkv *MemoryLoopbackKV) Remove(path string) error {
	_, exists := mkv.db[path]
	if exists {
		delete(mkv.db, path)
	}
	return nil
}
