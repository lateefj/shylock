package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lateefj/shylock/kafka"
)

func testProducer() {

	brokers := []string{"127.0.0.1:9092"}
	producer, err := kafka.NewProducer(brokers, "my_topic")
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
	brokers := []string{"127.0.0.1:9092"}
	topics := []string{"my_topic", "other_topic"}

	kc := kafka.KafkaConfig{Brokers: brokers, Topics: topics}
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
	go testProducer()
	mountKafka()
}
