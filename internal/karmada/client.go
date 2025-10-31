package karmada

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"federation-metric-api/config"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	api    string
	token  string
	client *http.Client
}

func NewClient() *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 사설 인증서 무시
		},
	}
	return &Client{
		api:    config.Env.KarmadaApi,
		token:  config.Env.KarmadaToken,
		client: &http.Client{Transport: tr},
	}
}

type MemberCluster struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

func (c *Client) GetMemberClusters(ctx context.Context) ([]MemberCluster, error) {
	url := fmt.Sprintf("%s/apis/cluster.karmada.io/v1alpha1/clusters", c.api)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("karmada 요청 실패: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		//body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("karmada 응답 오류: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				APIEndpoint string `json:"apiEndpoint"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	clusters := make([]MemberCluster, 0)
	for _, item := range result.Items {
		clusters = append(clusters, MemberCluster{
			Name:     item.Metadata.Name,
			Endpoint: item.Spec.APIEndpoint,
		})
	}
	return clusters, nil

}
