package main

import (
	"io"
	"io/ioutil"
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
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
	Path string
	IOC  *IOC
}

func NewSFS(path string) *SFS {
	ioc := NewIOC(10*time.Second, uint64(1), uint64(1))
	go ioc.Start()
	return &SFS{Path: path, IOC: ioc}
}

func (sfs *SFS) Root() (fs.Node, error) {
	d := &SDir{SFS: sfs, IOC: sfs.IOC, Path: sfs.Path}
	return d, nil

}

type SDir struct {
	SFS  *SFS
	IOC  *IOC
	Path string
}

func (sd *SDir) File() (*os.File, error) {
	f, err := os.Open(sd.Path)
	if err != nil {
		log.Errorf("Failed in SDir.File: %s", err)
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

	if sd.IsRoot() {
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
			return &SDir{Path: path, IOC: sd.IOC}, nil
		} else {
			return &SFile{Path: path, ioc: sd.IOC}, nil
		}
	}
	f, err := os.Open(path)
	defer f.Close()
	if err != nil && !os.IsExist(err) {
		log.Errorf("SDir lookup: File path %s does not exist :( error: %s", path, err)
		return nil, err
	}

	if err != nil {
		log.Errorf("SDir Lookup file %s is nil error: %s", path, err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return &SDir{Path: path, IOC: sd.IOC}, nil
	}
	return &SFile{Path: path, ioc: sd.IOC}, nil
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
	log.Infof("Removing file %s", req.Name)
	if req.Dir {
		return os.RemoveAll(path)
	} else {
		return os.Remove(path)
	}
}

var _ = fs.NodeRemover(&SDir{})

func (sd *SDir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	path := sd.Path + "/" + req.Name

	f := &SFile{Path: path, ioc: sd.IOC}
	return f, f, nil
}

var _ = fs.NodeCreater(&SDir{})

type SFile struct {
	Path string
	ioc  *IOC
	file *os.File
}

var _ fs.Node = (*SFile)(nil)

func (sf *SFile) Attr(ctx context.Context, a *fuse.Attr) error {
	file, err := os.Open(sf.Path)
	defer file.Close()
	if os.IsNotExist(err) {
		t := time.Now()
		a.Mtime = t
		a.Ctime = t
		a.Crtime = t
		return nil
	}
	if err != nil {
		log.Errorf("Failed to open file for attr %s", err)
		return err
	}
	info, err := file.Stat()
	if err != nil {
		log.Errorf("Stat failed for file %s open attr %s", sf.Path, err)
		return err
	}
	fileAttr(info, a)
	return nil
}

var _ = fs.NodeOpener(&SFile{})

func (sf *SFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {

	file, err := os.Open(sf.Path)
	if os.IsNotExist(err) {
		file, err = os.Create(sf.Path)
	}
	if err != nil {
		log.Errorf("Failed to open SFil.Open %s", err)
		return nil, err
	}

	sf.file = file
	return sf, nil
}

var _ fs.Handle = (*SFile)(nil)

var _ fs.HandleReleaser = (*SFile)(nil)

func (sfh *SFile) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return sfh.file.Close()
}

var _ = fs.HandleReader(&SFile{})

func (sfh *SFile) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Debugf("Copying bytes: %d", req.Size)
	stream := make(chan uint64, 1)
	go sfh.ioc.CheckoutRead(uint64(req.Size), stream)
	checkedOut := uint64(0)
	for c := range stream {
		log.Debugf("Got %d more bytes", c)
		checkedOut += c
	}

	buf := make([]byte, req.Size)
	sfh.file.Seek(req.Offset, 0)
	n, err := sfh.file.Read(buf)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		log.Errorf("IO Error %S", err)
		err = nil
	}
	resp.Data = buf[:n]
	log.Debugf("Ok copied %d", len(resp.Data))
	return err
}

var _ = fs.HandleWriter(&SFile{})

func (sf *SFile) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	size := len(req.Data)
	log.Debugf("Writing  bytes: %d", size)
	stream := make(chan uint64, 1)
	go sf.ioc.CheckoutWrite(uint64(size), stream)
	checkedOut := uint64(0)
	for c := range stream {
		log.Debugf("Got %d more bytes", c)
		checkedOut += c
	}

	//sf.file.Seek(req.Offset, 0)
	n, err := sf.file.WriteAt(req.Data, req.Offset)
	if err != nil {
		log.Errorf("Failed to write to file %s with offset %d", err, req.Offset)
	}
	resp.Size = n
	log.Debugf("Wrote %d bytes", n)
	return err
}

var _ = fs.HandleFlusher(&SFile{})

func (sfh *SFile) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return sfh.file.Sync()
}

func mount(path, mountpoint string) error {
	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := NewSFS(path)

	if err := fs.Serve(c, filesys); err != nil {
		log.Errorf("Failed to server because %s", err)

		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Errorf("Faield to mount because %s", err)
		return err
	}
	return nil

}
