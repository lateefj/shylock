package api

import (
	"errors"
	"fmt"
)

type Registered struct {
	MountPoint  string
	FSType      string
	Description string
	Config      map[string]string
}

type HeaderDeviceBuilder func(mountPoint string, config map[string]string) HeaderDevice
type DeviceBuilder func(mountPoint string, config map[string]string) Device

type Registrar struct {
	HeaderDevices map[string]HeaderDeviceBuilder
	Devices       map[string]DeviceBuilder
}

var reg Registrar

func init() {
	reg = Registrar{HeaderDevices: make(map[string]HeaderDeviceBuilder), Devices: make(map[string]DeviceBuilder)}
}

func RegisterHeaderDevice(fsType string, imp HeaderDeviceBuilder) {
	reg.HeaderDevices[fsType] = imp
}
func RegisterDevice(fsType string, imp DeviceBuilder) {
	reg.Devices[fsType] = imp
}

func MountHeaderDevice(fsType, mountPoint string, config map[string]string) (HeaderDevice, error) {
	imp, exists := reg.HeaderDevices[fsType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("No file system type %s", fsType))
	}
	return imp(mountPoint, config), nil
}
