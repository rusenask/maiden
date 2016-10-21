package maiden

import (
	"golang.org/x/time/rate"

	"github.com/anacrolix/torrent"
	// "github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

func (d *DefaultDistributor) getTorrentClient() (*torrent.Client, error) {
	var clientConfig torrent.Config
	if d.cfg.mMap {
		// clientConfig.DefaultStorage = storage.NewMMap("")
		clientConfig.DefaultStorage = storage.NewBoltDB(".storage.db")
	}
	if d.cfg.addr != nil {
		clientConfig.ListenAddr = d.cfg.addr.String()
	}

	clientConfig.Seed = true

	if d.cfg.uploadRate != -1 {
		clientConfig.UploadRateLimiter = rate.NewLimiter(rate.Limit(d.cfg.uploadRate), 256<<10)
	}
	if d.cfg.downloadRate != -1 {
		clientConfig.DownloadRateLimiter = rate.NewLimiter(rate.Limit(d.cfg.downloadRate), 1<<20)
	}

	return torrent.NewClient(&clientConfig)
}

// AddTorrent - adds given torrent for download/seeding
func (d *DefaultDistributor) addTorrent(name string) error {
	return nil
}
