package controller

import (
	"context"
	"encoding/json"
	"federation-metric-api/config"
	"federation-metric-api/internal/adapter"
	"federation-metric-api/internal/karmada"
	"federation-metric-api/internal/metricscollector"
	"federation-metric-api/internal/nats"
	"federation-metric-api/internal/util"
	"federation-metric-api/model"
	outnats "github.com/nats-io/nats.go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"time"
)

type KarmadaClient interface {
	GetMemberClusters(ctx context.Context) ([]karmada.MemberCluster, error)
}

type NatsClient interface {
	CreateKeyValue(bucket string) (outnats.KeyValue, error)
	KeyValue(bucket string) (outnats.KeyValue, error)
}

var (
	NewKarmadaClient = func() KarmadaClient { return karmada.NewClient() }
	NewNatsClient    = func() NatsClient { return nats.NewClient() }
	GetClusterInfos  = adapter.GetClusterInfos

	NewKubeClient            = func(cfg *rest.Config) (kubernetes.Interface, error) { return kubernetes.NewForConfig(cfg) }
	CollectMetricFunc        = metricscollector.CollectMetric
	CollectRequestMetricFunc = metricscollector.CollectRequestMetric
	NodeHealthCheckFunc      = metricscollector.NodeHealthCheck
	NodeSummaryFunc          = metricscollector.NodeSummary
)

var hostClusterName string

var natsBucketName string
var repeatTime time.Duration = 30

var natsSubjectName string

func init() {
	hostClusterName = config.Env.HostClusterName
	natsBucketName = config.Env.NatsBucketName
	natsSubjectName = config.Env.NatsSubjectName

}

func RepeatMetric(ctx context.Context) {
	ticker := time.NewTicker(repeatTime * time.Second)
	defer ticker.Stop()

	karmadaClient := NewKarmadaClient()
	natsClient := NewNatsClient()
	if natsClient == nil {

	}

	kv, err := natsClient.CreateKeyValue(natsBucketName)
	if err != nil {
		kv, err = natsClient.KeyValue(natsBucketName)
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		clusterInfos, err := GetClusterInfos()
		memberClusters, err := karmadaClient.GetMemberClusters(ctx)
		if err != nil {
			log.Fatalf("Karmada member 클러스터 조회 실패: %v", err)
		}

		var ClCpuRatio, ClMemRatio float64
		var hostCluster model.HostClusterStatus
		var metricStatus model.MetricStatus

		var memberClusterList []model.MemberClusterStatus
		for _, ci := range clusterInfos {
			if ci.ClusterID == hostClusterName {
				cfg := &rest.Config{
					Host:        ci.APIServerURL,
					BearerToken: ci.BearerToken,
					TLSClientConfig: rest.TLSClientConfig{
						Insecure: true,
					},
				}
				clientset, err := NewKubeClient(cfg)
				if err != nil {
					log.Printf("%s 클러스터 clientset 생성 실패: %v", ci.ClusterID, err)
					continue
				}

				ClCpuRatio, ClMemRatio, err = CollectMetricFunc(clientset)
				//Status 구하는 로직
				hostCluster.Status = NodeHealthCheckFunc(clientset)
				//Node Summary 구하는 로직
				totalNum, readyNum := NodeSummaryFunc(clientset)
				hostCluster.NodeSummary = model.NodeSummary{
					TotalNum: totalNum,
					ReadyNum: readyNum,
				}
				//RequestUsage 구하는 로직
				requestCPURatio, requestMemRatio, _ := CollectRequestMetricFunc(clientset)
				hostCluster.RequestUsage = model.NodeUsageFloat{
					Cpu:    util.Round(requestCPURatio, 2), //fmt.Sprintf("%.2f", requestCPURatio),
					Memory: util.Round(requestMemRatio, 2), //.Sprintf("%.2f", requestMemRatio),
				}

				hostCluster.ClusterId = hostClusterName
				hostCluster.RealTimeUsage = model.NodeUsageFloat{
					Cpu:    util.Round(ClCpuRatio, 2), //fmt.Sprintf("%.2f", ClCpuRatio),
					Memory: util.Round(ClMemRatio, 2), //fmt.Sprintf("%.2f", ClMemRatio),
				}

				continue
			}
			for _, member := range memberClusters {
				if member.Endpoint == ci.APIServerURL || ci.ClusterID != hostClusterName {
					cfg := &rest.Config{
						Host:        ci.APIServerURL,
						BearerToken: ci.BearerToken,
						TLSClientConfig: rest.TLSClientConfig{
							Insecure: true,
						},
					}
					clientset, err := NewKubeClient(cfg)
					if err != nil {
						log.Printf("%s 클러스터 clientset 생성 실패: %v", ci.ClusterID, err)
						continue
					}

					ClCpuRatio, ClMemRatio, err = CollectMetricFunc(clientset)

					memberClusterList = append(memberClusterList, model.MemberClusterStatus{
						ClusterId: ci.ClusterID,
						RealTimeUsage: model.NodeUsageFloat{
							Cpu:    util.Round(ClCpuRatio, 2), //fmt.Sprintf("%.2f", ClCpuRatio),
							Memory: util.Round(ClMemRatio, 2), // fmt.Sprintf("%.2f", ClMemRatio),
						},
					})

					break
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metricStatus = model.MetricStatus{
				HostClusterStatus:   hostCluster,
				MemberClusterStatus: memberClusterList,
				Time:                time.Now().UTC(),
			}
			data, _ := json.Marshal(metricStatus)

			_, err = kv.Put(natsSubjectName, data)
			if err != nil {
				log.Printf("Failed to send metrics: %v", err)
			} else {
				log.Printf("Metric transfer complete")
			}
		}
	}
}
