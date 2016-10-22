package main

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"os/signal"

	"github.com/anacrolix/tagflag"
	"github.com/rusenask/maiden"

	log "github.com/Sirupsen/logrus"

	docker "github.com/fsouza/go-dockerclient"
)

// docker required params
const (
	EnvDockerEndpoint = "DOCKER_ENDPOINT"
	// EnvDockerRegistryEmail = "DOCKER_REGISTRY_EMAIL"
	// EnvDockerRegistryAuth  = "DOCKER_REGISTRY_AUTH"
)

var flags = struct {
	Mmap  bool           `help:"memory-map torrent data"`
	Peers []*net.TCPAddr `help:"addresses of some starting peers"`
	Seed  bool           `help:"seed after download is complete"`

	Clean bool `help:"cleanup after download/upload"`

	Share string `help:"image name that should be shared"`

	Pull    string `help:"image name that should be downloaded"`
	Torrent string `help:"torrent file location"`

	Addr         *net.TCPAddr `help:"network listen addr"`
	UploadRate   int64        `help:"max piece bytes to send per second"`
	DownloadRate int64        `help:"max bytes per second down from peers"`
	// tagflag.StartPos
	// Share string `help:"image ID that should be shared"`
}{
	UploadRate:   -1,
	DownloadRate: -1,
}

func main() {
	tagflag.Parse(&flags)
	endpoint := "unix:///var/run/docker.sock"
	if os.Getenv(EnvDockerEndpoint) != "" {
		endpoint = os.Getenv(EnvDockerEndpoint)
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to get docker client from env")
	}

	config := &maiden.DHTDistributorConfig{
		MMap:         true,
		Peers:        flags.Peers,
		Addr:         flags.Addr,
		UploadRate:   flags.UploadRate,
		DownloadRate: flags.DownloadRate,
	}

	distributor, err := maiden.NewDHTDistributor(config, client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err,
			"listen_addr": config.Addr,
			"peers":       config.Peers,
		}).Fatal("failed to create DHT distributor")
	}
	if flags.Share != "" {
		_, err = distributor.ShareImage(flags.Share)
		if err != nil {
			log.Error(err)
		}

	}

	// pulling some image through torrent
	if flags.Pull != "" && flags.Torrent != "" {
		f, err := os.Open(flags.Torrent)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		torrent, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}

		err = distributor.PullImage(flags.Pull, torrent)
		if err != nil {
			log.Error(err)
		}
	}

	// if image is being shared - always seed
	if flags.Seed || flags.Share != "" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go distributor.Serve(ctx)

		signalChan := make(chan os.Signal, 1)
		cleanupDone := make(chan bool)
		signal.Notify(signalChan, os.Interrupt)
		go func() {
			for _ = range signalChan {
				log.Info("\nReceived an interrupt, closing connection...\n\n")
				err := distributor.Shutdown()
				if err != nil {
					log.Fatalf("error while shutting down distributor: %s", err)
				}
				cleanupDone <- true
			}
		}()
		<-cleanupDone
	}

	if flags.Clean {
		err = distributor.Cleanup()
		if err != nil {
			log.Fatal(err)
		}
	}

}
