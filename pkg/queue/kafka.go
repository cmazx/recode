package queue

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"os"
	"strconv"
	"time"
)

type Serializable interface {
	Serialize() []byte
}
type Queue struct {
	producer        *kafka.Producer
	consumer        *kafka.Consumer
	JobTopicName    string
	consumerTimeout time.Duration
}

func NewQueue() (*Queue, error) {
	broker := os.Getenv("KAFKA_URL")
	topicName := os.Getenv("KAFKA_JOB_TOPIC")

	consumerTimeout, err := strconv.Atoi(os.Getenv("KAFKA_CONSUMER_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	if os.Getenv("KAFKA_JOB_CREATE_TOPIC") == "1" {
		adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
		if err != nil {
			fmt.Printf("Failed to create Admin client: %s\n", err)
			os.Exit(1)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		results, err := adminClient.CreateTopics(
			ctx,
			[]kafka.TopicSpecification{{
				Topic:             topicName,
				NumPartitions:     1,
				ReplicationFactor: 1}},
			kafka.SetAdminOperationTimeout(60*time.Second))
		if err != nil {
			fmt.Printf("Failed to create topic: %v\n", err)
			os.Exit(1)
		}

		for _, result := range results {
			fmt.Printf("%s\n", result)
		}
		adminClient.Close()
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		panic(err)
	}

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"group.id":           "myGroup",
		"session.timeout.ms": 6000,
		"auto.offset.reset":  "latest",
	})

	if err != nil {
		panic(err)
	}

	err = consumer.SubscribeTopics([]string{topicName}, nil)
	if err != nil {
		return nil, err
	}

	return &Queue{
		producer,
		consumer,
		topicName,
		time.Duration(consumerTimeout),
	}, nil
}

func (q *Queue) StopService() {
	q.producer.Flush(15 * 1000)
	q.producer.Close()
}

func (q *Queue) Enqueue(topic string, payload Serializable) error {
	return q.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload.Serialize(),
	}, nil)

}

func (q *Queue) Consume(f func(payload []byte) error) error {
	message, err := q.consumer.ReadMessage(q.consumerTimeout * time.Second)

	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrTimedOut {
			log.Println(".")
			return nil
		}
		fmt.Println("Consume error")
		fmt.Println(err)
		return err
	}

	fmt.Println("Message consumed")

	return f(message.Value)
}
