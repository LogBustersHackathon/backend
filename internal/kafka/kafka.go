package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func SubscribeTopic(closerChn chan struct{}, processChn chan struct{}, address string, topic string, mechanism string, username string, password string) (err error) {
	config := kafka.ReaderConfig{}
	config.Brokers = []string{address}
	config.Topic = topic

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
		config.Dialer = &kafka.Dialer{SASLMechanism: m}
	}

	reader := kafka.NewReader(config)
	if reader == nil {
		return fmt.Errorf("can not create reader")
	}

	ctx := context.Background()

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

			// TODO parse message and send to processor channel
			data := struct{}{}

			processChn <- data

			err = reader.CommitMessages(ctx, m)
			if err != nil {
				break taskLoop
			}
		}

		time.Sleep(time.Millisecond * 10)
	}

	return nil
}
