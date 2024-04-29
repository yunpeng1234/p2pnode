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
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
)

var identifier string = "SnapInnovation"
var baseHostIp string = "0.0.0.0"
var protocolTag string = "/messaging"

var chatLock chan int = make(chan int, 1)

func handleStream(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	if len(chatLock) == 0 {
		fmt.Fprint(os.Stdin, "dummy\n")
		chatLock <- 2
	}
	go readData(rw, stream, chatLock)
	go writeData(rw, stream, chatLock)
}

func readData(rw *bufio.ReadWriter, stream network.Stream, c chan int) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			c <- 1
			return
		}

		if strings.TrimSpace(str) == "quit" || str == "" {
			c <- 1
			stream.Close()
			return
		}

		if str != "\n" {
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter, stream network.Stream, c chan int) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			c <- 1
			panic(err)
		}
		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if strings.TrimSpace(sendData) == "quit" {
			c <- 1
			stream.Close()
			return
		}

		if err != nil {
			fmt.Println("Error writing to buffer")
			c <- 1
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			c <- 1
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

func PrintAvailablePeers(peerChan chan peer.AddrInfo) {
	var peerAddrList []peer.AddrInfo
loop:
	for {
		select {
		case peer := <-peerChan:
			peerAddrList = append(peerAddrList, peer)
		default:
			break loop
		}
	}

	if len(peerAddrList) == 0 {
		fmt.Println("No available channels to connect to")
	} else {
		fmt.Println("Available peers to connect to:")
	loopTwo:
		for _, tempPeer := range peerAddrList {
			fmt.Printf("- %s\n", tempPeer)
			select {
			case peerChan <- tempPeer:
				fmt.Println("Sent")
			default:
				break loopTwo
			}
		}
	}
}

func makeHost(config config, randomness io.Reader) (host.Host, error) {
	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomness)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", baseHostIp, config.listenPort))

	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func startPeerAndConnect(ctx context.Context, h host.Host, destination string) error {
	for _, la := range h.Addrs() {
		log.Printf(" - %v\n", la)
	}
	log.Println()
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// Start a stream with the destination.
	// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
	s, err := h.NewStream(context.Background(), info.ID, protocol.ID(protocolTag))
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Established connection to destination")

	// Create a buffered stream so that read and writes are non-blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go writeData(rw, s, chatLock)
	go readData(rw, s, chatLock)
	return nil
}

func main() {
	// Handle parsing of input arguments
	help := flag.Bool("help", false, "Display Help")
	config := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		fmt.Printf("Usage: \n Run './p2pnode -port [port]'\n")
		os.Exit(0)
	}
	ctx := context.Background()
	r := rand.Reader

	host, err := makeHost(*config, r)

	if err != nil {
		panic(err)
	}

	host.SetStreamHandler(protocol.ID(protocolTag), handleStream)

	peerChan := initMDNS(host, identifier)

	// fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", config.listenHost, config.listenPort, host.ID())
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {

			fmt.Print("Enter ls or connect commands: ")
			input, _ := reader.ReadString('\n')
			select {
			case <-chatLock:
				if <-chatLock == 1 {
					continue
				}
			default:
			}
			input = strings.TrimSpace(input)

			if input == "ls peers -c" {
				PrintPeers(host)
				continue
			}

			if input == "ls peers -a" {
				PrintAvailablePeers(peerChan)
				continue
			}

			if !strings.HasPrefix(input, "connect") {
				fmt.Println("Invalid command, please use the available commands available")
				continue
			} else {
				toConnectID := strings.Split(input, " ")[1]

				err := startPeerAndConnect(ctx, host, toConnectID)
				if err != nil {
					fmt.Println(err)
					continue
				}
				// Create a thread to read and write data.

				if <-chatLock == 1 {
					continue
				}

			}

		}
	}()

	select {}
}
