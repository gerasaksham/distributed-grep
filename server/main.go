package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"

	//"net"
	//"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
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
					return nil // grep command found no matches, which is not a fatal error
				default:
					return fmt.Errorf("grep exited with status %d: %s", exitErr.ExitCode(), *reply)
				}
			}
			return fmt.Errorf("error executing grep: %v", err)
		}
		return nil
	}
}

func CountLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

func connectAndGrep(serverAddr string, input string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	const maxRetries = 5
	const initialDelay = 1 * time.Second

	var client *rpc.Client
	var err error

	// Retry mechanism to connect to the server
	for i := 0; i < maxRetries; i++ {
		client, err = rpc.DialHTTP("tcp", serverAddr)
		if err == nil {
			break
		}
		log.Printf("dialing to server %s failed: %v; retrying in %v", serverAddr, err, initialDelay)
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * initialDelay)
	}

	if err != nil {
		results <- fmt.Sprintf("Failed to connect to server %s after retries: %v", serverAddr, err)
		return
	}
	defer client.Close()

	// Perform the grep command on the server
	var grepReply string
	err = client.Call("FileServer.GrepFile", &input, &grepReply)
	if err != nil {
		results <- fmt.Sprintf("Error executing grep on server %s: %v", serverAddr, err)
		return
	}

	results <- fmt.Sprintf("Server: %s\n%s", serverAddr, grepReply)
}

func main() {
	// Example list of servers
	servers := []string{
		"localhost:2232", // First server address
		"localhost:2233",
		"localhost:2234",
		"localhost:2235", // Second server address
		// Add more server addresses as needed
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter grep command: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Channel to collect results from all servers
		results := make(chan string, len(servers))

		// WaitGroup to synchronize goroutines
		var wg sync.WaitGroup

		// Spawn a goroutine for each server
		for _, serverAddr := range servers {
			wg.Add(1)
			go connectAndGrep(serverAddr, input, results, &wg)
		}

		// Wait for all goroutines to finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect and print results from all servers
		for result := range results {

			fmt.Println(result)
		}
	}
}
