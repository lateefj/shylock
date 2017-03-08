package kafka

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

// This device is an example implementation of an in-memory block device

type KFS struct {
	Path string
}

func NewKFS(path string, p *sarama.Producer, c *cluster.Consuer) *KFS {

	&KFS{Path: path}
}

var _ fs.FS = (*KFS)(nil)

func (kfs *KFS) Root() (fs.Node, error) {
	n := &KDir{KFS: kfs, Path: kfs.Path}
	return n, nil
}

type KDir struct {
	KFS      *KFS
	Path     string
	Producer *sarama.Producer
	Consumer *cluster.Consumer
}

func (kd *KDir) IsRoot() bool {
	if kd.Path == kd.SFS.Path {
		return true
	}
	return false
}

func kTimeAttr(a *fuse.Attr) {
	a.Mtime = time.Now()
	a.Ctime = time.Now()
	a.Crtime = time.Now()
}

func (kd *KDir) Attr(ctx context.Context, a *fuse.Attr) error {

	if kd.IsRoot() {
		// root directory
		a.Mode = os.ModeDir | 0755
		return nil
	} else {
		a.Mode = os.ModeNamedPipe
	}
	kTimeAttr(a)
	return nil
}

func isDir(path string) bool {
	if path[:len(path)-1] == "/" {
		return true
	}
	return false
}
func (kd *KDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	path := kd.Path + "/" + req.Name
	if isDir(path) {
		return &KDir{Path: path, KFS: kd.KFS}, nil
	}
	parts = strings.Split(req.Name, "/")
	// TODO: Validate possible paths Temporary!!!
	if parts[len(parts)] != "reader" {
		return nil, errors.New(fmt.Sprintf("Path %s does not exists ...", path))

	}
	f = parts[len(parts)]
	cluster = parts[len(parts-1)]
	topic = parts[len(parts-2)]

	return &KFile{Topic: topic, Cluster: cluster, Action: f}, nil
}

// Register lookup callback
var _ = fs.NodeRequestLookuper(&KDir{})

//TODO: Implement ReadDirAll for helpers
// partitions/ - List of the partitions
// cluster/ - Cluster consumer

type KFile struct {
	Topic   string
	Cluster string
	Action  string
}

var _ fs.Node = (*KFile)(nil)

func (kf *KFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return kf, nil
}

var _ = fs.NodeOpener(&KFile{})

func (kf *KFile) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	// TODO: Need to close the connection
	return nil
}

var _ fs.HandleReleaser = (*SFile)(nil)

func (kf *KFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fmt.Printf("OK woot trying to read kafka topic %s\n", kf.Cluster)

	kc := kafka.KafkaConfig{Brokers: kafkaBrokers, Topics: []string{kf.Topic}}
	c := kafka.ClusterConfig{KafkaConfig: kc, Name: kf.Cluster}
	consumer, err := kafka.NewConsumer(c)
	if err != nil {
		log.Printf("ERROR: Failed to connect to kafka %s\n", err)
		return
	}

	for {
		select {
		case m := <-consumer.Messages():
			fmt.Printf("Message from topic %s Key %s and body is %s\n", m.Topic, string(m.Key), string(m.Value))
		case e := <-consumer.Errors():
			log.Printf("ERROR: From topic %s\n", e)
		}
	}
}

var _ = fs.HandleReader(&KFile{})
