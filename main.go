package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lateefj/shylock/kafka"
	"github.com/lateefj/shylock/pathioc"
)

const (
	progName = "skylock"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s type /mnt/point (types: path|kafka)", progName)
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
	mountType := flag.Arg(0)
	mountPoint := flag.Arg(1)
	log.Printf("Mount type %s point %s\n", mountType, mountPoint)
	switch mountType {
	case "path":
		if err := pathioc.Mount(mountPoint); err != nil {
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
