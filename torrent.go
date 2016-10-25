package maiden

import (
	"crypto/sha1"
	"fmt"
	"golang.org/x/time/rate"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"

	log "github.com/Sirupsen/logrus"
)

// path to the image itself
func imagePath(filename string) string {
	return filepath.Join(filename, filename)
}

// since most of the names are of format `org/name` - there can be problems
// generating paths
func generateImageName(name string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(name)))
}

func getTorrentName(name string) string {
	return fmt.Sprintf("image-%s.torrent", name)
}

func (d *DefaultDistributor) getTorrentClient() (*torrent.Client, error) {
	var clientConfig torrent.Config
	if d.cfg.MMap {
		clientConfig.DefaultStorage = storage.NewMMap("")
	}
	if d.cfg.Addr != nil {
		clientConfig.ListenAddr = d.cfg.Addr.String()
	}

	clientConfig.DataDir = filepath.Join(ImageDownloadPath)
	clientConfig.Debug = true

	clientConfig.Seed = true

	if d.cfg.UploadRate != -1 {
		clientConfig.UploadRateLimiter = rate.NewLimiter(rate.Limit(d.cfg.UploadRate), 256<<10)
	}
	if d.cfg.DownloadRate != -1 {
		clientConfig.DownloadRateLimiter = rate.NewLimiter(rate.Limit(d.cfg.DownloadRate), 1<<20)
	}

	return torrent.NewClient(&clientConfig)
}

// AddTorrent - adds given torrent for download/seeding
func (d *DefaultDistributor) addTorrent(name, arg string) error {
	// uiprogress.Start()
	t, err := func() (*torrent.Torrent, error) {
		if strings.HasPrefix(arg, "magnet:") {
			t, err := d.tClinet.AddMagnet(arg)
			if err != nil {
				log.Fatalf("error adding magnet: %s", err)
			}
			return t, nil
		} else {
			metaInfo, err := metainfo.LoadFromFile(arg)
			if err != nil {
				log.WithFields(log.Fields{
					"torrent": arg,
					"error":   err,
				}).Fatal("error loading torrent file")
				return nil, err
			}
			t, err := d.tClinet.AddTorrent(metaInfo)
			if err != nil {
				log.Fatal(err)
			}
			return t, nil
		}
	}()

	if err != nil {
		return err
	}

	// starting to check torrent info
	go torrentInfo(t)

	t.AddPeers(func() (ret []torrent.Peer) {
		for _, ta := range d.cfg.Peers {
			ret = append(ret, torrent.Peer{
				IP:   ta.IP,
				Port: ta.Port,
			})
		}
		return
	}())
	go func() {
		<-t.GotInfo()
		t.DownloadAll()

		d.mutex.Lock()

		// checking whether there was existing torrent with the same name
		existing, ok := d.active[name]
		if ok {
			existing.Drop()
		}
		// after drop - replacing it
		d.active[name] = t
		d.mutex.Unlock()
	}()

	if d.tClinet.WaitAll() {
		log.Print("downloaded ALL the torrents")
	} else {
		log.Error("y u no complete torrents?!")
	}
	return nil
}

func torrentInfo(t *torrent.Torrent) {
	for _ = range time.Tick(5 * time.Second) {
		select {
		case <-t.GotInfo():
			if t.BytesCompleted() == t.Info().TotalLength() {
				log.WithFields(log.Fields{
					"name": t.Name(),
				}).Info("completed")
				// sucess, stop tracking
				return
			}
		}

	}
}
