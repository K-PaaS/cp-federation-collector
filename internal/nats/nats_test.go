package nats

import (
	"errors"
	"testing"

	"federation-metric-api/config"
	outnats "github.com/nats-io/nats.go"
)

type fakeKV struct {
	outnats.KeyValue
	lastKey   string
	putCalled bool
}

func (f *fakeKV) CreateKeyValue(bucket *outnats.KeyValueConfig) (outnats.KeyValue, error) {
	f.lastKey = bucket.Bucket
	f.putCalled = true
	return f, nil
}

type fakeJS struct {
	outnats.JetStreamContext
	kv         outnats.KeyValue
	lastBucket string
}

func (f *fakeJS) CreateKeyValue(cfg *outnats.KeyValueConfig) (outnats.KeyValue, error) {
	f.lastBucket = cfg.Bucket
	return f.kv, nil
}

func (f *fakeJS) KeyValue(bucket string) (outnats.KeyValue, error) {
	f.lastBucket = bucket
	return f.kv, nil
}

type fakeJetStream struct {
	outnats.JetStreamContext
}

func TestNewClient_Success(t *testing.T) {
	oldConnect := natsConnect
	oldJS := jetStreamFromConn
	defer func() {
		natsConnect = oldConnect
		jetStreamFromConn = oldJS
	}()

	config.Env.NatsUrl = "nats://test:4222"
	config.Env.NatsId = "user"
	config.Env.NatsPassword = "pass"

	natsConnect = func(url, user, pass string) (*outnats.Conn, error) {
		if url != "nats://test:4222" || user != "user" || pass != "pass" {
			t.Fatalf("unexpected connect args: %q %q %q", url, user, pass)
		}
		return &outnats.Conn{}, nil
	}
	js := &fakeJetStream{}
	jetStreamFromConn = func(nc *outnats.Conn) (outnats.JetStreamContext, error) {
		if nc == nil {
			t.Fatalf("expected non-nil conn")
		}
		return js, nil
	}

	c := NewClient()
	if c == nil {
		t.Fatalf("expected client, got nil")
	}
	if c.api != "nats://test:4222" || c.id != "user" || c.password != "pass" {
		t.Fatalf("unexpected client fields: %+v", c)
	}
	if c.jetStream != js {
		t.Fatalf("client.jetStream was not set from hook")
	}
}

func TestNewClient_ConnectErrorReturnsNil(t *testing.T) {
	oldConnect := natsConnect
	oldJS := jetStreamFromConn
	defer func() {
		natsConnect = oldConnect
		jetStreamFromConn = oldJS
	}()

	natsConnect = func(url, user, pass string) (*outnats.Conn, error) {
		return nil, errors.New("connect-failed")
	}

	c := NewClient()
	if c != nil {
		t.Fatalf("expected nil client on connect error, got %+v", c)
	}
}

func TestNewClient_JetStreamErrorReturnsNil(t *testing.T) {
	oldConnect := natsConnect
	oldJS := jetStreamFromConn
	defer func() {
		natsConnect = oldConnect
		jetStreamFromConn = oldJS
	}()

	natsConnect = func(url, user, pass string) (*outnats.Conn, error) {
		return &outnats.Conn{}, nil
	}

	jetStreamFromConn = func(nc *outnats.Conn) (outnats.JetStreamContext, error) {
		return nil, errors.New("js-failed")
	}

	c := NewClient()
	if c != nil {
		t.Fatalf("expected nil client on jetstream error, got %+v", c)
	}
}

func TestClient_CreateKeyValue_DelegatesToJetStream(t *testing.T) {
	kv := &fakeKV{}
	js := &fakeJS{kv: kv}
	c := &Client{jetStream: js}

	_, err := c.CreateKeyValue("metrics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if js.lastBucket != "metrics" {
		t.Fatalf("expected bucket 'metrics', got %q", js.lastBucket)
	}
}

func TestClient_KeyValue_DelegatesToJetStream(t *testing.T) {
	kv := &fakeKV{}
	js := &fakeJS{kv: kv}
	c := &Client{jetStream: js}

	_, err := c.KeyValue("metrics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if js.lastBucket != "metrics" {
		t.Fatalf("expected bucket 'metrics', got %q", js.lastBucket)
	}
}
