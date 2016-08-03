package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	progName = "skylock"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s ZIP MOUNTPOINT\n", progName)
	flag.PrintDefaults()
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
	path := flag.Arg(0)
	mountPoint := flag.Arg(1)
	log.Printf("Path %s mount point %s", path, mountPoint)
	if err := mount(path, mountPoint); err != nil {
		log.Fatal(err)
	}
}
