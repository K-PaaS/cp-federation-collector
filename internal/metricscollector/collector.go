package metricscollector

import (
	"context"
	"encoding/json"
	"federation-metric-api/model"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

func CollectRequestMetric(clientset kubernetes.Interface) (float64, float64, error) {
	var node model.NodeModel
	nodeData, err := clientset.NodeV1().RESTClient().Get().AbsPath("api/v1/nodes").DoRaw(context.TODO())
	if err != nil {
		return -1, -1, err
	}
	err = json.Unmarshal(nodeData, &node)

	// 필터링된 Pod 목록 가져오기

	totalAllocatableCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalAllocatableMem := resource.NewQuantity(0, resource.BinarySI)
	totalRequestCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalRequestMem := resource.NewQuantity(0, resource.BinarySI)
	for _, i := range node.Items {
		podList, _ := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("spec.nodeName", i.MetaData.NodeName).String(),
		})
		totalNodeRequestCPU := resource.NewQuantity(0, resource.DecimalSI)
		totalNodeRequestMem := resource.NewQuantity(0, resource.BinarySI)
		for _, pod := range podList.Items {
			if pod.Status.Phase == "Running" {
				for _, c := range pod.Spec.Containers {
					if cpuQty, ok := c.Resources.Requests["cpu"]; ok {
						totalNodeRequestCPU.Add(cpuQty)
					}
					if memQty, ok := c.Resources.Requests["memory"]; ok {
						totalNodeRequestMem.Add(memQty)
					}
				}
			}
		}

		totalRequestCPU.Add(*totalNodeRequestCPU)
		totalRequestMem.Add(*totalNodeRequestMem)

		allocatableCPU, _ := resource.ParseQuantity(i.Status.Allocatable.Cpu)
		totalAllocatableCPU.Add(allocatableCPU)
		allocatableMem, _ := resource.ParseQuantity(i.Status.Allocatable.Memory)
		totalAllocatableMem.Add(allocatableMem)
	}
	return (float64(totalRequestCPU.MilliValue()) / float64(totalAllocatableCPU.MilliValue())) * 100, (float64(totalRequestMem.Value()) / float64(totalAllocatableMem.Value())) * 100, nil

}

func CollectMetric(clientset kubernetes.Interface) (float64, float64, error) {
	var CpuN float64 = 1000000000

	var nodeMetric model.NodeMetricModel
	nodeMetricData, err := clientset.NodeV1().RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/nodes").DoRaw(context.TODO())
	if err != nil {
		return -1, -1, err
	}
	err = json.Unmarshal(nodeMetricData, &nodeMetric)

	var node model.NodeModel
	nodeData, err := clientset.NodeV1().RESTClient().Get().AbsPath("api/v1/nodes").DoRaw(context.TODO())
	if err != nil {
		return -1, -1, err
	}
	err = json.Unmarshal(nodeData, &node)

	var ClCpuRatio, ClMemRatio, ClCpuRaw, ClMemRaw, ClCpuCore, ClMemSize float64

	for i := range nodeMetric.Items {

		for j := range node.Items {

			if nodeMetric.Items[i].NodeInfo.Name == node.Items[j].MetaData.NodeName {
				cpu := nodeMetric.Items[i].Usage.Cpu
				cpu = cpu[:len(cpu)-1]
				cpuRaw, _ := strconv.ParseFloat(cpu, 64)
				ClCpuRaw += cpuRaw

				memory := nodeMetric.Items[i].Usage.Memory
				memory = memory[:len(memory)-2]
				memRaw, _ := strconv.ParseFloat(memory, 64)
				ClMemRaw += memRaw

				cpuCore, _ := strconv.ParseFloat(node.Items[j].Status.Capacity.Cpu, 64)
				memory = node.Items[j].Status.Capacity.Memory
				memory = memory[:len(memory)-2]
				memSize, _ := strconv.ParseFloat(memory, 64)
				ClCpuCore += cpuCore
				ClMemSize += memSize
			}
		}

	}
	ClCpuRatio = (ClCpuRaw / (ClCpuCore * CpuN)) * 100
	ClMemRatio = (ClMemRaw / ClMemSize) * 100

	return ClCpuRatio, ClMemRatio, nil
}

func NodeHealthCheck(clientset kubernetes.Interface) string {
	content, err := clientset.Discovery().RESTClient().Get().AbsPath("/healthz").DoRaw(context.TODO())
	if err != nil {
		return "Unknown"
	}
	contentStr := string(content)
	if contentStr != "ok" {
		return "False"
	}
	return "True"
}

func NodeSummary(clientset kubernetes.Interface) (int, int) {
	var node model.NodeModel
	nodeData, err := clientset.NodeV1().RESTClient().Get().AbsPath("api/v1/nodes").DoRaw(context.TODO())
	if err != nil {
		return -1, -1
	}
	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return -1, -1
	}
	var readyNum = 0
	for _, item := range node.Items {
		for _, condition := range item.Status.Conditions {
			if (condition.Type == "Ready") && (condition.Status == "True") {
				readyNum++
			}
		}
	}
	return len(node.Items), readyNum
}
