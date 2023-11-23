package nats

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/LogBustersHackathon/backend/model"
	"github.com/LogBustersHackathon/backend/utils"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Helper struct {
	s             *server.Server
	c             *nats.Conn
	jsc           nats.JetStreamContext
	host          string
	certFile      string
	keyFile       string
	username      string
	password      string
	stream        string
	consumer      string
	subjects      []string
	natsPort      int
	websocketPort int
}

func StartServer(closingChn chan struct{}, host string, port int, websocketPort int, username string, password string,
	stream string, subjects []string, consumer string, certFile string, keyFile string) (*Helper, error) {
	h := &Helper{}
	h.host = host
	h.natsPort = port
	h.websocketPort = websocketPort
	h.certFile = certFile
	h.keyFile = keyFile
	h.username = username
	h.password = password
	h.stream = stream
	h.consumer = consumer
	h.subjects = subjects

	opts := server.Options{}
	opts.Host = h.host
	opts.Port = h.natsPort
	opts.Websocket = server.WebsocketOpts{}
	opts.Websocket.Host = h.host
	opts.Websocket.Port = h.websocketPort

	if h.username != "" && h.password != "" {
		opts.Username = h.username
		opts.Password = h.password
		opts.Websocket.Username = h.username
		opts.Websocket.Password = h.password
	}

	if h.stream != "" {
		opts.JetStream = true
	}

	if h.certFile != "" && h.keyFile != "" {
		cert, err := tls.LoadX509KeyPair(h.certFile, h.keyFile)
		if err != nil {
			return nil, err
		}
		opts.Websocket.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	} else {
		opts.Websocket.NoTLS = true
	}

	s, err := server.NewServer(&opts)
	if err != nil {
		return nil, err
	}

	h.s = s

	go h.s.Start()

	go func() {
		<-closingChn
		h.s.Shutdown()
	}()

	fmt.Printf("NATS server is running on %s:%d\n", h.host, h.natsPort)

	return h, nil
}

func (h *Helper) CreateConnection() error {
	var opts []nats.Option

	if h.username != "" && h.password != "" {
		opts = append(opts, nats.UserInfo(h.username, h.password))
	}

	if h.certFile != "" && h.keyFile != "" {
		cert, err := tls.LoadX509KeyPair(h.certFile, h.keyFile)
		if err != nil {
			return err
		}
		opts = append(opts, nats.Secure(&tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}))
	}

	conn, err := nats.Connect(fmt.Sprintf("ws://%s:%d", h.host, h.websocketPort), opts...)
	if err != nil {
		return err
	}

	h.c = conn

	return nil
}

func (h *Helper) CreateStream() error {
	if h.s == nil || h.c == nil {
		return errors.New("server is not started or connection is not initialized")
	}

	if !h.s.ReadyForConnections(time.Second * 10) {
		return errors.New("server is not ready for connections")
	}

	jsc, err := h.c.JetStream()
	if err != nil {
		return err
	}

	h.jsc = jsc

	streamInfo, err := h.jsc.StreamInfo(h.stream)
	if err != nil && err != nats.ErrStreamNotFound {
		return err
	}

	if streamInfo != nil {
		return nil
	}

	_, err = h.jsc.AddStream(&nats.StreamConfig{
		Name:     h.stream,
		Subjects: h.subjects,
	})

	return err
}

func (h *Helper) CreateConsumer() error {
	if h.jsc == nil {
		return errors.New("jetstream context is not initialized")
	}

	consumerInfo, err := h.jsc.ConsumerInfo(h.stream, h.consumer)
	if err != nil && err != nats.ErrConsumerNotFound {
		return err
	}

	if consumerInfo != nil {
		return nil
	}

	_, err = h.jsc.AddConsumer(h.stream, &nats.ConsumerConfig{
		Durable: h.consumer,
	})

	return err
}

func (h *Helper) Publisher(closingChn chan struct{}, publishChn chan model.AlarmResponse) error {
	if h.s == nil {
		return errors.New("server is not started")
	}

	if len(h.subjects) == 0 {
		return errors.New("subjects are not defined")
	}

	subject := h.subjects[0]

	fmt.Printf("Publishing to subject %s\n", subject)

taskLoop:
	for {
		select {
		case <-closingChn:
			break taskLoop
		case alarm := <-publishChn:
			data, err := utils.ConvertToJSON(alarm)
			if err != nil {
				continue taskLoop
			}

			if h.jsc != nil {
				_, err = h.jsc.Publish(subject, data)
				if err != nil {
					return err
				}

				fmt.Printf("Published new alarm...\n")

				continue taskLoop
			}

			err = h.c.Publish(subject, data)
			if err != nil {
				return err
			}

			fmt.Printf("Published new alarm...\n")
		}
	}

	return nil
}
