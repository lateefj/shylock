package api

// File ... Standard file interface
type File interface {
	Read() (body *[]byte, err error)
	Write(body *[]byte) error
}

// Device ... Simple interface for interacting with a device
type Device interface {
	Mount(config map[string]string) error
	Unmount() error
	List(path string) ([]string, error)
	Open(path string) (File, error)
}

/* Devices with headers */

// HeaderFile ... Support for a file with header
type HeaderFile interface {
	Read() (header, body *[]byte, err error)
	Write(header, body *[]byte) error
}

// HeaderDevice ... Devices that have a header as part of reading / writing
type HeaderDevice interface {
	Mount(config map[string]string) error
	Unmount() error
	List(path string) ([]string, error)
	Open(path string) (HeaderFile, error)
}
