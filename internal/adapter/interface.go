package adapter

import (
	"federation-metric-api/model"
)

type ClusterConfigAdapter interface {
	GetClusterInfos() ([]model.ClusterCredential, error)
}
