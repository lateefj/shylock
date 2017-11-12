package main

import (
	"errors"
	"fmt"
)

const (
	DOCKER = "DOCKER"
	FUSE   = "FUSE"
)

func mountFuse(mountPath, fsType, config map[string]string) error {
	return nil
}

func mountDocker(mountPath, fsType, config map[string]string) error {
	return nil
}

func mount(mountPath, mountType, fsType string, config map[string]string) error {
	switch mountType {
	case FUSE:
		mountFuse(mountPoint, fsType, config)
	case DOCKER:
		mountDocker(mountPoint, fsType, config)
	default:
		return errors.New(fmt.Sprintf("%s mount system not supported", mountType))
	}
	return nil
}
