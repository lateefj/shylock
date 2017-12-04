package api

// File ... Simplified file interface
type File interface {
	Read() (body []byte, err error)
	Write(body []byte) error
	Close() error
}

// Device ... Simplified interface for interacting with a device
type Device interface {
	Mount(config map[string]string) error
	Unmount() error
	List(path string) ([]string, error)
	Open(path string) (File, error)
}

/* Devices with headers */

// HeaderFile ... Support for a file with header
type HeaderFile interface {
	Read() (header, body []byte, err error)
	Write(offset int, header, body []byte) (int, error)
}

// HeaderDevice ... Devices that have a header as part of reading / writing
type HeaderDevice interface {
	Mount(config map[string]string) error
	Unmount() error
	List(path string) ([]string, error)
	Open(path string) (HeaderFile, error)
}

// LargeFile ... This interface support reading / writing large files
type LargeFile interface {
	Read(offest, size int) (body []byte, err error)
	Write(offset int, body []byte) (int, error)
	Close() error
}

// Device ... Support Large file
type LargeFileDevice interface {
	Device
	OpenLarge(path string) (LargeFile, error)
}
