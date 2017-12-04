// Bazil Fuse adapter interface
package buse

import (
	"hash/crc64"
	"log"
	"os"
	"path"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/lateefj/shylock/api"
	"github.com/lateefj/shylock/qos"
	"golang.org/x/net/context"
)

var (
	crc = crc64.MakeTable(crc64.ECMA)
)

// Get the checksum for a key
func checksum(key string) uint64 {
	return crc64.Checksum([]byte(key), crc)
}

// Fuse ... Fuse wrapper
type FuseDevice struct {
	MountPoint string
	api.Device
	fuseConn *fuse.Conn
}

// NewFuse ... Create a new Fuse instance
func NewFuseDevice(mountPoint string, device api.Device) (*FuseDevice, error) {
	return &FuseDevice{MountPoint: mountPoint, Device: device}, nil
}

// fsNode ... Looks up the in device
func (fd *FuseDevice) fsNode(ctx context.Context, key string) (fs.Node, error) {

	if key[:len(key)-1] == "/" {
		return &FDDir{FS: fd, Key: key}, nil
	}
	f, err := fd.Device.Open(key)
	if err != nil {
		return nil, err
	}
	return &FDFile{FS: fd, Key: key, File: f}, nil
}

// Root ... Required for fuse system
func (fd *FuseDevice) Root() (fs.Node, error) {
	return fd.fsNode(context.Background(), fd.MountPoint)
}

// Mount ... Connect to fuse
func (fd *FuseDevice) Mount(mountPoint string, ioMap *qos.IOMap) error {

	var err error
	fd.fuseConn, err = fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	defer fd.fuseConn.Close()

	err = fs.Serve(fd.fuseConn, fd)
	if err != nil {
		return err
	}
	// check if the mount process has an error to report
	<-fd.fuseConn.Ready
	if err := fd.fuseConn.MountError; err != nil {
		log.Printf("Failed to mount because %s", err)
	}
	return err
}

// Exit ... Exit hook
func (fd *FuseDevice) Exit() error {
	if fd.fuseConn != nil {
		return fd.fuseConn.Close()
	}
	return nil
}

// FDDir ... Directory entry which is not really a thing
type FDDir struct {
	Key string
	FS  *FuseDevice
}

// Attr ... Required for fuse
func (fdd *FDDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = checksum(fdd.Key)
	attr.Mode = os.ModeDir | 0555
	return nil
}

var _ fs.Node = (*FDDir)(nil)

// ReadDirAll ... Get everything in a directory
func (fdd *FDDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	// Refresh the directory listing
	fileNames, err := fdd.FS.Device.List(fdd.Key)
	if err != nil {
		return make([]fuse.Dirent, 0), err
	}
	nodes := make([]fuse.Dirent, len(fileNames))
	for i := 0; i < len(fileNames); i++ {
		n := fileNames[i]
		var t fuse.DirentType
		if n[:len(n)-1] == "/" {
			t = fuse.DT_Dir
		} else {
			t = fuse.DT_File
		}

		nodes[i] = fuse.Dirent{
			Inode: checksum(n),
			Name:  path.Base(n),
			Type:  t,
		}
	}
	return nodes, nil
}

var _ = fs.HandleReadDirAller(&FDDir{})

// Lookup ... Fuse lookup
func (fdd *FDDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {

	return fdd.FS.fsNode(ctx, path.Join(fdd.Key, req.Name))
}

var _ = fs.NodeRequestLookuper(&FDDir{})

// Create ... file creating implementation
func (fdd *FDDir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	p := path.Join(fdd.Key, req.Name)
	f, err := fdd.FS.Device.Open(p)
	if err != nil {
		return nil, nil, err
	}
	err = f.Write([]byte(""))
	if err != nil {
		return nil, nil, err
	}

	fdf := &FDFile{Key: p, File: f, FS: fdd.FS}

	return fdf, fdf, nil
}

var _ = fs.NodeCreater(&FDDir{})

// FDFile ... File entry in Device
type FDFile struct {
	File api.File
	Key  string
	FS   *FuseDevice
}

// Attr ... Fuse atter
func (fdf *FDFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = 0555
	if fdf.File != nil {
		attr.Inode = checksum(fdf.Key)
		// TODO: Maybe be worth setting the size
		//attr.Size = uint64(fdf.File.Size())
	}
	return nil
}

var _ fs.Node = (*FDFile)(nil)

// Open ... file should already be open
func (fdf *FDFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	// Disable cache
	resp.Flags |= fuse.OpenDirectIO
	return fdf, nil
}

var _ fs.NodeOpener = (*FDFile)(nil)

// Flush ... Basically closes the file
func (fdf *FDFile) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return fdf.File.Close()
}

var _ = fs.HandleFlusher(&FDFile{})

// ReadAll ... Read entire file since it is a simple interface
func (fdf *FDFile) ReadAll(ctx context.Context) ([]byte, error) {
	if fdf.File == nil {
		return make([]byte, 0), fuse.ENOENT
	}
	bits, err := fdf.File.Read()
	if err != nil {
		return make([]byte, 0), err
	}
	return bits, nil
}

var _ = fs.HandleReadAller(&FDFile{})

// Write ... Implements write fuse handler
func (fdf *FDFile) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {

	err := fdf.File.Write(req.Data)
	if err != nil {
		return err
	}

	resp.Size = len(req.Data)
	return err
}

var _ = fs.HandleWriter(&FDFile{})
