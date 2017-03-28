package kafka

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	cluster "github.com/bsm/sarama-cluster"
	"golang.org/x/net/context"
)

// This device is an example implementation of an in-memory block device

type KFS struct {
	Path    string
	Brokers []string
}

func NewKFS(path string, brokers []string) *KFS {

	return &KFS{Path: path, Brokers: brokers}
}

var _ fs.FS = (*KFS)(nil)

func (kfs *KFS) Root() (fs.Node, error) {
	n := &KDir{KFS: kfs, Path: kfs.Path}
	return n, nil
}

type KDir struct {
	KFS  *KFS
	Path string
}

func (kd *KDir) IsRoot() bool {
	if kd.Path == kd.KFS.Path {
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
		a.Mode = os.ModeDir | 0755
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

var _ fs.Node = (*KDir)(nil)

func (kd *KDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	path := kd.Path + "/" + req.Name

	bits := strings.Split(path[len(kd.KFS.Path):], "/")
	if isDir(path) {
		return &KDir{Path: path, KFS: kd.KFS}, nil
	}
	parts := make([]string, 0)
	for _, p := range bits {
		if p != "" {
			parts = append(parts, p)
		}
	}
	// TODO: Validate possible paths Temporary!!!
	if len(parts) > 0 {
		topic := parts[0]
		if len(parts) > 1 {
			cluster := parts[len(parts)-2]
			reader := parts[len(parts)-1]

			//XXX: Uhg this is so bad! FIX ME !!
			if reader == "reader" || reader == "errors" || reader == "writer" || reader == "messages" {
				return &ClusterPipe{Brokers: kd.KFS.Brokers, Topic: topic, Cluster: cluster, FileName: req.Name}, nil
			}
		}
	}

	return &KDir{Path: path, KFS: kd.KFS}, nil
}

// Register callback
var _ fs.NodeRequestLookuper = (*KDir)(nil)

func (kd *KDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	start := strings.Split(kd.Path[len(kd.KFS.Path)+1:], "/")
	p := start[0]
	topics, err := Topics(kd.KFS.Brokers)
	var res []fuse.Dirent
	if err != nil {
		log.Printf("ERROR: getting topics %s\n", err)
		return res, err
	}
	for _, t := range topics {
		if t == p {
			var reader fuse.Dirent
			reader.Name = "reader"
			reader.Type = fuse.DT_FIFO
			res = append(res, reader)
			var messages fuse.Dirent
			messages.Name = "messages"
			messages.Type = fuse.DT_FIFO
			res = append(res, messages)
			var errors fuse.Dirent
			errors.Name = "errors"
			errors.Type = fuse.DT_FIFO
			res = append(res, errors)
			var writer fuse.Dirent
			writer.Name = "writer"
			writer.Type = fuse.DT_FIFO
			res = append(res, writer)
			return res, nil
		}
	}
	for _, t := range topics {
		var de fuse.Dirent
		de.Type = fuse.DT_Dir
		de.Name = t
		res = append(res, de)
	}

	return res, nil
}

// Register callback
var _ fs.HandleReadDirAller = (*KDir)(nil)

//TODO: Implement ReadDirAll for helpers
// partitions/ - List of the partitions
// cluster/ - Cluster consumer

type ClusterPipe struct {
	Brokers  []string
	Topic    string
	Cluster  string
	FileName string
	Consumer *cluster.Consumer
	Producer *Producer
}

func (kp *ClusterPipe) Attr(ctx context.Context, a *fuse.Attr) error {
	//a.Mode = os.ModeNamedPipe | 0666
	//a.Mode = os.ModeAppend | 0666
	return nil
}

var _ fs.Node = (*ClusterPipe)(nil)

func (kp *ClusterPipe) connectConsumer() error {
	if kp.Consumer != nil {
		return nil
	}
	kc := KafkaConfig{Brokers: kp.Brokers, Topics: []string{kp.Topic}}
	c := ClusterConfig{KafkaConfig: kc, Name: kp.Cluster}
	consumer, err := NewConsumer(c)
	if err != nil {
		log.Printf("ERROR: Failed to connect to kafka %s\n", err)
		return err
	}
	kp.Consumer = consumer
	return nil
}

func (kp *ClusterPipe) connectProducer() error {
	if kp.Producer != nil {
		return nil
	}
	var err error
	kp.Producer, err = NewProducer(kafkaBrokers, kp.Topic)
	if err != nil {
		log.Printf("ERROR: Failed to create a producer: %s", err)
		return err
	}
	return nil
}

func (kp *ClusterPipe) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags |= fuse.OpenDirectIO
	return kp, nil
}

var _ = fs.NodeOpener(&ClusterPipe{})

func (kp *ClusterPipe) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	var err error
	if kp.Consumer != nil {
		err = kp.Consumer.Close()
		kp.Consumer = nil
	}
	if kp.Producer != nil {
		err = kp.Producer.Close()
		kp.Producer = nil
	}
	return err
}

