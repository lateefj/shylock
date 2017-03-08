package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lateefj/shylock/kafka"
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
		log.Printf("ERROR: Failed to create a producer")
		return
	}
	for i := 0; ; i++ {
		fmt.Printf("Sending message: %d\n", i)
		producer.KeySend("foo", []byte(fmt.Sprintf("Test message %d", i)))
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

func main() {
	fmt.Printf("Kafka topic %s and brokers %s\n", kafkaTopic, kafkaBrokers)
	go testProducer()
	mountKafka()
}
