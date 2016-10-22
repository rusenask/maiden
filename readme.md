# Maiden 

Alternative Docker registry that uses DHT torrents for image distribution

## Usage

Sharing image:

`./maiden -share="odise/busybox-python" -addr="192.168.0.8:50007"`

Options:
* -addr - your advertising address (how other nodes should find you)  


Downloading image (on another node):

    
`./maiden -addr="192.168.0.20:50007" -peers="192.168.0.8:50007"  -pull="odise/busybox-python" -torrent="image-6d0082131056c1f2e879505ca01b6908a09be918.torrent" -seed`

Options:
* -addr - your advertising address (how other nodes should find you)
* -peers - at least one additional peer to connect to
* -pull - image name
* -torrent - torrent file to use for download 

