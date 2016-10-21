package maiden

import (
	// "fmt"
	"os"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
)

// Distributor - placeholder for distributor
type Distributor interface {
	ShareImage(name string) error
	PullImage() error
}

type DefaultDistributor struct {
	client *docker.Client
}

func NewDefaultDistributor(client *docker.Client) *DefaultDistributor {
	return &DefaultDistributor{client: client}
}

func (d *DefaultDistributor) ShareImage(name string) error {
	err := d.getImage(name)
	if err != nil {
		return err
	}

	return nil
}

func (d *DefaultDistributor) getImage(name string) error {
	// checking whether we have this image
	f, err := os.Create(filepath.Join("images", name))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to create data torrent")
	}
	defer f.Close()

	opts := docker.ExportImageOptions{
		Name:         name,
		OutputStream: f,
	}

	return d.client.ExportImage(opts)
}
