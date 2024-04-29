package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/multiformats/go-multiaddr"
)

var mainHost *host.Host = nil

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

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}

		if str != "\n" {
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeDataMain(rw *bufio.ReadWriter, h *host.Host) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}
		// Intercept the command to print peerStore of current host
		if strings.TrimSpace(sendData) == "list peers" {
			PrintPeers(*h)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func PrintPeers(h host.Host) {
	peers := h.Peerstore().Peers()
	fmt.Println("Connected Peers:")
	for _, peer := range peers[1:] {
		fmt.Printf("- %s\n", peer)
	}
}

func makeHost(config config, randomness io.Reader) (host.Host, error) {
	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomness)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", config.listenHost, config.listenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func main() {
	help := flag.Bool("help", false, "Display Help")
	config := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		fmt.Printf("Usage: \n Run './p2pnode -port [port]'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on port: %d\n", config.listenPort)

	ctx := context.Background()
	r := rand.Reader

	host, err := makeHost(*config, r)

	if err != nil {
		panic(err)
	}

	if config.main {
		mainHost = &host
		host.SetStreamHandler(protocol.ID(config.protocol), handleHostStream)
	} else {
		host.SetStreamHandler(protocol.ID(config.protocol), handleStream)
	}
	// fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", config.listenHost, config.listenPort, host.ID())

	peerChan := initMDNS(host, config.identifier)
	for { // allows multiple peers to join
		peer := <-peerChan // will block until we discover a peer
		if !config.main {
			// only main is in charge of connecting
			continue
		}

		if err := host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}

		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.protocol))

		if err != nil {
			fmt.Println("Stream open failed", err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeDataMain(rw, &host)
			go readData(rw)
			fmt.Println("Connected to:", peer)
		}
	}
}
