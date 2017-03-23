package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/lateefj/shylock/kafka"
)

const (
	progName = "skylock"
)

var (
	kafkaBrokers []string
	kafkaTopic   string
)

func init() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		kafkaBrokers = []string{"127.0.0.1:9092"}
	} else {
		kafkaBrokers = strings.Split(brokers, ",")
	}

	kafkaTopic = os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "my_topic"
	}
}

func testProducer() {

	producer, err := kafka.NewProducer(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Printf("ERROR: Failed to create a producer: %s", err)
		return
	}
	for i := 0; ; i++ {
		fmt.Printf("Sending message: %d\n", i)
		producer.KeySend("foo", []byte(fmt.Sprintf("Test message %d\n", i)))
		time.Sleep(5 * time.Second)
	}
}

func mountKafka() {

	topics := []string{kafkaTopic}
	kc := kafka.KafkaConfig{Brokers: kafkaBrokers, Topics: topics}
	c := kafka.ClusterConfig{KafkaConfig: kc, Name: "test"}
	consumer, err := kafka.NewConsumer(c)
	if err != nil {
		log.Printf("ERROR: Failed to connect to kafka %s\n", err)
		return
	}

	for {
		select {
		case m := <-consumer.Messages():
			fmt.Printf("Message from topic %s Key %s and body is %s\n", m.Topic, string(m.Key), string(m.Value))
		case e := <-consumer.Errors():
			log.Printf("ERROR: From topic %s\n", e)
		}
	}

}

func mount(mountPoint string) error {
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := kafka.NewKFS(mountPoint, kafkaBrokers)

	if err := fs.Serve(c, filesys); err != nil {
		log.Printf("ERROR: Failed to server because %s\n", err)

		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Printf("ERROR: Failed to mount because %s\n", err)
		return err
	}
	return nil

}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: shylock /path/to/thing \n", "shylock")
	flag.PrintDefaults()
}
func main() {
	fmt.Printf("Kafka topic %s and brokers %s\n", kafkaTopic, kafkaBrokers)
	go testProducer()
	//mountKafka()

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
	mount(mountPoint)
}
