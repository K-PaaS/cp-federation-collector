package adapter

import (
	"fmt"

	"federation-metric-api/internal/vault"
	"federation-metric-api/model"
)

type VaultClusterInfoClient interface {
	GetClusterInfos() ([]model.ClusterCredential, error)
}

var newVaultClient = func(cfg *vault.Config) (VaultClusterInfoClient, error) {
	return vault.NewClient(cfg)
}

func getClusterInfosFrom(c VaultClusterInfoClient) ([]model.ClusterCredential, error) {
	return c.GetClusterInfos()
}

func GetClusterInfos() ([]model.ClusterCredential, error) {
	vaultCfg := vault.ConfigFromEnv()
	vaultClient, err := newVaultClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}
	clusterInfos, err := getClusterInfosFrom(vaultClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load cluster infos: %w", err)
	}
	return clusterInfos, nil
}
