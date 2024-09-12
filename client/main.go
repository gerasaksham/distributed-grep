package main

import (
	"bufio"
	"errors"
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

func (fs *FileServer) GrepFile(req *string, reply *string) error {
	inputSplit := strings.Split(*req, " ")
	if inputSplit[0] != "grep" {
		return errors.New("non-grep command not supported")
	} else {
		args := inputSplit[1:]
		cmd := exec.Command("grep", args...)
		out, err := cmd.CombinedOutput()
		*reply = string(out)
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				switch exitErr.ExitCode() {
				case 1:
					return nil
				default:
					return fmt.Errorf("grep exited with status %d: %s", exitErr.ExitCode(), *reply)
				}
			}
			return fmt.Errorf("error executing grep: %v", err)
		}
		return nil
	}
}

// write a function that counts the number of lines in a string
func CountLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

func main() {
	var fileServer FileServer
	rpc.Register(&fileServer)
	rpc.HandleHTTP()

	go func() {
		l, e := net.Listen("tcp", ":2232")
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
		client, err = rpc.DialHTTP("tcp", "localhost:2233")
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
		if input != "" {
			var grepReply string
			err = client.Call("FileServer.GrepFile", &input, &grepReply)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(grepReply)
		}
	}
}
