package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lateefj/shylock/kafka"
	"github.com/lateefj/shylock/pathqos"
	"github.com/lateefj/shylock/qos"
)

const (
	progName = "skylock"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s type /mnt/point (types: pathqos|kafka)", progName)
}

func httpInterface(iom *qos.IOMap) {
	port := os.Getenv("HTTP_PORT")
	if port == "" { // If port is not set don't start the http server
		log.Println("Not starting http server")
		return
	}
	go func() {
		qos.Setup(iom)
		log.Printf("Http server on port %s\n", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}()

}

func loadIOCConfig(configFile string) (*qos.IOMap, error) {
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Failed to open configuration file %s with error: %s", configFile, err)
	}
	defer f.Close()
	return qos.LoadIOCConfig(f), nil
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
	iom := qos.NewIOMap()
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
	httpInterface(iom)
	switch mountType {
	case "pathqos":
		if err := pathqos.Mount(mountPoint, iom); err != nil {
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
