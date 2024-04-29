# p2p chat app with libp2p [support peer discovery using mdns]

This program demonstrates a simple p2p chat application.

## How to build this example?

```
go mod tidy
go build
```

## Usage

Use two different terminal windows to run

```
./message -port 4000 -main
./meesage -port 4001
```


## So how does it work?


1. **Set a default handler function for incoming connections.**

This function is called on the local peer when a remote peer initiate a connection and starts a stream with the local peer.
```go
// Set a function as stream handler.
host.SetStreamHandler("/chat/1.1.0", handleStream)
```

```handleStream``` is executed for each new stream incoming to the local peer. ```stream``` is used to exchange data between local and remote peer. This example uses non blocking functions for reading and writing from this stream.

One ```handleStream``` is for the host to intercept ```list peers``` commands to print it's peer


```go
func handleHostStream(stream network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)
	go writeDataMain(rw, mainHost)
}

func handleStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)
	go writeData(rw)
}
```

