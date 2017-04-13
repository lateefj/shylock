// Package etcd ... This was influenced by https://github.com/Merovius/etcdfs/blob/master/etcdfs.go since both fuse and etcd are the same.
package etcd

import (
	"bytes"
	"hash/crc64"
	"log"
	"os"
	"path"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/coreos/etcd/client"
	"github.com/lateefj/shylock/qos"
	"golang.org/x/net/context"
)

var (
	crc      = crc64.MakeTable(crc64.ECMA)
	fuseConn *fuse.Conn
)

// Get the checksum for a key
func checksum(key string) uint64 {
	return crc64.Checksum([]byte(key), crc)
}

// EDFS ... etcd root structure
type EDFS struct {
	Path     string
	KApi     client.KeysAPI
	ReadOnly bool
}

// NewEDFS ... Create a new EDFS instance
func NewEDFS(mountPoint string, servers []string, readOnly bool) (*EDFS, error) {

	cfg := client.Config{
		Endpoints: servers,
		Transport: client.DefaultTransport,
	}

	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	kapi := client.NewKeysAPI(c)
	return &EDFS{Path: mountPoint, KApi: kapi, ReadOnly: readOnly}, nil
}

// fsNode ... Looks up the key in etcd and handles the appropriate error
func (ed *EDFS) fsNode(ctx context.Context, key string) (fs.Node, error) {
	if key == ed.Path {
		key = "/"
	}
	v, err := ed.KApi.Get(ctx, key, &client.GetOptions{
		Sort:   true,
		Quorum: true,
	})
	if err != nil {
		switch err.(type) {
		case *client.ClusterError:
			log.Printf("ERROR: Cluster connection error %s", err)
			return nil, fuse.ENOSYS
		}

		switch err.(client.Error).Code {
		case client.ErrorCodeKeyNotFound:
			return nil, fuse.ENOENT
		case client.ErrorCodeNotDir:
			return nil, fuse.Errno(syscall.ENOTDIR)
		case client.ErrorCodeUnauthorized:
			return nil, fuse.Errno(syscall.EPERM)
		default:
			return nil, err
		}
	}

	var node *client.Node
	if v != nil {
		node = v.Node
	}

	if (node != nil && node.Dir) || key[:len(key)-1] == "/" {
		return &EDDir{FS: ed, Node: node, Key: key}, nil
	}
	return &EDFile{FS: ed, Node: node, Key: key}, nil
}

// Root ... Required for fuse system
func (ed *EDFS) Root() (fs.Node, error) {
	return ed.fsNode(context.Background(), ed.Path)
}

// EDDir ... Directory entry in etcd
type EDDir struct {
	Key  string
	FS   *EDFS
	Node *client.Node
}

// Attr ... Required for fuse
func (e *EDDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = checksum(e.Key)
	attr.Mode = os.ModeDir | 0555
	return nil
}

var _ fs.Node = (*EDDir)(nil)

// Mkdir ... Fuse make directory hook
func (e *EDDir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	resp, err := e.FS.KApi.Set(ctx, req.Name, "", &client.SetOptions{
		Dir: true,
	})
	if err != nil {
		return nil, err
	}
	return &EDDir{Key: req.Name, Node: resp.Node, FS: e.FS}, nil
}

var _ = fs.NodeMkdirer(&EDDir{})

// ReadDirAll ... Get everything in a directory
func (e *EDDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	// Refresh the directory listing
	v, err := e.FS.KApi.Get(ctx, e.Key, &client.GetOptions{
		Sort:   true,
		Quorum: true,
	})
	if err != nil {
		return make([]fuse.Dirent, 0), err
	}
	e.Node = v.Node
	nodes := make([]fuse.Dirent, len(e.Node.Nodes))
	for i := 0; i < len(e.Node.Nodes); i++ {
		n := e.Node.Nodes[i]
		var t fuse.DirentType
		if n.Dir {
			t = fuse.DT_Dir
		} else {
			t = fuse.DT_File
		}

		nodes[i] = fuse.Dirent{
			Inode: checksum(n.Key),
			Name:  path.Base(n.Key),
			Type:  t,
		}
	}
	return nodes, nil
}

var _ = fs.HandleReadDirAller(&EDDir{})

// Lookup ... Fuse lookup
func (e *EDDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {

	return e.FS.fsNode(ctx, path.Join(e.Node.Key, req.Name))
}

var _ = fs.NodeRequestLookuper(&EDDir{})

// Create ... file creating implementation
func (e *EDDir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if e.FS.ReadOnly {
		return nil, nil, fuse.Errno(syscall.EACCES)
	}
	p := path.Join(e.Key, req.Name)
	v, err := e.FS.KApi.Set(ctx, p, "", &client.SetOptions{
		Dir: false,
	})
	if err != nil {
		return nil, nil, err
	}
	f := &EDFile{Key: p, Node: v.Node, FS: e.FS}

	return f, f, nil
}

var _ = fs.NodeCreater(&EDDir{})

// EDFile ... File entry in etcd
type EDFile struct {
	Key  string
	FS   *EDFS
	Node *client.Node
}

// Attr ... Fuse atter
func (ef *EDFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = 0555
	if ef.FS.ReadOnly {
		attr.Mode = 0444
	}
	if ef.Node != nil {
		attr.Inode = checksum(ef.Node.Key)
		attr.Size = uint64(len(ef.Node.Value))
	}
	return nil
}

var _ fs.Node = (*EDFile)(nil)

// Open ... manage cache ect
func (ef *EDFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	// Disable cache
	resp.Flags |= fuse.OpenDirectIO
	return ef, nil
}

var _ fs.NodeOpener = (*EDFile)(nil)

// ReadAll ... Etcd files are small so read the entire thing
func (ef *EDFile) ReadAll(ctx context.Context) ([]byte, error) {
	if ef.Node == nil {
		return make([]byte, 0), fuse.ENOENT
	}
	return []byte(ef.Node.Value), nil
}

var _ = fs.HandleReadAller(&EDFile{})

// Write ... Implements write fuse handler
func (ef *EDFile) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	// ReadOnly should not be writing
	if ef.FS.ReadOnly {
		return fuse.Errno(syscall.EACCES)
	}

	buf := bytes.NewBufferString("")

	if ef.Node != nil {
		buf.WriteString(ef.Node.Value)
	}
	buf.Truncate(int(req.Offset))
	buf.Write(req.Data)
	resp.Size = buf.Len()
	v, err := ef.FS.KApi.Set(ctx, ef.Key, buf.String(), nil)
	ef.Node = v.Node
	return err
}

var _ = fs.HandleWriter(&EDFile{})

var (
	etcdHosts    []string
	etcdReadOnly bool
)

func init() {
	hosts := os.Getenv("ETC_HOSTS")
	if hosts == "" {
		etcdHosts = []string{"http://127.0.0.1:2379"}
	} else {
		etcdHosts = strings.Split(hosts, ",")
	}

	ro := os.Getenv("ETC_READ_ONLY")
	if ro != "" {
		etcdReadOnly = true
	}

}

// Mount ... Place to mount etcd
func Mount(mountPoint string, ioMap *qos.IOMap) error {

	var err error
	fuseConn, err = fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	defer fuseConn.Close()

	filesys, err := NewEDFS(mountPoint, etcdHosts, etcdReadOnly)
	if err != nil {
		return err
	}

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
