package kafka

import (
	"log"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	Topics  []string `json:"topics"`
}

type ClusterConfig struct {
	KafkaConfig
	Name string `json:"name"`
}

func Topics(brokers []string) ([]string, error) {
	topics := make([]string, 0)
	c, err := sarama.NewClient(brokers, sarama.NewConfig())

	if err != nil {
		log.Printf("ERROR: Failed to get topics %s\n", err)
		return topics, err
	}
	return c.Topics()
}

func NewConsumer(cc ClusterConfig) (*cluster.Consumer, error) {

	// init (custom) config, enable errors and notifications
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true

	consumer, err := cluster.NewConsumer(cc.Brokers, cc.Name, cc.Topics, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

type Producer struct {
	client sarama.AsyncProducer
	topic  string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {

	client, err := sarama.NewAsyncProducer(brokers, nil)
	return &Producer{client, topic}, err
}

func (p *Producer) Send(bits []byte) {
	p.client.Input() <- &sarama.ProducerMessage{Topic: p.topic, Key: nil, Value: sarama.ByteEncoder(bits)}
}
func (p *Producer) KeySend(key string, bits []byte) {
	p.client.Input() <- &sarama.ProducerMessage{Topic: p.topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(bits)}
}
