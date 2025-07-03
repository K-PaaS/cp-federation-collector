package adapter

import (
	"federation-metric-api/internal/vault"
	"federation-metric-api/model"
	"log"
)

func GetClusterInfos() ([]model.ClusterCredential, error) {
	vaultCfg := vault.ConfigFromEnv()
	vaultClient, err := vault.NewClient(vaultCfg)
	if err != nil {
		log.Fatalf("Vault 클라이언트 생성 실패: %v", err)
	}
	clusterInfos, err := vaultClient.GetClusterInfos()
	if err != nil {
		log.Fatalf("클러스터 정보 조회 실패: %v", err)
	}
	return clusterInfos, err
}
