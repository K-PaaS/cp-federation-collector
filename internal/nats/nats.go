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

var natsConnect = func(url, user, pass string) (*nats.Conn, error) {
	return nats.Connect(url, nats.UserInfo(user, pass))
}

var jetStreamFromConn = func(nc *nats.Conn) (nats.JetStreamContext, error) {
	return nc.JetStream()
}

func NewClient() *Client {
	nc, err := natsConnect(config.Env.NatsUrl, config.Env.NatsId, config.Env.NatsPassword)
	if err != nil {
		log.Printf("NATS 연결 실패: %v", err)
		return nil
	}
	js, err := jetStreamFromConn(nc)
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
