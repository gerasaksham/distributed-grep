package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"
)

type FileServer struct{}

type FileRequest struct {
	Filename string
	Content  string
}

type GrepRequest struct {
	Filename string
	Pattern  string
	Flag     string
}

// func (fs *FileServer) CreateFile(req FileRequest, reply *string) error {
// 	err := os.WriteFile(req.Filename, []byte(req.Content), 0644)
// 	if err != nil {
// 		return err
// 	}
// 	*reply = "File created successfully"
// 	return nil
// }

func (fs *FileServer) GrepFile(req *GrepRequest, reply *string) error {
	cmd2 := exec.Command("grep", req.Flag, req.Pattern, req.Filename)
	out, err := cmd2.Output()
	if err != nil {
		log.Println("GrepFile error:", err)
	}
	*reply = string(out)
	return nil
}

func main() {
	var fileServer FileServer
	rpc.Register(&fileServer)
	rpc.HandleHTTP()

	go func() {
		l, e := net.Listen("tcp", ":2233")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}()

	// filename := "test.txt"
	// content := "Hello you beautiful human beings from the server"

	// Create file on the server
	// time.Sleep(1 * time.Second)

	const maxRetries = 5
	const initialDelay = 1 * time.Second

	var client *rpc.Client
	var err error

	for i := 0; i < maxRetries; i++ {
		client, err = rpc.DialHTTP("tcp", "localhost:2232")
		if err == nil {
			break
		}

		log.Printf("dialing failed: %v; retrying in %v", err, initialDelay)
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * initialDelay)
	}

	if err != nil {
		log.Fatal("dialing failed after retries:", err)
	}
	// defer client.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		fmt.Println("Input:", input)
		if input != "" {
			input_split := strings.Split(input, " ")
			grepCommand := input_split[0]
			flag := input_split[1]
			pattern := input_split[2]
			filename := input_split[3]
			if grepCommand != "grep" {
				fmt.Println("Invalid command")
				continue
			}

			grepReq := GrepRequest{Filename: filename, Pattern: pattern, Flag: flag}
			var grepReply string
			err = client.Call("FileServer.GrepFile", &grepReq, &grepReply)
			if err != nil {
				fmt.Println("Error running grep:", err)
				return
			}
			fmt.Println(grepReply)
		}
	}
}
