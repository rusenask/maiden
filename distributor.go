package maiden

import (
	"net"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
	docker "github.com/fsouza/go-dockerclient"
)

// ImageDownloadPath - default image download/seed path
// const ImageDownloadPath = "images"
const ImageDownloadPath = ""

// DHTDistributorConfig - distributor config
type DHTDistributorConfig struct {
	MMap         bool           // memory-map torrent data
	Peers        []*net.TCPAddr //addresses of some starting Peers
	Addr         *net.TCPAddr   // network listen addr
	UploadRate   int64
	DownloadRate int64
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
	filename := generateImageName(name)
	err := d.getImage(name, filename)
	if err != nil {
		return err
	}

	err = d.createTorrentFile(filename)
	if err != nil {
		return err
	}

	// Starting seed
	tPath := filepath.Join(ImageDownloadPath, getTorrentName(filename))
	err = d.addTorrents([]string{tPath})
	if err != nil {
		return err
	}

	d.seed()

	return nil
}

func (d *DefaultDistributor) createTorrentFile(name string) error {
	contents, err := Create(filepath.Join(ImageDownloadPath, name))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(ImageDownloadPath, getTorrentName(name)))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(contents)
	return err
}
