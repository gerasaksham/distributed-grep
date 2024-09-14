package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
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

type GrepReply struct {
	Output    string
	Linecount int
}

func countLines(content string) int {
	return strings.Count(content, "\n")
}

// func findLogFiles() ([]string, error) {
// 	var logFiles []string
// 	entries, err := os.ReadDir(".")
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, entry := range entries {
// 		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
// 			logFiles = append(logFiles, entry.Name())
// 		}
// 	}
// 	return logFiles, nil
// }

func (fs *FileServer) GrepFile(req *string, reply *GrepReply) error {
	inputSplit := strings.Split(*req, " ")
	if inputSplit[0] != "grep" {
		return errors.New("non-grep command not supported")
	} else {
		args := inputSplit[1:]
		// fileNames, err := findLogFiles()
		// if err != nil {
		// 	return fmt.Errorf("error finding log files: %v", err)
		// }
		args = append([]string{"-H"}, args...)
		// args = append(args, fileNames[0]) // Assuming there is only one log file in the folder.
		cmd := exec.Command("grep", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() == 1 {
					return nil
				} else {
					return fmt.Errorf("grep exited with status %d: %s", exitErr.ExitCode(), string(out))
				}
			}
			return fmt.Errorf("error executing grep: %v", err)
		}
		lineCount := countLines(string(out))
		reply.Output = string(out) + "\nNumber of lines: " + fmt.Sprint(lineCount)
		reply.Linecount = lineCount
		return nil
	}
}

func connectAndGrep(serverAddr string, input string, results chan<- GrepReply, wg *sync.WaitGroup) {
	defer wg.Done()

	var client *rpc.Client
	var err error
	client, err = rpc.DialHTTP("tcp", serverAddr)

	if err != nil {
		results <- GrepReply{Output: fmt.Sprintf("Failed to connect to server %s with error: %v", serverAddr, err)}
		return
	}
	defer client.Close()

	// Perform the grep command on the server
	var grepReply GrepReply
	err = client.Call("FileServer.GrepFile", &input, &grepReply)
	if err != nil {
		results <- GrepReply{Output: fmt.Sprintf("Error executing grep on server %s: %v", serverAddr, err)}
		return
	}

	results <- GrepReply{
		Output:    fmt.Sprintf("Server: %s\n%s", serverAddr, grepReply.Output),
		Linecount: grepReply.Linecount,
	}
}

func (fs *FileServer) GrepMultipleServers(req *string, reply *string) error {
	// List of other servers to send grep requests
	servers := []string{
		"localhost:2232", // Second server address
		"localhost:2233", // Third server address
	}

	filenameMap := map[string]string{
		"localhost:2232": "vm2.log",
		"localhost:2233": "vm1.log",
	}

	// Channel to collect results from all servers
	results := make(chan GrepReply, len(servers))

	// WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Spawn a goroutine for each server
	for _, serverAddr := range servers {
		wg.Add(1)
		updatedReq := *req + " " + filenameMap[serverAddr]
		go connectAndGrep(serverAddr, updatedReq, results, &wg)
		// go connectAndGrep(serverAddr, *req, results, &wg)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results from all servers
	var finalOutput strings.Builder
	totalLineCount := 0
	for result := range results {
		finalOutput.WriteString(result.Output + "\n")
		totalLineCount += result.Linecount
	}
	finalOutput.WriteString(fmt.Sprintf("\nTotal number of matching lines: %d", totalLineCount))

	*reply = finalOutput.String()
	return nil
}

func main() {
	var fileServer FileServer
	rpc.Register(&fileServer)
	rpc.HandleHTTP()

	// Run the main server on localhost:2232
	go func() {
		l, e := net.Listen("tcp", ":2233")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}()

	// Wait for the user to input commands
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter grep command: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			var grepReply string

			// Call the GrepMultipleServers function to dispatch requests to other servers
			err := fileServer.GrepMultipleServers(&input, &grepReply)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(grepReply)
			}
		}
	}
}
