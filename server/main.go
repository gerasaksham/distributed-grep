package main

import (
	//"bufio"
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	//"log"
	"net"
	//"net/http"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	//"strings"
)

type Args struct {
	Str string
}

type StringArgument struct{}

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

//	func (s *StringArgument) ReturnCapitalizedString(args *Args, reply *string) error {
//		log.Println("serverfile:", args.Str)
//		*reply = strings.ToUpper(args.Str)
//		return nil
//	}
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
// 	fmt.Println("cmd:", cmd)
// 	cmd2 := exec.Command("grep", args.Str)
// 	cmd2.Stdin, _ = cmd.StdoutPipe()
// 	fmt.Println("cmd2:", cmd2)
// 	out, err := cmd2.Output()
// 	fmt.Println("out:", out)
// 	if err != nil {
// 		log.Println("error:", err)
// 	}
// 	*reply = string(out)
// 	return nil
// }

func main() {
	fileServer := new(FileServer)
	rpc.Register(fileServer)
	rpc.HandleHTTP()
	go func() {
		l, e := net.Listen("tcp", ":2233")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}()
	// add a timeout here to give the server time to start
	time.Sleep(1 * time.Second)
	client, err := rpc.Dial("tcp", "localhost:2232")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()
	filename := "test.txt"
	content := "Hello you beautiful human beings from the server"

	// Create file on the server
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

			fmt.Println("Grep result from client:")
			fmt.Println(grepReply)
		}
	}
}
