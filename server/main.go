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

type Args struct {
	Str string
}

type StringArgument struct{}

func (s *StringArgument) ReturnCapitalizedString(args *Args, reply *string) error {
	log.Println("serverfile:", args.Str)
	*reply = strings.ToUpper(args.Str)
	return nil
}

func main() {
	var strarg StringArgument
	rpc.Register(&strarg)
	rpc.HandleHTTP()

	go func() {
		l, e := net.Listen("tcp", ":2233")
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
			client, err := rpc.DialHTTP("tcp", "localhost:2232")
			if err != nil {
				log.Fatal("dialing:", err)
			}
			var reply string
			args := Args{Str: input}
			err = client.Call("StringArgument.ReturnCapitalizedString", &args, &reply)
			if err != nil {
				log.Fatal("rpc error:", err)
			}
			fmt.Println("Capitalized string is:", reply)
		}
	}
}
