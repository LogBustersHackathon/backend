package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LogBustersHackathon/backend/model"

	"github.com/LogBustersHackathon/backend/internal/kafka"
	"github.com/LogBustersHackathon/backend/internal/nats"
	"github.com/LogBustersHackathon/backend/internal/processor"
)

var (
	natsHost          string
	natsPort          int
	natsWebsocketPort int
	natsUsername      string
	natsPassword      string
	natsStream        string
	natsSubject       string
	natsConsumer      string
	kafkaAddress      string
	kafkaTopic        string
	kafkaMechanism    string
	kafkaUsername     string
	kafkaPassword     string
)

func main() {
	flag.StringVar(&natsHost, "nats-host", "0.0.0.0", "NATS server host")
	flag.IntVar(&natsPort, "nats-port", 4222, "NATS server port")
	flag.IntVar(&natsWebsocketPort, "nats-websocket-port", 4223, "NATS server websocket port")
	flag.StringVar(&natsUsername, "nats-username", "", "NATS username")
	flag.StringVar(&natsPassword, "nats-password", "", "NATS password")
	flag.StringVar(&natsStream, "nats-stream", "", "NATS stream name")
	flag.StringVar(&natsSubject, "nats-subject", "alarms", "NATS subject name")
	flag.StringVar(&natsConsumer, "nats-consumer", "server", "NATS consumer name")
	flag.StringVar(&kafkaAddress, "kafka-address", "192.168.1.66:9092", "Kafka server address")
	flag.StringVar(&kafkaTopic, "kafka-topic", "vmqueue", "Kafka topic name")
	flag.StringVar(&kafkaMechanism, "kafka-mechanism", "", "Kafka mechanism. 256 or 512")
	flag.StringVar(&kafkaUsername, "kafka-username", "", "Kafka username")
	flag.StringVar(&kafkaPassword, "kafka-password", "", "Kafka password")

	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	closingChn := make(chan struct{})

	go Application(closingChn)

	<-signalChan

	close(closingChn)

	println("Application will be closed in 10 seconds...")

	time.Sleep(time.Second * 10)
}

func Application(closingChn chan struct{}) {
	if natsSubject == "" {
		fmt.Printf("NATS subject name is required\n")
		os.Exit(1)
	}

	if kafkaTopic == "" {
		fmt.Printf("Kafka topic name is required\n")
		os.Exit(1)
	}

	h, err := nats.StartServer(closingChn, natsHost, natsPort, natsWebsocketPort, natsUsername, natsPassword, natsStream, []string{natsSubject}, natsConsumer)
	if err != nil {
		fmt.Printf("Error starting NATS server: %v\n", err)
		os.Exit(1)
	}

	err = h.CreateConnection()
	if err != nil {
		fmt.Printf("Error creating NATS connection: %v\n", err)
		os.Exit(1)
	}

	if natsStream != "" && natsConsumer != "" {
		err = h.CreateStream()
		if err != nil {
			fmt.Printf("Error creating NATS stream: %v\n", err)
			os.Exit(1)
		}

		err = h.CreateConsumer()
		if err != nil {
			fmt.Printf("Error creating NATS consumer: %v\n", err)
			os.Exit(1)
		}
	}

	processChn := make(chan []model.KafkaAlarm)
	publishChn := make(chan model.AlarmResponse)

	go func() {
		err := kafka.SubscribeTopic(closingChn, processChn, kafkaAddress, kafkaTopic, kafkaMechanism, kafkaUsername, kafkaPassword)
		if err != nil {
			fmt.Printf("Error subscribing to Kafka topic: %v\n", err)
			os.Exit(1)
		}
	}()

	go func() {
		err := processor.Process(closingChn, processChn, publishChn)
		if err != nil {
			fmt.Printf("Error processing messages: %v\n", err)
			os.Exit(1)
		}
	}()

	go func() {
		err := h.Publisher(closingChn, publishChn)
		if err != nil {
			fmt.Printf("Error publishing to NATS: %v\n", err)
			os.Exit(1)
		}
	}()
}
