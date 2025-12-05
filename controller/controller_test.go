package controller

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"federation-metric-api/internal/karmada"
	"federation-metric-api/model"
	outnats "github.com/nats-io/nats.go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type fakeKarm struct {
	clusters []karmada.MemberCluster
}

func (f *fakeKarm) GetMemberClusters(ctx context.Context) ([]karmada.MemberCluster, error) {
	return f.clusters, nil
}

type fakeKV struct {
	outnats.KeyValue
	puts [][]byte
}

func (f *fakeKV) Put(key string, val []byte) (uint64, error) {
	f.puts = append(f.puts, val)
	return 1, nil
}

type fakeNats struct {
	kv outnats.KeyValue
}

func (f *fakeNats) CreateKeyValue(bucket string) (outnats.KeyValue, error) {
	return f.kv, nil
}

func (f *fakeNats) KeyValue(bucket string) (outnats.KeyValue, error) {
	return f.kv, nil
}

func fakeGetClusterInfos() ([]model.ClusterCredential, error) {
	return []model.ClusterCredential{}, nil
}

func TestRepeatMetric_UsesNatsKVPut(t *testing.T) {
	oldRepeat := repeatTime
	repeatTime = 1
	defer func() { repeatTime = oldRepeat }()

	fakeStore := &fakeKV{}

	oldKarm := NewKarmadaClient
	oldNats := NewNatsClient
	oldGet := GetClusterInfos

	NewKarmadaClient = func() KarmadaClient { return &fakeKarm{clusters: []karmada.MemberCluster{}} }
	NewNatsClient = func() NatsClient { return &fakeNats{kv: fakeStore} }
	GetClusterInfos = fakeGetClusterInfos

	defer func() {
		NewKarmadaClient = oldKarm
		NewNatsClient = oldNats
		GetClusterInfos = oldGet
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go RepeatMetric(ctx)

	time.Sleep(1500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)

	if len(fakeStore.puts) == 0 {
		t.Fatalf("expected at least one KV Put call")
	}

	var ms model.MetricStatus
	if err := json.Unmarshal(fakeStore.puts[0], &ms); err != nil {
		t.Fatalf("invalid MetricStatus JSON stored in KV: %v", err)
	}
}

func TestRepeatMetric_FillsHostAndMemberMetrics(t *testing.T) {
	oldRepeat := repeatTime
	repeatTime = 1
	defer func() { repeatTime = oldRepeat }()

	fakeStore := &fakeKV{}

	oldKarm := NewKarmadaClient
	oldNats := NewNatsClient
	oldGetClusters := GetClusterInfos
	oldKube := NewKubeClient
	oldCollect := CollectMetricFunc
	oldCollectReq := CollectRequestMetricFunc
	oldHealth := NodeHealthCheckFunc
	oldSummary := NodeSummaryFunc

	defer func() {
		NewKarmadaClient = oldKarm
		NewNatsClient = oldNats
		GetClusterInfos = oldGetClusters
		NewKubeClient = oldKube
		CollectMetricFunc = oldCollect
		CollectRequestMetricFunc = oldCollectReq
		NodeHealthCheckFunc = oldHealth
		NodeSummaryFunc = oldSummary
	}()

	hostClusterName = "host-1"

	NewKarmadaClient = func() KarmadaClient {
		return &fakeKarm{clusters: []karmada.MemberCluster{
			{Name: "host-1", Endpoint: "https://host"},
			{Name: "member-1", Endpoint: "https://member"},
		}}
	}

	GetClusterInfos = func() ([]model.ClusterCredential, error) {
		return []model.ClusterCredential{
			{ClusterID: "host-1", APIServerURL: "https://host", BearerToken: "th"},
			{ClusterID: "member-1", APIServerURL: "https://member", BearerToken: "tm"},
		}, nil
	}

	NewNatsClient = func() NatsClient { return &fakeNats{kv: fakeStore} }

	NewKubeClient = func(cfg *rest.Config) (kubernetes.Interface, error) { return nil, nil }

	CollectMetricFunc = func(client kubernetes.Interface) (float64, float64, error) {
		return 10.0, 20.0, nil
	}
	CollectRequestMetricFunc = func(client kubernetes.Interface) (float64, float64, error) {
		return 30.0, 40.0, nil
	}
	NodeHealthCheckFunc = func(client kubernetes.Interface) string {
		return "Healthy"
	}
	NodeSummaryFunc = func(client kubernetes.Interface) (int, int) {
		return 5, 4
	}

	ctx, cancel := context.WithCancel(context.Background())
	go RepeatMetric(ctx)

	time.Sleep(1500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)

	if len(fakeStore.puts) == 0 {
		t.Fatalf("expected at least one KV Put call")
	}

	var ms model.MetricStatus
	if err := json.Unmarshal(fakeStore.puts[0], &ms); err != nil {
		t.Fatalf("invalid MetricStatus JSON stored in KV: %v", err)
	}

	if ms.HostClusterStatus.ClusterId != "host-1" {
		t.Fatalf("expected host ClusterId 'host-1', got %q", ms.HostClusterStatus.ClusterId)
	}
	if ms.HostClusterStatus.Status != "Healthy" {
		t.Fatalf("expected host Status 'Healthy', got %q", ms.HostClusterStatus.Status)
	}
	if ms.HostClusterStatus.NodeSummary.TotalNum != 5 || ms.HostClusterStatus.NodeSummary.ReadyNum != 4 {
		t.Fatalf("unexpected host NodeSummary: %+v", ms.HostClusterStatus.NodeSummary)
	}
	if len(ms.MemberClusterStatus) != 1 {
		t.Fatalf("expected 1 member cluster, got %d", len(ms.MemberClusterStatus))
	}
	if ms.MemberClusterStatus[0].ClusterId != "member-1" {
		t.Fatalf("unexpected member cluster id: %q", ms.MemberClusterStatus[0].ClusterId)
	}
}
