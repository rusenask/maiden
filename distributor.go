package maiden

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
)

// ImageDownloadPath - default image download/seed path
const ImageDownloadPath = "images"

// DHTDistributorConfig - distributor config
type DHTDistributorConfig struct {
	mMap         bool           // memory-map torrent data
	peers        []*net.TCPAddr //addresses of some starting peers
	addr         *net.TCPAddr   // network listen addr
	uploadRate   int64
	downloadRate int64
}

// Distributor - placeholder for distributor
type Distributor interface {
	ShareImage(name string) error
	PullImage() error
}

// DefaultDistributor - default DHT based image distributor
type DefaultDistributor struct {
	cfg *DHTDistributorConfig

	dClient *docker.Client
	tClinet *torrent.Client
}

// NewDHTDistributor - create new default DHT Distributor
func NewDHTDistributor(cfg *DHTDistributorConfig, dClient *docker.Client) (*DefaultDistributor, error) {
	dist := &DefaultDistributor{
		cfg:     cfg,
		dClient: dClient,
	}

	// preparing torrent client
	tc, err := dist.getTorrentClient()
	if err != nil {
		return nil, err
	}
	dist.tClinet = tc
	return dist, nil
}

// ShareImage - start sharing specified image
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

	return d.dClient.ExportImage(opts)
}
