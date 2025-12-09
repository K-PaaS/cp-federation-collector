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

func extractClusterInfos(keys []interface{}, readSecret func(path string) (*api.Secret, error)) []model.ClusterCredential {
	var infos []model.ClusterCredential
	for _, key := range keys {
		path := fmt.Sprintf("secret/data/cluster/%s", key.(string))
		secret, err := readSecret(path)
		if err != nil || secret == nil || secret.Data == nil || secret.Data["data"] == nil {
			continue
		}
		data, ok := secret.Data["data"].(map[string]interface{})
		if !ok {
			continue
		}
		apiURL, ok := data["clusterApiUrl"].(string)
		if !ok {
			continue
		}
		token, ok := data["clusterToken"].(string)
		if !ok {
			continue
		}

		infos = append(infos, model.ClusterCredential{
			ClusterID:    strings.TrimSuffix(key.(string), "/"),
			APIServerURL: apiURL,
			BearerToken:  token,
		})
	}
	return infos
}

var (
	logicalList = func(c *api.Client, path string) (*api.Secret, error) {
		return c.Logical().List(path)
	}
	logicalRead = func(c *api.Client, path string) (*api.Secret, error) {
		return c.Logical().Read(path)
	}
)

func (c *Client) GetClusterInfos() ([]model.ClusterCredential, error) {
	secrets, err := logicalList(c.api, "secret/metadata/cluster")
	if err != nil {
		return nil, err
	}

	keys, ok := secrets.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for keys")
	}

	infos := extractClusterInfos(keys, func(path string) (*api.Secret, error) {
		return logicalRead(c.api, path)
	})
	return infos, nil
}
