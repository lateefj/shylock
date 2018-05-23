package shylock

import (
	"errors"
	"fmt"
	"log"

	"github.com/lateefj/shylock/api"
	"github.com/lateefj/shylock/buse"
)

const (
	DOCKER = "DOCKER"
	FUSE   = "FUSE"
)

var (
	mountSystemNotSupported = errors.New("Mount System Not Supported")
	mountedDevices          = make([]api.Device, 0)
	mountedFuse             = make([]*buse.FuseDevice, 0)
)

// mountFuse ... binds together using fuse and whatever the custom interface
// decoupling fuse and the custom systems
func MountFuse(mountPath, fsType string, config []byte) error {
	device, err := api.MountDevice(fsType, mountPath, config)
	if err != nil {
		return err
	}

	// Append to the list of devices
	fuseDevice, err := buse.NewFuseDevice(mountPath, device)
	if err != nil {
		// Try to exit cleanly
		device.Unmount()
		return err
	}
	go func() {
		err = fuseDevice.Mount(mountPath, nil)
		if err != nil {
			log.Panicf("Failed to mount %s\n", mountPath)
		}
		// Start tracking list of devices
		mountedDevices = append(mountedDevices, device)
		mountedFuse = append(mountedFuse, fuseDevice)

	}()
	return nil
}

func Exit() {
	for _, mf := range mountedFuse {
		mf.Unmount()
	}
}

// TODO: Probably would be good to implement someday.
func MountDocker(mountPath, fsType string, config []byte) error {
	return mountSystemNotSupported
}

func Mount(mountPoint, mountType, fsType string, config []byte) error {
	switch mountType {
	case FUSE:
		MountFuse(mountPoint, fsType, config)
	case DOCKER:
		MountDocker(mountPoint, fsType, config)
	default:
		return errors.New(fmt.Sprintf("%s mount system not supported", mountType))
	}
	return nil
}
