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
)

type Args struct {
	Str  string
	File string
}

type StringArgument struct{}

func (s *StringArgument) ReturnCapitalizedString(args *Args, reply *string) error {
	log.Println("clientfile:", args.Str)
	*reply = strings.ToUpper(args.Str)
	return nil
}

func (s *StringArgument) GrepString(args *Args, reply *string) error {
	cmd := exec.Command("grep", args.Str, args.File)
	out, err := cmd.Output()
	if err != nil {
		log.Println("error:", err)
	}
	*reply = string(out)
	return nil
}

func main() {
	var strarg StringArgument
	rpc.Register(&strarg)
	rpc.HandleHTTP()

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
			client, err := rpc.DialHTTP("tcp", "localhost:2233")
			if err != nil {
				log.Fatal("dialing:", err)
			}
			var reply string
			args := Args{Str: input}
			err = client.Call("StringArgument.GrepString", &args, &reply)
			if err != nil {
				log.Fatal("rpc error:", err)
			}
			fmt.Println("Capitalized string is:", reply)
		}
	}
}
