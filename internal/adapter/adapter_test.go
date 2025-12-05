package adapter

import (
	"errors"
	"testing"

	"federation-metric-api/internal/vault"
	"federation-metric-api/model"
)

type fakeVaultClient struct {
	infos []model.ClusterCredential
	err   error
}

func (f *fakeVaultClient) GetClusterInfos() ([]model.ClusterCredential, error) {
	return f.infos, f.err
}

func Test_getClusterInfosFrom_Success(t *testing.T) {
	expected := []model.ClusterCredential{
		{ClusterID: "c1", APIServerURL: "https://c1.example", BearerToken: "token1"},
		{ClusterID: "c2", APIServerURL: "https://c2.example", BearerToken: "token2"},
	}

	fake := &fakeVaultClient{infos: expected}
	got, err := getClusterInfosFrom(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d infos, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("index %d mismatch: got %+v, want %+v", i, got[i], expected[i])
		}
	}
}

func Test_getClusterInfosFrom_Error(t *testing.T) {
	fake := &fakeVaultClient{err: errors.New("boom")}

	_, err := getClusterInfosFrom(fake)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

type fakeVaultFull struct {
	infos []model.ClusterCredential
	err   error
}

func (f *fakeVaultFull) GetClusterInfos() ([]model.ClusterCredential, error) {
	return f.infos, f.err
}

func TestGetClusterInfos_Success(t *testing.T) {
	oldNew := newVaultClient
	newVaultClient = func(cfg *vault.Config) (VaultClusterInfoClient, error) {
		return &fakeVaultFull{
			infos: []model.ClusterCredential{
				{ClusterID: "x1", APIServerURL: "https://x1.example", BearerToken: "tt"},
			},
		}, nil
	}
	defer func() { newVaultClient = oldNew }()

	got, err := GetClusterInfos()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ClusterID != "x1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetClusterInfos_NewClientError(t *testing.T) {
	oldNew := newVaultClient
	newVaultClient = func(cfg *vault.Config) (VaultClusterInfoClient, error) {
		return nil, errors.New("cannot-init-client")
	}
	defer func() { newVaultClient = oldNew }()

	_, err := GetClusterInfos()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetClusterInfos_LoadError(t *testing.T) {
	oldNew := newVaultClient
	newVaultClient = func(cfg *vault.Config) (VaultClusterInfoClient, error) {
		return &fakeVaultFull{err: errors.New("fetch-failed")}, nil
	}
	defer func() { newVaultClient = oldNew }()

	_, err := GetClusterInfos()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
