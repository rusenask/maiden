package maiden

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

var (
	builtinAnnounceList = [][]string{}
)

// example usage: ./create example

// Create - creates torrent file from given directory
func Create(dir string) (torrent []byte, err error) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	mi := metainfo.MetaInfo{
		AnnounceList: builtinAnnounceList,
	}
	mi.SetDefaults()
	info := metainfo.Info{
		PieceLength: 256 * 1024,
	}
	// err := info.BuildFromFilePath(args.Root)
	err = info.BuildFromFilePath(dir)
	if err != nil {
		log.Fatal(err)
	}
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	err = mi.Write(buf)
	if err != nil {
		return
	}
	fmt.Println("infohash:")
	fmt.Printf("%s\n", mi.HashInfoBytes().HexString())
	fmt.Fprintf(os.Stdout, "%s\n", mi.Magnet(info.Name, mi.HashInfoBytes()).String())

	return buf.Bytes(), nil
}
