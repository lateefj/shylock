package api

import (
	"errors"
	"fmt"
)

type Registered struct {
	MountPoint  string
	FSType      string
	Description string
	Config      []byte
}

//type HeaderDeviceBuilder func(mountPoint string, config []byte) HeaderDevice
type DeviceBuilder func(mountPoint string, config []byte) Device
type SimpleDeviceBuilder func(mountPoint string, config []byte) SimpleDevice

type Registrar struct {
	//HeaderDevices map[string]HeaderDeviceBuilder
	Devices       map[string]DeviceBuilder
	SimpleDevices map[string]SimpleDeviceBuilder
}

var reg Registrar

func init() {
	reg = Registrar{SimpleDevices: make(map[string]SimpleDeviceBuilder), Devices: make(map[string]DeviceBuilder)}
	//reg = Registrar{HeaderDevices: make(map[string]HeaderDeviceBuilder), Devices: make(map[string]DeviceBuilder)}
}

/*func RegisterHeaderDevice(fsType string, imp HeaderDeviceBuilder) {
	reg.HeaderDevices[fsType] = imp
}*/
func RegisterDevice(fsType string, imp DeviceBuilder) {
	reg.Devices[fsType] = imp
}

func RegisterSimpleDevice(fsType string, imp SimpleDeviceBuilder) {
	reg.SimpleDevices[fsType] = imp
}

/*func MountHeaderDevice(fsType, mountPoint string, config []byte) (HeaderDevice, error) {
	imp, exists := reg.HeaderDevices[fsType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("No file system type %s", fsType))
	}
	return imp(mountPoint, config), nil
}*/
func MountSimpleDevice(fsType, mountPoint string, config []byte) (SimpleDevice, error) {
	imp, exists := reg.SimpleDevices[fsType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("No file system type %s", fsType))
	}
	return imp(mountPoint, config), nil
}
func MountDevice(fsType, mountPoint string, config []byte) (Device, error) {
	imp, exists := reg.Devices[fsType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("No file system type %s", fsType))
	}
	return imp(mountPoint, config), nil
}
