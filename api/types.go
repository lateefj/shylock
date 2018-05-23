package api

// StdDevice ... Shared device functions
type StdDevice interface {
	Mount(config []byte) error
	Unmount() error
	List(path string) ([]string, error)
	Remove(path string) error
}

// File ... Simplified file interface
type SimpleFile interface {
	Read() (body []byte, err error)
	Write(body []byte) error
	Close() error
}

// SimpleDevice ... Simplified interface for interacting with a device
type SimpleDevice interface {
	StdDevice
	Open(path string) (SimpleFile, error)
}

// File ... This interface support reading / writing large files
type File interface {
	Read(offest, size int) (body []byte, err error)
	Write(offset int, body []byte) (int, error)
	Close() error
}

// Device ... Normal file interface
type Device interface {
	StdDevice
	OpenLarge(path string) (File, error)
}

/* Devices with headers concept */
// HeaderFile ... Support for a file with header
type HeaderFile interface {
	Read() (header, body []byte, err error)
	Write(offset int, header, body []byte) (int, error)
}

// HeaderDevice ... Devices that have a header as part of reading / writing
type HeaderDevice interface {
	Mount(config []byte) error
	Unmount() error
	List(path string) ([]string, error)
	Open(path string) (HeaderFile, error)
}
