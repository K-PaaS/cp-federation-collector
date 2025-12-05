package vault

import (
	"errors"
	"testing"

	"federation-metric-api/config"
	"github.com/hashicorp/vault/api"
)

func TestConfigFromEnv_UsesConfigEnv(t *testing.T) {
	if config.Env == nil {
		t.Fatalf("config.Env is nil")
	}

	oldURL := config.Env.VaultUrl
	oldRole := config.Env.VaultRoleId
	oldSecret := config.Env.VaultSecretId

	defer func() {
		config.Env.VaultUrl = oldURL
		config.Env.VaultRoleId = oldRole
		config.Env.VaultSecretId = oldSecret
	}()

	config.Env.VaultUrl = "http://vault.example"
	config.Env.VaultRoleId = "role-id"
	config.Env.VaultSecretId = "secret-id"

	cfg := ConfigFromEnv()
	if cfg.URL != "http://vault.example" {
		t.Fatalf("URL mismatch: got %q, want %q", cfg.URL, "http://vault.example")
	}
	if cfg.RoleID != "role-id" {
		t.Fatalf("RoleID mismatch: got %q, want %q", cfg.RoleID, "role-id")
	}
	if cfg.SecretID != "secret-id" {
		t.Fatalf("SecretID mismatch: got %q, want %q", cfg.SecretID, "secret-id")
	}
}

func TestNewClient_InvalidAddress(t *testing.T) {
	cfg := &Config{
		URL:      "http://127.0.0.1:0",
		RoleID:   "role",
		SecretID: "secret",
	}

	client, err := NewClient(cfg)
	if err == nil {
		t.Fatalf("expected error for invalid address, got nil (client=%+v)", client)
	}
}

func Test_extractClusterInfos_ParsesValidSecrets(t *testing.T) {
	keys := []interface{}{"cluster-a/", "cluster-b/"}

	secrets := map[string]*api.Secret{
		"secret/data/cluster/cluster-a/": {
			Data: map[string]interface{}{
				"data": map[string]interface{}{
					"clusterApiUrl": "https://a.api",
					"clusterToken":  "token-a",
				},
			},
		},
		"secret/data/cluster/cluster-b/": {
			Data: map[string]interface{}{
				"data": map[string]interface{}{
					"clusterApiUrl": "https://b.api",
					"clusterToken":  "token-b",
				},
			},
		},
	}

	got := extractClusterInfos(keys, func(path string) (*api.Secret, error) {
		if s, ok := secrets[path]; ok {
			return s, nil
		}
		return nil, nil
	})

	if len(got) != 2 {
		t.Fatalf("expected 2 cluster infos, got %d", len(got))
	}
	if got[0].ClusterID != "cluster-a" || got[0].APIServerURL != "https://a.api" || got[0].BearerToken != "token-a" {
		t.Fatalf("unexpected first cluster: %+v", got[0])
	}
	if got[1].ClusterID != "cluster-b" || got[1].APIServerURL != "https://b.api" || got[1].BearerToken != "token-b" {
		t.Fatalf("unexpected second cluster: %+v", got[1])
	}
}

func Test_extractClusterInfos_SkipsInvalid(t *testing.T) {
	keys := []interface{}{"ok/", "bad/"}

	secrets := map[string]*api.Secret{
		"secret/data/cluster/ok/": {
			Data: map[string]interface{}{
				"data": map[string]interface{}{
					"clusterApiUrl": "https://ok.api",
					"clusterToken":  "token-ok",
				},
			},
		},
	}

	got := extractClusterInfos(keys, func(path string) (*api.Secret, error) {
		if s, ok := secrets[path]; ok {
			return s, nil
		}
		return nil, nil
	})

	if len(got) != 1 {
		t.Fatalf("expected 1 valid cluster, got %d", len(got))
	}
	if got[0].ClusterID != "ok" {
		t.Fatalf("unexpected cluster id: %q", got[0].ClusterID)
	}
}

func TestGetClusterInfos_ListError(t *testing.T) {
	oldList := logicalList
	oldRead := logicalRead
	defer func() {
		logicalList = oldList
		logicalRead = oldRead
	}()

	logicalList = func(c *api.Client, path string) (*api.Secret, error) {
		return nil, errors.New("list failed")
	}

	c := &Client{api: &api.Client{}}
	infos, err := c.GetClusterInfos()
	if err == nil {
		t.Fatalf("expected error, got nil (infos=%v)", infos)
	}
}

func TestGetClusterInfos_UnexpectedKeysType(t *testing.T) {
	oldList := logicalList
	oldRead := logicalRead
	defer func() {
		logicalList = oldList
		logicalRead = oldRead
	}()

	logicalList = func(c *api.Client, path string) (*api.Secret, error) {
		return &api.Secret{
			Data: map[string]interface{}{
				"keys": "not-a-slice",
			},
		}, nil
	}

	c := &Client{api: &api.Client{}}
	infos, err := c.GetClusterInfos()
	if err == nil {
		t.Fatalf("expected error for unexpected keys type, got nil (infos=%v)", infos)
	}
}

func TestGetClusterInfos_SuccessWithHooks(t *testing.T) {
	oldList := logicalList
	oldRead := logicalRead
	defer func() {
		logicalList = oldList
		logicalRead = oldRead
	}()

	logicalList = func(c *api.Client, path string) (*api.Secret, error) {
		if path != "secret/metadata/cluster" {
			t.Fatalf("unexpected list path: %q", path)
		}
		return &api.Secret{
			Data: map[string]interface{}{
				"keys": []interface{}{"cluster-x/"},
			},
		}, nil
	}

	logicalRead = func(c *api.Client, path string) (*api.Secret, error) {
		if path != "secret/data/cluster/cluster-x/" {
			t.Fatalf("unexpected read path: %q", path)
		}
		return &api.Secret{
			Data: map[string]interface{}{
				"data": map[string]interface{}{
					"clusterApiUrl": "https://x.api",
					"clusterToken":  "token-x",
				},
			},
		}, nil
	}

	c := &Client{api: &api.Client{}}
	infos, err := c.GetClusterInfos()
	if err != nil {
		t.Fatalf("GetClusterInfos returned error: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(infos))
	}
	if infos[0].ClusterID != "cluster-x" || infos[0].APIServerURL != "https://x.api" || infos[0].BearerToken != "token-x" {
		t.Fatalf("unexpected cluster info: %+v", infos[0])
	}
}
