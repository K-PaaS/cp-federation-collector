package model

import "time"

type NodeMetricModel struct {
	Kind       string     `json:"kind"`
	ApiVersion string     `json:"apiVersion"`
	Items      []NodeItem `json:"items"`
}

type NodeItem struct {
	NodeInfo  MetricMetadata  `json:"metadata"`
	Timestamp string          `json:"timestamp"`
	Window    string          `json:"window"`
	Usage     NodeUsageString `json:"usage"`
}

type MetricMetadata struct {
	Name              string `json:"name"`
	CreationTimestamp string `json:"creationTimestamp"`
}

type NodeUsageString struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
}

type NodeUsageFloat struct {
	Cpu    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

// METRIC JSON STRUCT
type NodeSummary struct {
	TotalNum int `json:"totalNum"`
	ReadyNum int `json:"readyNum"`
}
type MetricStatus struct {
	Time                time.Time             `json:"time"`
	HostClusterStatus   HostClusterStatus     `json:"hostClusterStatus"`
	MemberClusterStatus []MemberClusterStatus `json:"memberClusterStatus"`
}

type HostClusterStatus struct {
	ClusterId     string         `json:"clusterId"`
	Status        string         `json:"status"`
	NodeSummary   NodeSummary    `json:"nodeSummary"`
	RealTimeUsage NodeUsageFloat `json:"realTimeUsage"`
	RequestUsage  NodeUsageFloat `json:"requestUsage"`
}
type MemberClusterStatus struct {
	ClusterId     string         `json:"clusterId"`
	RealTimeUsage NodeUsageFloat `json:"realTimeUsage"`
}
