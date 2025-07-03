package main

import (
	"context"
	"federation-metric-api/controller"
)

var hostClusterName = "host-cluster"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	controller.RepeatMetric(ctx)
	/*
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			<-c
			cancel()
		}()

		//karmadaToken := os.Getenv("KARMADA_TOKEN")
		//karmadaAPI := os.Getenv("KARMADA_API")
		//karmadaClient := karmada.NewClient(karmadaAPI, karmadaToken)
		//karmadaClient := karmada.NewClient()
		karmadaClient := karmada.NewClient()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		nc, err := nats.Connect(os.Getenv("NATS_URL"), nats.UserInfo("cpnats", "cpnats"))
		if err != nil {
			log.Printf("NATS 연결 실패: %v", err)
			return
		}
		defer nc.Drain()

		js, err := nc.JetStream()
		if err != nil {
			log.Fatal(err)
		}
		kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "recent_values",
		})
		if err != nil {
			kv, err = js.KeyValue("recent_values")
			if err != nil {
				log.Fatal(err)
			}
		}

		for {
			clusterInfos, err := adapter.GetClusterInfos()
			memberClusters, err := karmadaClient.GetMemberClusters(ctx)
			if err != nil {
				log.Fatalf("Karmada member 클러스터 조회 실패: %v", err)
			}

			var ClCpuRatio, ClMemRatio float64
			var hostCluster model.HostClusterStatus
			var metricStatus model.MetricStatus

			var memberClusterList []model.MemberClusterStatus
			for _, ci := range clusterInfos {
				for _, member := range memberClusters {
					if member.Endpoint == ci.APIServerURL || ci.ClusterID == hostClusterName {
						cfg := &rest.Config{
							Host:        ci.APIServerURL,
							BearerToken: ci.BearerToken,
							TLSClientConfig: rest.TLSClientConfig{
								Insecure: true,
							},
						}
						clientset, err := kubernetes.NewForConfig(cfg)
						if err != nil {
							log.Printf("%s 클러스터 clientset 생성 실패: %v", ci.ClusterID, err)
							continue
						}

						ClCpuRatio, ClMemRatio, err = metricscollector.CollectMetric(clientset)

						if ci.ClusterID == hostClusterName {
							//Status 구하는 로직
							hostCluster.Status = metricscollector.NodeHealthCheck(clientset)
							//Node Summary 구하는 로직
							totalNum, readyNum := metricscollector.NodeSummary(clientset)
							hostCluster.NodeSummary = model.NodeSummary{
								TotalNum: totalNum,
								ReadyNum: readyNum,
							}
							//RequestUsage 구하는 로직
							requestCPURatio, requestMemRatio, _ := metricscollector.CollectRequestMetric(clientset)
							hostCluster.RequestUsage = model.NodeUsage{
								Cpu:    fmt.Sprintf("%.2f", requestCPURatio),
								Memory: fmt.Sprintf("%.2f", requestMemRatio),
							}

							hostCluster.ClusterId = hostClusterName
							hostCluster.RealTimeUsage = model.NodeUsage{
								Cpu:    fmt.Sprintf("%.2f", ClCpuRatio),
								Memory: fmt.Sprintf("%.2f", ClMemRatio),
							}
						} else {
							memberClusterList = append(memberClusterList, model.MemberClusterStatus{
								ClusterId: ci.ClusterID,
								RealTimeUsage: model.NodeUsage{
									Cpu:    fmt.Sprintf("%.2f", ClCpuRatio),
									Memory: fmt.Sprintf("%.2f", ClMemRatio),
								},
							})
						}
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
				//err := nc.Publish("cluster.metrics", data)

				_, err = kv.Put("cluster.metrics", data)
				if err != nil {
					log.Printf("Failed to send metrics: %v", err)
				} else {
					log.Printf("Metric transfer complete")
				}

				latest, _ := kv.Get("cluster.metrics")
				fmt.Println(string(latest.Value()))
			}
		}*/
}
