package main

import (
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"testing"
)

var once sync.Once

func setupServer() {
	// Setup the FileServer
	fileServer := &FileServer{}
	rpc.Register(fileServer)
	once.Do(func() {
		rpc.HandleHTTP()
	})
	go func() {
		l, err := net.Listen("tcp", ":2232")
		if err != nil {
			panic(err)
		}
		http.Serve(l, nil)
	}()
}

func TestGrepFile(t *testing.T) {
	// Setup a temporary log file
	logFileName := "test1.log"
	logContent := "This is a test log file\nwith multiple lines\nand some test content\n"
	err := os.WriteFile(logFileName, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	defer os.Remove(logFileName)

	// Setup the server
	setupServer()

	// Create an RPC client
	client, err := rpc.DialHTTP("tcp", "localhost:2232")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	// Test the GrepFile method
	input := "grep test test1.log"
	var grepReply GrepReply
	err = client.Call("FileServer.GrepFile", &input, &grepReply)
	if err != nil {
		t.Fatalf("GrepFile RPC call failed: %v", err)
	}

	expectedOutput := "test1.log:This is a test log file\n" +
		"test1.log:and some test content\n\n" +
		"Number of lines: 2"
	if strings.TrimSpace(grepReply.Output) != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, grepReply.Output)
	}
	if grepReply.Linecount != 2 {
		t.Errorf("Expected line count: 2, got: %d", grepReply.Linecount)
	}
}

func TestGrepMultipleServers(t *testing.T) {
	// Setup temporary log files for multiple servers
	logFileName1 := "test1.log"
	logContent1 := "This is a test log file for test1\nwith multiple lines\nand some test content\n"
	err := os.WriteFile(logFileName1, []byte(logContent1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file for test1: %v", err)
	}
	defer os.Remove(logFileName1)

	logFileName2 := "test2.log"
	logContent2 := "This is a test log file for test2\nwith multiple lines\nand some test content\n"
	err = os.WriteFile(logFileName2, []byte(logContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file for test2: %v", err)
	}
	defer os.Remove(logFileName2)

	// Setup the server
	setupServer()

	// Test the GrepMultipleServers method
	input := "grep test"
	var grepReply string
	err = (&FileServer{}).GrepMultipleServers(&input, &grepReply)
	if err != nil {
		t.Fatalf("GrepMultipleServers call failed: %v", err)
	}

	expectedOutput := "Server: localhost:2232\ntest2.log:This is a test log file for test2\n" +
		"test2.log:and some test content\n\n" +
		"Number of lines: 2\n\n" +
		"Server: localhost:2233\ntest1.log:This is a test log file for test1\n" +
		"test1.log:and some test content\n\n" +
		"Number of lines: 2\n\n" +
		"Total number of matching lines: 4"
	if strings.TrimSpace(grepReply) != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, grepReply)
	}
}

func TestMain(m *testing.M) {
	// Setup the server
	setupServer()

	// Run tests
	os.Exit(m.Run())
}
