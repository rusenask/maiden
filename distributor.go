package maiden

import (
	"fmt"
	"os"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
)

const ImageDownloadPath = "images"

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

	err = d.createTorrentFile(name)
	if err != nil {
		return err
	}

	return nil
}

func (d *DefaultDistributor) createTorrentFile(name string) error {
	contents, err := Create(filepath.Join(ImageDownloadPath, name))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(ImageDownloadPath, fmt.Sprintf("image-%s.torrent", name)))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(contents)
	return err
}

func (d *DefaultDistributor) getImage(name string) error {
	// checking whether we have this image
	path := filepath.Join(ImageDownloadPath, name)

	// TODO: sensible perms?
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

	f, err := os.Create(filepath.Join(path, name))
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
