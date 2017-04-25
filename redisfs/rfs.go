package redisfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/go-redis/redis"
	"github.com/lateefj/shylock/qos"
	"golang.org/x/net/context"
)

const (
	pubSubRaw      = "raw"
	pubSubMessages = "messages"
)

var (
	redisPathRegex = regexp.MustCompile("/(?P<operation>.*)/(?P<topic>.*)/(?P<name>.*)")
	fuseConn       *fuse.Conn
)

func parsePath(path string) (string, string, string, error) {
	matches := redisPathRegex.FindStringSubmatch(path)
	names := redisPathRegex.SubexpNames()
	pMap := make(map[string]string)
	for i, match := range matches {
		if i != 0 {
			pMap[names[i]] = match
		}
	}
	if len(pMap) != 3 {
		return "", "", "", fmt.Errorf("Expected path wit /operation/topic/name however could not parse this path %s", path)
	}
	return pMap["operation"], pMap["topic"], pMap["name"], nil
}

// RFS ... redis file system
type RFS struct {
	Path string
	Host string
	Opts *redis.Options
}

// NewRFS ... create a RFS
func NewRFS(path string, opts *redis.Options) *RFS {
	return &RFS{Path: path, Opts: opts}
}

// Root ... root for filesystem
func (rfs *RFS) Root() (fs.Node, error) {
	return &RDir{RFS: rfs, Path: rfs.Path}, nil
}

// RDir ... directory structure
type RDir struct {
	RFS  *RFS
	Path string
}

func rTimeAttr(a *fuse.Attr) {
	a.Mtime = time.Now()
	a.Ctime = time.Now()
	a.Crtime = time.Now()
}

// IsRoot ... check to see if path is root
func (rd *RDir) IsRoot() bool {
	if rd.Path == rd.RFS.Path {
		return true
	}
	return false
}

// Attr ... set attributes
func (rd *RDir) Attr(ctx context.Context, a *fuse.Attr) error {

	if rd.IsRoot() {
		// root directory
		a.Mode = os.ModeDir | 0755
		return nil
	}
	a.Mode = os.ModeDir | 0755
	rTimeAttr(a)
	return nil
}

var _ fs.Node = (*RDir)(nil)

func isDir(path string) bool {
	if path[:len(path)-1] == "/" {
		return true
	}
	return false
}

// Lookup ... lookup a node
func (rd *RDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	path := rd.Path + "/" + req.Name

	if isDir(path) {
		return &RDir{Path: path, RFS: rd.RFS}, nil
	}
	if len(path) > len(rd.RFS.Path) {
		operation, topic, name, _ := parsePath(path[len(rd.RFS.Path):])
		switch operation {
		case "pubsub":
			return &RedisPipe{Topic: topic, FileName: name, Opts: rd.RFS.Opts}, nil
		}

	}
	return &RDir{Path: path, RFS: rd.RFS}, nil
}

// Register callback
var _ fs.NodeRequestLookuper = (*RDir)(nil)

// RedisPipe ... redis pipe like file
type RedisPipe struct {
	FileName string
	DB       string
	Topic    string
	Opts     *redis.Options
	Client   *redis.Client
	PubSub   *redis.PubSub
}

// Attr ... hmmm
func (rp *RedisPipe) Attr(ctx context.Context, a *fuse.Attr) error {
	//a.Mode = os.ModeNamedPipe | 0666
	//a.Mode = os.ModeAppend | 0666
	return nil
}

var _ fs.Node = (*RedisPipe)(nil)

func (rp *RedisPipe) connect() {
	if rp.Client == nil {
		rp.Client = redis.NewClient(rp.Opts)
	}
}

func (rp *RedisPipe) subscribe() {
	rp.connect()
	if rp.PubSub == nil {
		rp.PubSub = rp.Client.Subscribe(rp.Topic)
	}
}

// Open ... manage cache ect
func (rp *RedisPipe) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	// Disable cache
	resp.Flags |= fuse.OpenDirectIO
	return rp, nil
}

var _ fs.NodeOpener = (*RedisPipe)(nil)

// Read ... read stream
// This provides 3 reading reading strategies for the cluster.
// reader - This is just writing raw messages out
// messages - provides binary provides a single bit
// errors - stream of error messages
func (rp *RedisPipe) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	if rp.PubSub == nil {
		rp.subscribe()
	}
	buf := new(bytes.Buffer)
	switch rp.FileName {
	case pubSubRaw:
		m, err := rp.PubSub.ReceiveMessage()
		if err != nil {
			return err
		}
		buf.Write([]byte(m.Payload))
	case pubSubMessages:
		m, err := rp.PubSub.ReceiveMessage()
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, len(m.Payload))
		if err != nil {
			return err
		}
		buf.Write([]byte(m.Payload))
	}
	resp.Data = buf.Bytes()
	return err
}

var _ = fs.HandleReader(&RedisPipe{})

// Write ... Write a message to the PubSub
func (rp *RedisPipe) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	var err error
	if rp.Client == nil {
		rp.subscribe()
	}

	buf := new(bytes.Buffer)
	switch rp.FileName {
	case pubSubRaw:
		buf.Write([]byte(req.Data))
		err = rp.Client.Publish(rp.Topic, buf.String()).Err()
		if err != nil {
			return err
		}
	case pubSubMessages:
		size := len(req.Data)
		err = binary.Write(buf, binary.LittleEndian, size)
		if err != nil {
			return err
		}
		buf.Write([]byte(req.Data))
		err := rp.Client.Publish(rp.Topic, buf.String()).Err()
		if err != nil {
			return err
		}
	}
	resp.Size = len(req.Data)
	return err
}

var _ = fs.HandleWriter(&RedisPipe{})

// Mount ... Place to mount etcd
func Mount(mountPoint string, ioMap *qos.IOMap) error {

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	var err error
	db := 0
	dbEnv := os.Getenv("REDIS_DB")
	if dbEnv != "" {
		db, err = strconv.Atoi(dbEnv)
		if err != nil {
			log.Fatalf("Unable to parse REDIS_DB %s\n", dbEnv)
		}
	}
	opts := &redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	}

	fuseConn, err = fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	defer fuseConn.Close()

	filesys := NewRFS(mountPoint, opts)

	err = fs.Serve(fuseConn, filesys)
	if err != nil {
		return err
	}
	// check if the mount process has an error to report
	<-fuseConn.Ready
	if err := fuseConn.MountError; err != nil {
		log.Printf("Failed to mount because %s", err)
	}
	return err
}

// Exit ... Exit hook
func Exit() error {
	return fuseConn.Close()
}
