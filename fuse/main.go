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
	flag.PrintDefaults()
}
func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	mountPoint := flag.Arg(0)
	log.Printf("Mount point %s\n", mountPoint)
	if err := mount(mountPoint); err != nil {
		log.Fatal(err)
	}
}
