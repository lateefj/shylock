package kafka

import (
	"fmt"
	"log"
	"os"
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
	fmt.Printf("Doing Lookup for req.Name %s\n", req.Name)
	path := kd.Path + "/" + req.Name

	fmt.Printf("Doing Lookup for path %s\n", path)
	bits := strings.Split(path[len(kd.KFS.Path):], "/")
	if isDir(path) {
		fmt.Printf("Returning directory!\n")
		return &KDir{Path: path, KFS: kd.KFS}, nil
	}
	parts := make([]string, 0)
	for _, p := range bits {
		if p != "" {
			parts = append(parts, p)
		}
	}
	fmt.Printf("Parts are size %d and values %v\n", len(parts), parts)
	// TODO: Validate possible paths Temporary!!!
	if len(parts) > 0 {
		topic := parts[0]
		if len(parts) > 1 {
			cluster := parts[len(parts)-2]
			reader := parts[len(parts)-1]

			fmt.Printf("Cluster %s and topic %s and action %s\n", cluster, topic, req.Name)
			if reader == "reader" || reader == "errors" {
				return &KPipe{Brokers: kd.KFS.Brokers, Topic: topic, Cluster: cluster, FileName: req.Name}, nil
			}
		}
	}

	fmt.Printf("OK now returning a directory with path %s\n", path)
	return &KDir{Path: path, KFS: kd.KFS}, nil
}

// Register callback
var _ fs.NodeRequestLookuper = (*KDir)(nil)

func (kd *KDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	fmt.Printf("ReadDirAll path is %s\n", kd.Path)
	start := strings.Split(kd.Path[len(kd.KFS.Path)+1:], "/")
	p := start[0]
	fmt.Printf("p is %s\n", p)
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
			return res, nil
		}
	}
	for _, t := range topics {
		fmt.Printf("Getting topic %s\n", t)
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

type KPipe struct {
	Brokers  []string
	Topic    string
	Cluster  string
	Consumer *cluster.Consumer
	FileName string
}

func (kp *KPipe) Attr(ctx context.Context, a *fuse.Attr) error {
	//a.Mode = os.ModeNamedPipe | 0666
	return nil
}

var _ fs.Node = (*KPipe)(nil)

func (kp *KPipe) connectConsumer() error {
	if kp.Consumer != nil {
		return nil
	}

	fmt.Printf("OK open being topic %s cluster %s with FileName %s\n", kp.Topic, kp.Cluster, kp.FileName)
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

func (kp *KPipe) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	err := kp.connectConsumer()
	resp.Flags |= fuse.OpenDirectIO
	return kp, err
}

var _ = fs.NodeOpener(&KPipe{})

func (kp *KPipe) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	if kp.Consumer == nil {
		kp.connectConsumer()
	}
	switch kp.FileName {
	case "reader":
		m, more := <-kp.Consumer.Messages()
		if more {
			resp.Data = m.Value
			kp.Consumer.MarkOffset(m, "")
		}
	case "errors":
		err = <-kp.Consumer.Errors()
		if err != nil {
			log.Printf("ERROR: From topic %s\n", err)
			resp.Data = []byte(err.Error())
		}
	}
	return err
}

var _ = fs.HandleReader(&KPipe{})

func (kp *KPipe) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return kp.Consumer.Close()
}

var _ = fs.HandleReader(&KPipe{})

func (kp *KPipe) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	fmt.Printf("Creating a new file %s\n", req.Name)
	return kp, kp, nil
}

var _ = fs.NodeCreater(&KPipe{})

var (
	kafkaBrokers []string
	kafkaTopic   string
)

func testProducer() {

	producer, err := NewProducer(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Printf("ERROR: Failed to create a producer: %s", err)
		return
	}
	for i := 0; ; i++ {
		fmt.Printf("Sending message: %d\n", i)
		producer.KeySend("foo", []byte(fmt.Sprintf("Test message %d\n", i)))
		time.Sleep(1 * time.Second)
	}
}

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
}

func Mount(mountPoint string) error {
	fmt.Printf("Kafka topic %s and brokers %s\n", kafkaTopic, kafkaBrokers)
	go testProducer()
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
