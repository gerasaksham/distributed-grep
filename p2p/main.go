package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type FileServer struct{}

type FileRequest struct {
	Filename string
	Content  string
}

type GrepRequest struct {
	Filename string
	Pattern  string
}

func (fs *FileServer) CreateFile(req FileRequest, reply *string) error {
	err := os.WriteFile(req.Filename, []byte(req.Content), 0644)
	if err != nil {
		return err
	}
	*reply = "File created successfully"
	return nil
}

func (fs *FileServer) GrepFile(req GrepRequest, reply *string) error {
	cmd := exec.Command("grep", req.Pattern, req.Filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	*reply = string(output)
	return nil
}

type Peer struct {
	address string
	peers   map[string]*rpc.Client
	mu      sync.Mutex
}

func NewPeer(address string) *Peer {
	return &Peer{
		address: address,
		peers:   make(map[string]*rpc.Client),
	}
}

func (p *Peer) StartServer() error {
	fileServer := new(FileServer)
	rpc.Register(fileServer)

	listener, err := net.Listen("tcp", p.address)
	if err != nil {
		return err
	}

	fmt.Printf("Peer server is running on %s\n", listener.Addr().String())
	go rpc.Accept(listener)
	return nil
}

func (p *Peer) ConnectToPeer(address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.peers[address]; exists {
		return nil // Already connected
	}

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return err
	}

	p.peers[address] = client
	fmt.Printf("Connected to peer: %s\n", address)
	return nil
}

func (p *Peer) CreateFile(peerAddress, filename, content string) error {
	p.mu.Lock()
	client, exists := p.peers[peerAddress]
	p.mu.Unlock()

	if !exists {
		return fmt.Errorf("peer not connected: %s", peerAddress)
	}

	req := FileRequest{Filename: filename, Content: content}
	var reply string
	err := client.Call("FileServer.CreateFile", req, &reply)
	if err != nil {
		return err
	}

	fmt.Println(reply)
	return nil
}

func (p *Peer) GrepFile(peerAddress, filename, pattern string) error {
	p.mu.Lock()
	client, exists := p.peers[peerAddress]
	p.mu.Unlock()

	if !exists {
		return fmt.Errorf("peer not connected: %s", peerAddress)
	}

	req := GrepRequest{Filename: filename, Pattern: pattern}
	var reply string
	err := client.Call("FileServer.GrepFile", req, &reply)
	if err != nil {
		return err
	}

	fmt.Println("Grep result:")
	fmt.Println(reply)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run peer.go <address>")
		return
	}

	address := os.Args[1]
	peer := NewPeer(address)

	err := peer.StartServer()
	if err != nil {
		log.Fatal("Error starting peer server:", err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (connect/create/grep/quit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.Split(input, " ")
		command := parts[0]

		switch command {
		case "connect":
			if len(parts) != 2 {
				fmt.Println("Usage: connect <peer_address>")
				continue
			}
			err := peer.ConnectToPeer(parts[1])
			if err != nil {
				fmt.Println("Error connecting to peer:", err)
			}
		case "create":
			if len(parts) != 4 {
				fmt.Println("Usage: create <peer_address> <filename> <content>")
				continue
			}
			err := peer.CreateFile(parts[1], parts[2], parts[3])
			if err != nil {
				fmt.Println("Error creating file:", err)
			}
		case "grep":
			if len(parts) != 4 {
				fmt.Println("Usage: grep <peer_address> <filename> <pattern>")
				continue
			}
			err := peer.GrepFile(parts[1], parts[2], parts[3])
			if err != nil {
				fmt.Println("Error grepping file:", err)
			}
		case "quit":
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}
