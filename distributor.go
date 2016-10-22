package maiden

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/anacrolix/torrent"
	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
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
	Serve(ctx context.Context) error
	Shutdown() error

	ShareImage(name string) (torrent []byte, err error)
	StopSharing(name string) error

	PullImage(name string) error
}

// DefaultDistributor - default DHT based image distributor
type DefaultDistributor struct {
	cfg *DHTDistributorConfig

	dClient *docker.Client
	tClinet *torrent.Client

	mutex  *sync.Mutex
	active map[string]*torrent.Torrent
}

// NewDHTDistributor - create new default DHT Distributor
func NewDHTDistributor(cfg *DHTDistributorConfig, dClient *docker.Client) (*DefaultDistributor, error) {
	mu := &sync.Mutex{}
	active := make(map[string]*torrent.Torrent)

	dist := &DefaultDistributor{
		cfg:     cfg,
		dClient: dClient,
		mutex:   mu,
		active:  active,
	}

	// preparing torrent client
	tc, err := dist.getTorrentClient()
	if err != nil {
		return nil, err
	}
	dist.tClinet = tc
	return dist, nil
}

// Serve - continues serving torrent files
func (d *DefaultDistributor) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("default distributor: stopping...")
			// cleaning up
			d.Shutdown()
			return nil
		}
	}
}

// Shutdown - cleansup and shuts down torrent server
func (d *DefaultDistributor) Shutdown() error {
	for k := range d.active {
		err := d.StopSharing(k)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"image": k,
			}).Error("got error while cleaning up")
		}
	}

	// done, unsubscribing
	d.tClinet.Close()

	return nil
}

// PullImage - pulls image from network and imports it
func (d *DefaultDistributor) PullImage(name string) error {

	filename := imagePath(generateImageName(name))
	// opening file
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("image not found")
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = d.importImage(f)
	if err != nil {
		return err
	}

	// success
	return nil
}

// ShareImage - start sharing specified image
func (d *DefaultDistributor) ShareImage(name string) (torrent []byte, err error) {
	filename := generateImageName(name)
	err = d.exportImage(name, filename)
	if err != nil {
		return
	}

	torrent, err = d.createTorrentFile(filename)
	if err != nil {
		return
	}

	// Starting seed
	tPath := getTorrentName(filename)
	err = d.addTorrent(name, tPath)
	if err != nil {
		return
	}

	return
}

// StopSharing - stops sharing image, cleans up
func (d *DefaultDistributor) StopSharing(name string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.active[name].Drop()

	err := d.cleanup(name)
	if err != nil {
		return err
	}

	return nil
}

func (d *DefaultDistributor) cleanup(name string) error {
	// removing torrent file
	filename := generateImageName(name)

	err := os.Remove(getTorrentName(filename))
	if err != nil {
		return err
	}

	err = RemoveContents(filename)
	if err != nil {
		return err
	}

	// removing directory as well
	return os.Remove(filename)
}

// RemoveContents - removes contents from directory
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DefaultDistributor) createTorrentFile(name string) (torrent []byte, err error) {
	contents, err := Create(filepath.Join(ImageDownloadPath, name))
	if err != nil {
		return
	}

	f, err := os.Create(filepath.Join(ImageDownloadPath, getTorrentName(name)))
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(contents)
	if err != nil {
		return
	}

	torrent = contents

	return
}
