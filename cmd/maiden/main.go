package main

import (
	"flag"
	"os"

	"github.com/rusenask/maiden"

	log "github.com/Sirupsen/logrus"
)

func main() {
	dir := flag.String("dir", "", "directory from which to create a torrent file")
	flag.Parse()

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
