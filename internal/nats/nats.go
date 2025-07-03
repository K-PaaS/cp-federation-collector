package nats

import (
	"github.com/nats-io/nats.go"
	"log"
	"os"
)

type Client struct {
	api        string
	id         string
	password   string
	natsClient *nats.Conn
	jetStream  nats.JetStreamContext
}

func NewClient() *Client {
	nc, err := nats.Connect(os.Getenv("NATS_URL"), nats.UserInfo(os.Getenv("NATS_ID"), os.Getenv("NATS_PASSWORD")))
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
		api:        os.Getenv("NATS_URL"),
		id:         os.Getenv("NATS_ID"),
		password:   os.Getenv("NATS_PASSWORD"),
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
