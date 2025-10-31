package vault

import (
	"federation-metric-api/config"
	"federation-metric-api/model"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

type Config struct {
	URL      string
	RoleID   string
	SecretID string
}

func ConfigFromEnv() *Config {
	return &Config{
		URL:      config.Env.VaultUrl,
		RoleID:   config.Env.VaultRoleId,
		SecretID: config.Env.VaultSecretId,
	}
}

type Client struct {
	api *api.Client
}

func NewClient(cfg *Config) (*Client, error) {
	vaultCfg := api.DefaultConfig()
	vaultCfg.Address = cfg.URL

	client, err := api.NewClient(vaultCfg)
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
		"role_id":   cfg.RoleID,
		"secret_id": cfg.SecretID,
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(resp.Auth.ClientToken)

	return &Client{api: client}, nil
}

func (c *Client) GetClusterInfos() ([]model.ClusterCredential, error) {
	secrets, err := c.api.Logical().List("secret/metadata/cluster")
	if err != nil {
		return nil, err
	}

	var infos []model.ClusterCredential
	keys := secrets.Data["keys"].([]interface{})
	for _, key := range keys {
		path := fmt.Sprintf("secret/data/cluster/%s", key.(string))
		secret, err := c.api.Logical().Read(path)
		if err != nil || secret == nil || secret.Data == nil || secret.Data["data"] == nil {
			continue
		}
		data := secret.Data["data"].(map[string]interface{})
		clusterApiUrl := data["clusterApiUrl"].(string)

		infos = append(infos, model.ClusterCredential{
			ClusterID:    strings.TrimSuffix(key.(string), "/"),
			APIServerURL: clusterApiUrl,
			BearerToken:  data["clusterToken"].(string),
		})

	}
	return infos, nil
}
