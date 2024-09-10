package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
)

// type Args struct {
// 	Str string
// }

type FileRequest struct {
	Filename string
	Content  string
}

type GrepRequest struct {
	Filename string
	Pattern  string
}

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

func main() {

	go func() {
		l, e := net.Listen("tcp", ":2232")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			client, err := rpc.Dial("tcp", "localhost:2233")
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

			// Run grep command on the server
			grepReq := GrepRequest{Filename: filename, Pattern: "Hello"}
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
