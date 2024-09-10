package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

func (fs *FileServer) CreateFile(req FileRequest, reply *string) error {
	err := os.WriteFile(req.Filename, []byte(req.Content), 0644)
	if err != nil {
		return err
	}
	*reply = "File created successfully"
	return nil
}

func (fs *FileServer) GrepFile(req GrepRequest, reply *string) error {

	// cmd2 := exec.Command(req.Command, req.Filename)
	cmd2 := exec.Command("grep", req.Flag, req.Pattern, req.Filename)
	out, err := cmd2.Output()
	if err != nil {
		log.Println("GrepFile error:", err)
	}
	*reply = string(out)
	return nil
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
	var ans string
	fileServer := new(FileServer)
	req := GrepRequest{Filename: "test.txt", Pattern: "Hello", Flag: "-i"}
	fileServer.GrepFile(req, &ans)
	lines := CountLines(ans)
	ans = fmt.Sprintf("The number of lines in the file is: %d\n The Lines are:\n%s", lines, ans)
	print(ans)

	// rpc.Register(fileServer)

	// go func() {
	// 	l, e := net.Listen("tcp", ":2232")
	// 	if e != nil {
	// 		log.Fatal("listen error:", e)
	// 	}
	// 	http.Serve(l, nil)
	// }()

	// filename := "test.txt"
	// content := "Hello you beautiful human beings from the server"

	// // Create file on the server
	// time.Sleep(1 * time.Second)
	// client, err := rpc.Dial("tcp", "localhost:2233")
	// if err != nil {
	// 	log.Fatal("dialing:", err)
	// }
	// defer client.Close()
	// createReq := FileRequest{Filename: filename, Content: content}
	// var createReply string
	// err = client.Call("FileServer.CreateFile", createReq, &createReply)
	// if err != nil {
	// 	fmt.Println("Error creating file:", err)
	// 	return
	// }
	// fmt.Println(createReply)

	// reader := bufio.NewReader(os.Stdin)
	// for {
	// 	input, _ := reader.ReadString('\n')
	// 	input = strings.TrimSpace(input)
	// 	if input != "" {

	// 		// Run grep command on the server
	// 		grepReq := GrepRequest{Filename: filename, Pattern: input}
	// 		var grepReply string
	// 		err = client.Call("FileServer.GrepFile", grepReq, &grepReply)
	// 		if err != nil {
	// 			fmt.Println("Error running grep:", err)
	// 			return
	// 		}

	// 		fmt.Println("Grep result:")
	// 		fmt.Println(grepReply)
	// 	}
	// }
}
