package nats

import (
	"federation-metric-api/config"
	"github.com/nats-io/nats.go"
	"log"
)

type Client struct {
	api        string
	id         string
	password   string
	natsClient *nats.Conn
	jetStream  nats.JetStreamContext
}

func NewClient() *Client {
	nc, err := nats.Connect(config.Env.NatsUrl, nats.UserInfo(config.Env.NatsId, config.Env.NatsPassword))
	if err != nil {
		log.Printf("NATS 연결 실패: %v", err)
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		log.Printf("JetStream 연결 실패: %v", err)
		return nil
	}
	return &Client{
		api:        config.Env.NatsUrl,
		id:         config.Env.NatsId,
		password:   config.Env.NatsPassword,
		natsClient: nc,
		jetStream:  js,
	}
}
func (c *Client) CreateKeyValue(bucket string) (nats.KeyValue, error) {
	return c.jetStream.CreateKeyValue(&nats.KeyValueConfig{Bucket: bucket})
}

func (c *Client) KeyValue(bucket string) (nats.KeyValue, error) {
	return c.jetStream.KeyValue(bucket)
}
