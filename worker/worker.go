package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func connectConsumer(brokers []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func main() {
	topics := []string{"company_created", "company_deleted", "company_updated"}
	consumer, err := connectConsumer([]string{"kafka:9092"})
	if err != nil {
		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	msgCount := 0
	doneChan := make(chan struct{})

	for _, topic := range topics {
		go func(topic string) {
			partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
			if err != nil {
				return
			}
			defer partitionConsumer.Close()

			for {
				select {
				case msg := <-partitionConsumer.Messages():
					msgCount++
					fmt.Printf("received message count: %d, message: %s, topic: %s\n", msgCount, string(msg.Value), topic)
				case <-sigchan:
					fmt.Println("Interrupt is detected")
					doneChan <- struct{}{}
					return
				}
			}
		}(topic)
	}

	<-doneChan
	fmt.Println("Consumer stopped")
}
