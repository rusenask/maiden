package maiden

import (
	"os"
	// "path/filepath"

	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
)

func (d *DefaultDistributor) exportImage(image, filename string) error {
	// checking whether we have this image
	// path := filepath.Join(ImageDownloadPath, filename)

	// TODO: sensible perms?
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.MkdirAll(filename, os.ModePerm)
	}

	f, err := os.Create(imagePath(filename))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to create data torrent")
	}
	defer f.Close()

	opts := docker.ExportImageOptions{
		Name:         image,
		OutputStream: f,
	}
	log.WithFields(log.Fields{
		"image":    image,
		"filename": filename,
	}).Info("image exported")
	return d.dClient.ExportImage(opts)
}
