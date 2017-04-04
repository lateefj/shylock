package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lateefj/shylock/ioc"
	"github.com/lateefj/shylock/kafka"
	"github.com/lateefj/shylock/pathioc"
)

const (
	progName = "skylock"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s type /mnt/point (types: pathioc|kafka)", progName)
}

func loadIOCConfig(configFile string) (*ioc.IOMap, error) {
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Failed to open configuration file %s with error: %s", configFile, err)
	}
	defer f.Close()
	return ioc.LoadIOCConfig(f), nil
}

func main() {

	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}
	var err error
	iom := ioc.NewIOMap()
	configFile := os.Getenv("IOC_FILE")
	if configFile != "" {
		iom, err = loadIOCConfig(configFile)
		if err != nil {
			log.Fatalf("Could not load config file %s with error: %s", configFile, err)
		}
	}

	mountType := flag.Arg(0)
	mountPoint := flag.Arg(1)
	log.Printf("Mount type %s point %s\n", mountType, mountPoint)
	switch mountType {
	case "pathioc":
		if err := pathioc.Mount(mountPoint, iom); err != nil {
			log.Fatal(err)
		}
	case "kafka":
		if err := kafka.Mount(mountPoint); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
	}
}
