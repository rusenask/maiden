package maiden

import (
	docker "github.com/fsouza/go-dockerclient"
)

// Distributor - placeholder for distributor
type Distributor interface {
	ShareImage() error
	PullImage() error
}

type DefaultDistributor struct {
	client *docker.Client
}

func NewDefaultDistributor(client *docker.Client) *DefaultDistributor {
	return &DefaultDistributor{client: client}
}

func (d *DefaultDistributor) ShareImage() error {
	return nil
}
