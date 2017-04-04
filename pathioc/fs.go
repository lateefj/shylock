package pathioc

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/lateefj/shylock/ioc"
	"golang.org/x/net/context"
)

func fileAttr(fi os.FileInfo, a *fuse.Attr) {

	a.Size = uint64(fi.Size())
	a.Mode = fi.Mode()
	a.Mtime = fi.ModTime()
	a.Ctime = fi.ModTime()
	a.Crtime = fi.ModTime()
}

// SFS Shylock File System
type SFS struct {
	Path  string
	IOMap *ioc.IOMap
}

func NewSFS(path string, iocMap *ioc.IOMap) *SFS {
	//TODO: Read from configuration file
	return &SFS{Path: path, IOMap: iocMap}
}

func (sfs *SFS) Root() (fs.Node, error) {
	d := &SDir{SFS: sfs, Path: sfs.Path, IOMap: sfs.IOMap}
	return d, nil
}

type SDir struct {
	SFS   *SFS
	Path  string
	IOMap *ioc.IOMap
}

func (sd *SDir) File() (*os.File, error) {
	f, err := os.Open(sd.Path)
	if err != nil {
		log.Printf("Failed in SDir.File: %s\n", err)
	}
	return f, err
}

func (sd *SDir) IsRoot() bool {
	if sd.Path == sd.SFS.Path {
		return true
	}
	return false
}

var _ fs.Node = (*SDir)(nil)

func (sd *SDir) Attr(ctx context.Context, a *fuse.Attr) error {

	if isDir(sd.Path) {
		// root directory
		a.Mode = os.ModeDir | 0755
		return nil
	}

	f, err := sd.File()
	defer f.Close()
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	fileAttr(fi, a)
	return nil
}

func isDir(path string) bool {
	if path[:len(path)-1] == "/" {
		return true
	}
	return false
}
func (sd *SDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	path := sd.Path + "/" + req.Name
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if isDir(req.Name) {
			return &SDir{Path: path, IOMap: sd.IOMap}, nil
		} else {
			return &SFile{Path: path}, nil
		}
	}
	f, err := os.Open(path)
	defer f.Close()
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return &SDir{Path: path, IOMap: sd.IOMap}, nil
	}
	return &SFile{Path: path, IOMap: sd.IOMap}, nil
}

// Register callback
var _ = fs.NodeRequestLookuper(&SDir{})

func (sd *SDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	var res []fuse.Dirent
	files, err := ioutil.ReadDir(sd.Path)
	if err != nil {
		return res, err
	}
	for _, fileInfo := range files {
		name := fileInfo.Name()
		var de fuse.Dirent
		if name[len(name)-1] == '/' {
			// directory
			name = name[:len(name)-1]
			de.Type = fuse.DT_Dir
		}
		de.Name = name
		res = append(res, de)
	}
	return res, nil
}

// Register callback
var _ = fs.HandleReadDirAller(&SDir{})

func (sd *SDir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {

	path := sd.Path + "/" + req.Name
	fmt.Printf("Removing file %s\n", req.Name)
	if req.Dir {
		return os.RemoveAll(path)
	} else {
		return os.Remove(path)
	}
}

var _ = fs.NodeRemover(&SDir{})

func (sd *SDir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	fmt.Printf("Creating a file %s\n", req.Name)
	path := sd.Path + "/" + req.Name

	f := &SFile{Path: path, IOMap: sd.IOMap}
	return f, f, nil
}

var _ = fs.NodeCreater(&SDir{})

type SFile struct {
	Path  string
	IOMap *ioc.IOMap
	ioc   *ioc.IOC
	file  *os.File
}

var _ fs.Node = (*SFile)(nil)

func (sf *SFile) openFile(flags fuse.OpenFlags) (bool, error) {
	var err error
	exists := true
	if sf.ioc == nil {
		sf.ioc = sf.IOMap.FindPath(sf.Path)
		if sf.ioc == nil {
			fmt.Printf("Failed to find path %s\n", sf.Path)
		}
	}
	if sf.file == nil {

		// Open for writing if write flag is set
		if flags == fuse.OpenWriteOnly {
			sf.file, err = os.OpenFile(sf.Path, os.O_APPEND|os.O_WRONLY, 0600)
		} else { // Just open for reading
			sf.file, err = os.Open(sf.Path)
		}
		if os.IsNotExist(err) {
			exists = false
		}
	}
	return exists, err
}

func (sf *SFile) Attr(ctx context.Context, a *fuse.Attr) error {

	// If file exists then get stat info
	info, err := os.Stat(sf.Path)
	if err != nil {
		return err
	}
	fileAttr(info, a)
	return nil
}

var _ fs.Handle = (*SFile)(nil)

func (sf *SFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	exists, err := sf.openFile(req.Flags)
	if err != nil {
		return nil, err
	}
	// If new file set mods now and return
	if !exists {
		fmt.Println("Having to create file that doesn't exist")
		sf.file, err = os.Create(sf.Path)
	}
	return sf, err
}

var _ = fs.NodeOpener(&SFile{})

func (sfh *SFile) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return sfh.file.Close()
}

var _ fs.HandleReleaser = (*SFile)(nil)

func (sfh *SFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if sfh.ioc != nil {
		stream := make(chan uint64, 1)
		go sfh.ioc.CheckoutRead(uint64(req.Size), stream)
		checkedOut := uint64(0)
		for c := range stream {
			checkedOut += c
		}
	}

	buf := make([]byte, req.Size)
	sfh.file.Seek(req.Offset, 0)
	n, err := sfh.file.Read(buf)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}
	resp.Data = buf[:n]
	return err
}

var _ = fs.HandleReader(&SFile{})

func (sf *SFile) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	if sf.ioc != nil {
		size := len(req.Data)
		stream := make(chan uint64, 1)
		go sf.ioc.CheckoutWrite(uint64(size), stream)
		checkedOut := uint64(0)
		// This checks out all the bytes before wonder if there is a better way to do this?
		for c := range stream {
			checkedOut += c // Probably don't need this we can just checkout until the channel closes
		}
	}

	sf.file.Close()
	f, err := os.OpenFile(sf.Path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	n, err := f.WriteAt(req.Data, req.Offset)
	if err != nil {
		return err
	}
	resp.Size = n
	return err
}

var _ = fs.HandleWriter(&SFile{})

func (sfh *SFile) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return sfh.file.Sync()
}

var _ = fs.HandleFlusher(&SFile{})

var (
	iocDir     string
	configFile string
)

func Mount(mountPoint string, ioMap *ioc.IOMap) error {
	iocDir = os.Getenv("PATHIOC_DIR")
	if iocDir == "" {
		log.Fatalf("PATHIOC_DIR (path to the actual files) is a required environment variable")
	}
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := NewSFS(iocDir, ioMap)

	if err := fs.Serve(c, filesys); err != nil {
		log.Printf("Failed to server because %s", err)

		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Printf("Failed to mount because %s", err)
		return err
	}
	return nil
}
