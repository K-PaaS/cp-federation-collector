package model

type NodeModel struct {
	Kind       string  `json:"kind"`
	ApiVersion string  `json:"apiVersion"`
	Items      []Items `json:"items"`
}

type Items struct {
	MetaData NodeMetadata `json:"metadata"`
	Status   Status       `json:"status"`
}

type NodeMetadata struct {
	NodeName string `json:"name"`
}

type Status struct {
	Capacity    Capacity          `json:"capacity"`
	Allocatable Allocatable       `json:"allocatable"`
	Addresses   []Addresses       `json:"addresses"`
	NodeInfo    NodeInfo          `json:"nodeInfo"`
	Phase       string            `json:"phase"`
	ConStatus   []ContainerStatus `json:"containerStatuses"`
	Conditions  []Conditions      `json:"conditions"`
}

type Conditions struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type Capacity struct {
	Cpu              string `json:"cpu"`
	EphemeralStorage string `json:"ephemeral-storage"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
}

type Allocatable struct {
	Cpu              string `json:"cpu"`
	EphemeralStorage string `json:"ephemeral-storage"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
}

type Addresses struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

type NodeInfo struct {
	MachineID               string `json:"machineID"`
	SystemUUID              string `json:"systemUUID"`
	BootID                  string `json:"bootID"`
	KernelVersion           string `json:"kernelVersion"`
	OsImage                 string `json:"osImage"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
	KubeletVersion          string `json:"kubeletVersion"`
	KubeProxyVersion        string `json:"kubeProxyVersion"`
	OperationSystem         string `json:"operationSystem"`
	Architecture            string `json:"architecture"`
}

type NodeStateInfo struct {
	ClusterId  string  `sql:"cluster_id"`
	NodeName   string  `sql:"node_name"`
	CpuRatio   float64 `sql:"cpu_ratio"`
	MemRatio   float64 `sql:"mem_ratio"`
	UpdateTime string  `sql:"update_time"`
}

type ContainerStatus struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
}
