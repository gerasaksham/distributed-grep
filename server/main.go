package main

import (
	//"bufio"
	"fmt"
	//"log"
	"net"
	//"net/http"
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

	listener, err := net.Listen("tcp", ":2233") // Use port 0 to get a random available port
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server is running on %s\n", listener.Addr().String())
	rpc.Accept(listener)
}
