package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"
)

// type Args struct {
// 	Str string
// }

//type StringArgument struct{}

// func (s *StringArgument) ReturnCapitalizedString(args *Args, reply *string) error {
// 	log.Println("clientfile:", args.Str)
// 	*reply = strings.ToUpper(args.Str)
// 	return nil
// }

// func (s *StringArgument) GrepString(args *Args, reply *string) error {
// 	// create a new file on the server named "clientfile.txt" and write "Hellop World" to it
// 	file, err := os.Create("serverfile.txt")
// 	if err != nil {
// 		log.Fatal("error creating file:", err)
// 	}
// 	defer file.Close()
// 	file.WriteString("Hello World from server")
// 	// cat the file then grep the string from the file
// 	cmd := exec.Command("cat", file.Name())
// 	cmd2 := exec.Command("grep", args.Str)
// 	cmd2.Stdin, _ = cmd.StdoutPipe()
// 	out, err := cmd2.Output()
// 	if err != nil {
// 		log.Println("error:", err)
// 	}
// 	*reply = string(out)
// 	return nil
// }

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

func main() {
	fileServer := new(FileServer)
	rpc.Register(fileServer)

	go func() {
		l, e := net.Listen("tcp", ":2232")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}()

	filename := "test.txt"
	content := "Hello you beautiful human beings from the server"

	// Create file on the server
	time.Sleep(1 * time.Second)
	client, err := rpc.Dial("tcp", "localhost:2233")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()
	createReq := FileRequest{Filename: filename, Content: content}
	var createReply string
	err = client.Call("FileServer.CreateFile", createReq, &createReply)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	fmt.Println(createReply)

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {

			// Run grep command on the server
			grepReq := GrepRequest{Filename: filename, Pattern: input}
			var grepReply string
			err = client.Call("FileServer.GrepFile", grepReq, &grepReply)
			if err != nil {
				fmt.Println("Error running grep:", err)
				return
			}

			fmt.Println("Grep result:")
			fmt.Println(grepReply)
		}
	}
}
