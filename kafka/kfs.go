package kafka

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
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
		//a.Mode = os.ModeNamedPipe
		//a.Mode = os.ModeTemporary
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
	fmt.Printf("Bits are %v\n", bits)
	if isDir(path) {
		fmt.Printf("Returning directory!\n")
		return &KDir{Path: path, KFS: kd.KFS}, nil
	}
	parts := make([]string, 0)
	for _, p := range bits {
		fmt.Printf("Part is %s\n", p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	fmt.Printf("Parts are %v\n", parts)
	// TODO: Validate possible paths Temporary!!!
	if len(parts) > 0 {
		topic := parts[0]
		if len(parts) > 1 {
			cluster := parts[len(parts)-1]
			reader := parts[len(parts)-2]
			if reader == "reader" {
				fmt.Printf("Cluster %s and topic %s and action %s\n", cluster, topic, req.Name)
				kp := &KPipe{Brokers: kd.KFS.Brokers, Topic: topic, Cluster: cluster, Action: req.Name}
				kp.connectConsumer()
				return kp, nil
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
	p := kd.Path[len(kd.KFS.Path):]
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
			reader.Type = fuse.DT_Dir
			res = append(res, reader)
			var bar fuse.Dirent
			bar.Name = "bar"
			bar.Type = fuse.DT_Char
			res = append(res, bar)
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
	Action   string
	Consumer *cluster.Consumer
}

func (kp KPipe) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeNamedPipe | 0666
	return nil
}

var _ fs.Node = (*KPipe)(nil)

func (kp KPipe) connectConsumer() error {
	if kp.Consumer != nil {
		return nil
	}

	fmt.Printf("OK open being topic %s cluster %s with action %s\n", kp.Topic, kp.Cluster, kp.Action)
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

func (kp KPipe) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	err := kp.connectConsumer()
	resp.Flags |= fuse.OpenNonSeekable
	return &kp, err
}

var _ = fs.NodeOpener(&KPipe{})

func (kp KPipe) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	fmt.Printf("OK woot trying to read kafka topic %s\n", kp.Cluster)
	if kp.Consumer == nil {
	}

	select {
	case m := <-kp.Consumer.Messages():
		fmt.Printf("Message from topic %s Key %s and body is %s\n", m.Topic, string(m.Key), string(m.Value))
		fuseutil.HandleRead(req, resp, m.Value)

		return nil
	case e := <-kp.Consumer.Errors():
		log.Printf("ERROR: From topic %s\n", e)
		return err
	}
}

/*
type KFile struct {
	Brokers []string
	Topic   string
	Cluster string
	Action  string
}

func (sf *KFile) Attr(ctx context.Context, a *fuse.Attr) error {
	t := time.Now()
	a.Mtime = t
	a.Ctime = t
	a.Crtime = t
	a.Mode = os.ModeNamedPipe | 0755
	//a.Size = 512
	return nil
}

var _ fs.Node = (*KFile)(nil)

func (kf *KFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	fmt.Printf("OK open being topic %s cluster %s with action %s\n", kf.Topic, kf.Cluster, kf.Action)
	kc := KafkaConfig{Brokers: kf.Brokers, Topics: []string{kf.Topic}}
	c := ClusterConfig{KafkaConfig: kc, Name: kf.Cluster}
	consumer, err := NewConsumer(c)
	if err != nil {
		log.Printf("ERROR: Failed to connect to kafka %s\n", err)
		return kf, err
	}
	resp.Flags |= fuse.OpenNonSeekable
	return &KFHandler{kf, consumer}, nil
}

var _ = fs.NodeOpener(&KFile{})

type KFHandler struct {
	KFile    *KFile
	Consumer *cluster.Consumer
}

var _ fs.Handle = (*KFHandler)(nil)

func (kfh *KFHandler) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return kfh.Consumer.Close()
}

var _ = fs.HandleReleaser(&KFHandler{})

func (kfh *KFHandler) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	fmt.Printf("OK woot trying to read kafka topic %s\n", kfh.KFile.Cluster)

	select {
	case m := <-kfh.Consumer.Messages():
		fmt.Printf("Message from topic %s Key %s and body is %s\n", m.Topic, string(m.Key), string(m.Value))
		fuseutil.HandleRead(req, resp, m.Value)

		return nil
	case e := <-kfh.Consumer.Errors():
		log.Printf("ERROR: From topic %s\n", e)
		return err
	}
}

var _ = fs.HandleReader(&KFHandler{})
*/
