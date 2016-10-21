package main

import (
	"flag"
	"os"

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

func main() {
	dir := flag.String("dir", "", "directory from which to create a torrent file")
	share := flag.String("share", "", "share selected image to your peers")
	flag.Parse()

	if *dir != "" {
		contents, err := maiden.Create(*dir)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("failed to create data torrent")
		}

		f, err := os.Create("data.torrent")
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("failed to create data torrent")
		}
		defer f.Close()

		f.Write(contents)
		log.Infof("torrent for `%s` created", dir)
	}

	if *share != "" {
		endpoint := "unix:///var/run/docker.sock"
		if os.Getenv(EnvDockerEndpoint) != "" {
			endpoint = os.Getenv(EnvDockerEndpoint)
		}

		client, err := docker.NewClient(endpoint)
		// client, err := docker.NewClientFromEnv()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("failed to get docker client from env")
		}

		distributor := maiden.NewDefaultDistributor(client)
		distributor.ShareImage(*share)
	}

}
