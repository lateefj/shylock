package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"


	"github.com/lateefj/shylock/etcd"

	"github.com/lateefj/shylock/kafka"
	"github.com/lateefj/shylock/pathqos"
	"github.com/lateefj/shylock/qos"
)

const (
	progName = "skylock"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s type /mnt/point (types: pathqos|kafka|etcd)", progName)

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

type exitFunc func() error

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

	var exf exitFunc

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
	case "etcd":
		if err := etcd.Mount(mountPoint, iom); err != nil {
			log.Fatal(err)
		}
		exf = etcd.Exit
	default:
		usage()
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	case s := <-sigs:
		log.Printf("Caught signal: %v", s)
		closeErrChan := make(chan error)
		if exf != nil {
			go func() {
				closeErrChan <- exf()
			}()
		}
		select {
		case e := <-closeErrChan:
			if e != nil {
				log.Printf("Error closing %s\n", e)
			}
		case <-time.After(1 * time.Second):
			log.Fatalf("FAILED Waiting for fuse mount to close\n")
			os.Exit(1)
		}
	}
}
