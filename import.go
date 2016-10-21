package maiden

import (
	"io"

	docker "github.com/fsouza/go-dockerclient"
)

func (d *DefaultDistributor) importImage(input io.Reader) error {
	opts := docker.LoadImageOptions{
		InputStream: input,
	}

	return d.dClient.LoadImage(opts)
}
