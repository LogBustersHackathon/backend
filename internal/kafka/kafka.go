package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/LogBustersHackathon/backend/model"
	"github.com/LogBustersHackathon/backend/utils"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func SubscribeTopic(closerChn chan struct{}, processChn chan []model.KafkaAlarm, address string, topic string, mechanism string, username string, password string) (err error) {
	config := kafka.ReaderConfig{}
	config.Brokers = []string{address}
	config.Topic = topic

	config.Dialer = &kafka.Dialer{}

	if username != "" && password != "" {
		var algorithm scram.Algorithm
		switch mechanism {
		case "256":
			algorithm = scram.SHA256
		case "512":
			algorithm = scram.SHA512
		default:
			return fmt.Errorf("algorithm must be a valid options")
		}
		m, err := scram.Mechanism(algorithm, username, password)
		if err != nil {
			return err
		}
		config.Dialer.SASLMechanism = m
	}

	reader := kafka.NewReader(config)
	if reader == nil {
		return fmt.Errorf("can not create reader")
	}

	ctx := context.Background()

	fmt.Printf("Subscribing to topic %s\n", topic)

taskLoop:
	for {
		select {
		case <-closerChn:
			break taskLoop
		default:
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				break taskLoop
			}

			value := []byte(nil)
			value = append(value, '[')
			value = append(value, m.Value...)
			value = append(value, ']')

			data, err := utils.ConvertFromJSON[[]model.KafkaAlarm](value)
			if err != nil {
				continue taskLoop
			}

			processChn <- data

			err = reader.CommitMessages(ctx, m)
			if err != nil {
				continue taskLoop
			}
		}

		time.Sleep(time.Millisecond * 10)
	}

	return nil
}
