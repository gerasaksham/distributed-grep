package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type FileServer struct{}

// FileRequest is the struct that holds the filename and content of the file
type FileRequest struct {
	Filename string
	Content  string
}

// GrepReply is the struct that holds the output of the grep command
type GrepReply struct {
	Output    string
	Linecount int
}

// countLines counts the number of lines in a string
func countLines(content string) int {
	return strings.Count(content, "\n")
}

// WriteFile writes the content to a file with the given filename
func (fs *FileServer) WriteFile(req *FileRequest, reply *string) error {
	file, err := os.Create(req.Filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(req.Content)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	return nil
}

// This method is used to grep a file for a given pattern
func (fs *FileServer) GrepFile(req *string, reply *GrepReply) error {
	inputSplit := strings.Split(*req, " ")
	if inputSplit[0] != "grep" {
		return errors.New("non-grep command not supported")
	} else {
		args := inputSplit[1:]
		args = append([]string{"-H"}, args...) // To print the filename along with the matching line
		cmd := exec.Command("grep", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() == 1 {
					return nil // status code = 1 when no match found
				} else {
					return fmt.Errorf("grep exited with status %d: %s", exitErr.ExitCode(), string(out))
				}
			}
			return fmt.Errorf("error executing grep: %v", err)
		}
		lineCount := countLines(string(out))
		reply.Output = string(out) + "\nNumber of lines: " + fmt.Sprint(lineCount) // Appending lineCount
		reply.Linecount = lineCount
		return nil
	}
}

// This method connects to the server and executes the grep command
func connectAndGrep(serverAddr string, input string, results chan<- GrepReply, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.DialTimeout("tcp", serverAddr, 2*time.Second) // 2 second timeout for unresponsive servers
	if err != nil {
		results <- GrepReply{Output: fmt.Sprintf("Failed to connect to server %s with error: %v", serverAddr, err)}
		return
	}
	client := rpc.NewClient(conn)
	defer client.Close()

	var grepReply GrepReply
	err = client.Call("FileServer.GrepFile", &input, &grepReply) // Calling the RPC
	if err != nil {
		results <- GrepReply{Output: fmt.Sprintf("Error executing grep on server %s: %v", serverAddr, err)} // Send the error to the channel
		return
	}

	results <- GrepReply{
		Output:    fmt.Sprintf("Server: %s\n%s", serverAddr, grepReply.Output),
		Linecount: grepReply.Linecount,
	}
}

// Method to concurrently grep multiple servers
func (fs *FileServer) GrepMultipleServers(req *string, filenameMap *map[string]string, reply *string) (error, int) {
	// List of servers addresses
	servers := []string{
		"fa24-cs425-3101.cs.illinois.edu:2232",
		"fa24-cs425-3102.cs.illinois.edu:2232",
		"fa24-cs425-3103.cs.illinois.edu:2232",
		"fa24-cs425-3104.cs.illinois.edu:2232",
		"fa24-cs425-3105.cs.illinois.edu:2232",
		"fa24-cs425-3106.cs.illinois.edu:2232",
		"fa24-cs425-3107.cs.illinois.edu:2232",
		"fa24-cs425-3108.cs.illinois.edu:2232",
		"fa24-cs425-3109.cs.illinois.edu:2232",
		"fa24-cs425-3110.cs.illinois.edu:2232",
	}

	results := make(chan GrepReply, len(servers)) // Channel to store the results
	var wg sync.WaitGroup                         // WaitGroup to wait for all goroutines to finish

	for _, serverAddr := range servers {
		wg.Add(1)
		updatedReq := *req + " " + (*filenameMap)[serverAddr]
		go connectAndGrep(serverAddr, updatedReq, results, &wg) // Separate goroutine for each server
	}

	go func() {
		wg.Wait() // Waiting for all goroutines to finish
		close(results)
	}()

	var finalOutput strings.Builder
	totalLineCount := 0
	for result := range results {
		finalOutput.WriteString(result.Output + "\n") // Concatenating the outputs from all servers
		totalLineCount += result.Linecount
	}
	finalOutput.WriteString(fmt.Sprintf("\nTotal number of matching lines: %d", totalLineCount))

	*reply = finalOutput.String()
	return nil, totalLineCount
}

func main() {
	var fileServer FileServer
	rpc.Register(&fileServer)

	go func() {
		l, e := net.Listen("tcp", ":2232") // Listening for grep requests
		if e != nil {
			log.Fatal("listen error:", e)
		}
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal("accept error:", err)
			}
			go rpc.ServeConn(conn)
		}
	}()

	// Mapping of server addresses to log filenames
	filenameMap := map[string]string{
		"fa24-cs425-3101.cs.illinois.edu:2232": "vm1.log",
		"fa24-cs425-3102.cs.illinois.edu:2232": "vm2.log",
		"fa24-cs425-3103.cs.illinois.edu:2232": "vm3.log",
		"fa24-cs425-3104.cs.illinois.edu:2232": "vm4.log",
		"fa24-cs425-3105.cs.illinois.edu:2232": "vm5.log",
		"fa24-cs425-3106.cs.illinois.edu:2232": "vm6.log",
		"fa24-cs425-3107.cs.illinois.edu:2232": "vm7.log",
		"fa24-cs425-3108.cs.illinois.edu:2232": "vm8.log",
		"fa24-cs425-3109.cs.illinois.edu:2232": "vm9.log",
		"fa24-cs425-3110.cs.illinois.edu:2232": "vm10.log",
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter grep command: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			var grepReply string
			startTime := time.Now() // To measure the latency
			err, _ := fileServer.GrepMultipleServers(&input, &filenameMap, &grepReply)
			elapsedTime := time.Since(startTime)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(grepReply)
				fmt.Println("Time taken:", elapsedTime)
			}
		}
	}
}
