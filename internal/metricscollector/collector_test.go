package metricscollector

import (
	"encoding/json"
	"fmt"
	"testing"

	"federation-metric-api/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
)

func TestCountReady(t *testing.T) {
	node := model.NodeModel{
		Items: []model.Items{
			{
				Status: model.Status{
					Conditions: []model.Conditions{
						{Type: "Ready", Status: "True"},
					},
				},
			},
			{
				Status: model.Status{
					Conditions: []model.Conditions{
						{Type: "Ready", Status: "False"},
					},
				},
			},
			{
				Status: model.Status{
					Conditions: []model.Conditions{
						{Type: "MemoryPressure", Status: "True"},
						{Type: "Ready", Status: "True"},
					},
				},
			},
		},
	}

	total, ready := CountReady(node)
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if ready != 2 {
		t.Fatalf("expected ready=2, got %d", ready)
	}
}

func TestNodeSummary_UsesGetNodeListRaw(t *testing.T) {
	oldGetNodes := getNodeListRaw
	defer func() { getNodeListRaw = oldGetNodes }()

	node := model.NodeModel{
		Items: []model.Items{
			{
				Status: model.Status{
					Conditions: []model.Conditions{
						{Type: "Ready", Status: "True"},
					},
				},
			},
			{
				Status: model.Status{
					Conditions: []model.Conditions{
						{Type: "Ready", Status: "False"},
					},
				},
			},
		},
	}
	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}

	getNodeListRaw = func(client kubernetes.Interface) ([]byte, error) {
		return data, nil
	}

	total, ready := NodeSummary(nil)
	if total != 2 || ready != 1 {
		t.Fatalf("unexpected summary: total=%d ready=%d", total, ready)
	}
}

func TestNodeHealthCheck_HealthyAndUnhealthy(t *testing.T) {
	oldHealth := getHealthzRaw
	defer func() { getHealthzRaw = oldHealth }()

	getHealthzRaw = func(client kubernetes.Interface) ([]byte, error) {
		return []byte("ok"), nil
	}
	if got := NodeHealthCheck(nil); got != "True" {
		t.Fatalf("expected True, got %q", got)
	}

	getHealthzRaw = func(client kubernetes.Interface) ([]byte, error) {
		return []byte("ng"), nil
	}
	if got := NodeHealthCheck(nil); got != "False" {
		t.Fatalf("expected False, got %q", got)
	}

	getHealthzRaw = func(client kubernetes.Interface) ([]byte, error) {
		return nil, fmt.Errorf("boom")
	}
	if got := NodeHealthCheck(nil); got != "Unknown" {
		t.Fatalf("expected Unknown, got %q", got)
	}
}

func TestCollectMetric_UsesHooks(t *testing.T) {
	oldGetMetrics := getNodeMetricsRaw
	oldGetNodes := getNodeListRaw
	defer func() {
		getNodeMetricsRaw = oldGetMetrics
		getNodeListRaw = oldGetNodes
	}()

	nodeMetric := model.NodeMetricModel{
		Items: []model.NodeItem{
			{
				NodeInfo: model.MetricMetadata{Name: "node1"},
				Usage: model.NodeUsageString{
					Cpu:    "500m",
					Memory: "1024Mi",
				},
			},
		},
	}
	nmBytes, err := json.Marshal(nodeMetric)
	if err != nil {
		t.Fatalf("marshal nodeMetric: %v", err)
	}

	node := model.NodeModel{
		Items: []model.Items{
			{
				MetaData: model.NodeMetadata{NodeName: "node1"},
				Status: model.Status{
					Capacity: model.Capacity{
						Cpu:    "1",
						Memory: "2048Mi",
					},
				},
			},
		},
	}
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal node: %v", err)
	}

	getNodeMetricsRaw = func(client kubernetes.Interface) ([]byte, error) {
		return nmBytes, nil
	}
	getNodeListRaw = func(client kubernetes.Interface) ([]byte, error) {
		return nodeBytes, nil
	}

	cpuRatio, memRatio, err := CollectMetric(nil)
	if err != nil {
		t.Fatalf("CollectMetric returned error: %v", err)
	}
	if cpuRatio <= 0 || memRatio <= 0 {
		t.Fatalf("expected positive ratios, got cpu=%f mem=%f", cpuRatio, memRatio)
	}
}

func TestCollectRequestMetric_UsesHooks(t *testing.T) {
	oldGetNodes := getNodeListRaw
	oldListPods := listPodsOnNode
	defer func() {
		getNodeListRaw = oldGetNodes
		listPodsOnNode = oldListPods
	}()

	node := model.NodeModel{
		Items: []model.Items{
			{
				MetaData: model.NodeMetadata{NodeName: "node1"},
				Status: model.Status{
					Allocatable: model.Allocatable{
						Cpu:    "2",
						Memory: "4096Mi",
					},
				},
			},
		},
	}
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal node: %v", err)
	}

	getNodeListRaw = func(client kubernetes.Interface) ([]byte, error) {
		return nodeBytes, nil
	}

	cpuReq := resource.MustParse("500m")
	memReq := resource.MustParse("1024Mi")
	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuReq,
									corev1.ResourceMemory: memReq,
								},
							},
						},
					},
				},
			},
		},
	}

	listPodsOnNode = func(client kubernetes.Interface, nodeName string) (*corev1.PodList, error) {
		if nodeName != "node1" {
			t.Fatalf("unexpected nodeName: %q", nodeName)
		}
		return podList, nil
	}

	cpuRatio, memRatio, err := CollectRequestMetric(nil)
	if err != nil {
		t.Fatalf("CollectRequestMetric returned error: %v", err)
	}
	if cpuRatio <= 0 || memRatio <= 0 {
		t.Fatalf("expected positive ratios, got cpu=%f mem=%f", cpuRatio, memRatio)
	}
}