func (kp *ClusterPipe) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	return kp, kp, nil
}

var _ = fs.NodeCreater(&ClusterPipe{})

// This provides 3 reading reading strategies for the cluster.
// reader - This is just writing raw messages out
// messages - provides binary provides a single bit
// errors - stream of error messages
func (kp *ClusterPipe) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	if kp.Consumer == nil {
		kp.connectConsumer()
	}
	buf := new(bytes.Buffer)
	switch kp.FileName {
	case "reader":
		m, more := <-kp.Consumer.Messages()
		if more {
			buf.Write(m.Value)
		}
	case "messages":
		m, more := <-kp.Consumer.Messages()
		if more {
			err = binary.Write(buf, binary.LittleEndian, len(m.Value))
			buf.Write(m.Value)
		}

	case "errors":
		err = <-kp.Consumer.Errors()
		if err != nil {
			log.Printf("ERROR: From topic %s\n", err)
			buf.WriteString(err.Error())
		}
	}
	resp.Data = buf.Bytes()
	return err
}

var _ = fs.HandleReader(&ClusterPipe{})

func (kp *ClusterPipe) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	if kp.Producer == nil {
		kp.connectProducer()
	}
	// TODO: Figure out how to do a KeySend
	kp.Producer.Send(req.Data)
	resp.Size = len(req.Data)
	return nil
}

var _ = fs.HandleWriter(&ClusterPipe{})

var (
	kafkaBrokers   []string
	kafkaTopic     string
	autoCommitSize int
	autoCommitTime *time.Duration
)

func init() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		kafkaBrokers = []string{"127.0.0.1:9092"}
	} else {
		kafkaBrokers = strings.Split(brokers, ",")
	}

	kafkaTopic = os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "my_topic"
	}

	autoCommitDuration := os.Getenv("KAFKA_AUTOCOMMIT_DURATION")
	if autoCommitDuration == "" {
		autoCommitTime = nil
	} else {
		t, err := strconv.Atoi(autoCommitDuration)
		if err != nil {
			log.Fatalf("Unable to parse KAFKA_AUTOCOMMIT_DURATION Milliseconds %s\n", autoCommitDuration)
		}
		d := time.Duration(t) * time.Millisecond
		autoCommitTime = &d
	}
	var err error
	autoCommitBatch := os.Getenv("KAFKA_AUTOCOMMIT_BATCH")
	if autoCommitBatch == "" {
		autoCommitSize = -1
	} else {
		autoCommitSize, err = strconv.Atoi(autoCommitBatch)
		if err != nil {
			log.Fatalf("Unable to parse KAFKA_AUTOCOMMIT_BATCH size %s\n", autoCommitBatch)
		}
	}
}

func Mount(mountPoint string) error {
	fmt.Printf("Kafka topic %s and brokers %s\n", kafkaTopic, kafkaBrokers)
	//go testProducer()
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := NewKFS(mountPoint, kafkaBrokers)

	if err := fs.Serve(c, filesys); err != nil {
		log.Printf("ERROR: Failed to server because %s\n", err)
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Printf("ERROR: Failed to mount because %s\n", err)
		return err
	}
	return nil

}
