package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"testing"
	"time"
)

var logPatterns = map[string]string{
	"INFO":     "Func executed successfully", // Frequent Pattern
	"ERROR":    "Func failed",                // Rare Pattern
	"DEBUG":    "Debugging info",
	"CRITICAL": "Critical error", // One log
	"WARNING":  "Warning",        // Somewhat Frequent
}

var fileMap = map[string]string{
	"server1": "vm1_test.log",
	// "server2":  "vm2_test.log",
	// "server3":  "vm3_test.log",
	// "server4":  "vm4_test.log",
	// "server5":  "vm5_test.log",
	// "server6":  "vm6_test.log",
	// "server7":  "vm7_test.log",
	// "server8":  "vm8_test.log",
	// "server9":  "vm9_test.log",
	"server10": "vm10_test.log",
}

func createPattern(logType string, tmp int) string {
	content := ""
	for i := 0; i < tmp; i++ {
		content += fmt.Sprintf("%s: %s\n", logType, logPatterns[logType])
	}
	return content
}

func createTestLog() (string, map[string]int) {
	rand.Seed(time.Now().UnixNano())
	content := ""
	counts := make(map[string]int)

	// frequent pattern
	tmp := rand.Intn(1200-1000+1) + 1000
	content += createPattern("INFO", tmp)
	counts["INFO"] = tmp
	tmp = rand.Intn(1200-1000+1) + 1000
	content += createPattern("DEBUG", tmp)
	counts["DEBUG"] = tmp

	// Rare Pattern
	tmp = rand.Intn(100)
	if tmp > 80 {
		content += createPattern("ERROR", 100-tmp)
		counts["ERROR"] = 100 - tmp
	} else {
		counts["ERROR"] = 0
	}

	// Somwhat Frequent
	tmp = rand.Intn(300-200+1) + 200
	content += createPattern("WARNING", tmp)
	counts["WARNING"] = tmp

	return content, counts

}

func createTestLogFiles() map[string]int {
	var client *rpc.Client
	var err error
	var reply string

	fileCountMap := map[string]int{
		"INFO":     0,
		"ERROR":    0,
		"DEBUG":    0,
		"CRITICAL": 0,
		"WARNING":  0,
	}
	for server, filename := range fileMap {
		content, countMap := createTestLog()
		if server == "server10" {
			content += createPattern("CRITICAL", 1)
			countMap["CRITICAL"] = 1
		}

		client, err = rpc.DialHTTP("tcp", server)
		if err != nil {
			fmt.Printf("Failed to connect to server %s with error: %v\n", server, err)
			continue
		}
		defer client.Close()

		err = client.Call("FileServer.WriteFile", &FileRequest{Filename: filename, Content: content}, &reply)
		fmt.Printf("Created test log file for %s\n", server)

		for k, v := range countMap {
			fileCountMap[k] += v
		}
	}

	return fileCountMap
}

func TestGrepMultipleServers(t *testing.T) {
	fileCountMap := createTestLogFiles()
	var fileServer FileServer
	var input string
	var reply string

	// Test for INFO
	input = "grep INFO"
	err, totalLineCount := fileServer.GrepMultipleServers(&input, &fileMap, &reply)
	if err != nil {
		fmt.Println("Error:", err)
	}
	if totalLineCount != fileCountMap["INFO"] {
		t.Errorf("Expected %d, got %d", fileCountMap["INFO"], totalLineCount)
	}
	fmt.Println("INFO:", totalLineCount)

	// Test for ERROR
	input = "grep ERROR"
	err, totalLineCount = fileServer.GrepMultipleServers(&input, &fileMap, &reply)
	if err != nil {
		fmt.Println("Error:", err)
	}
	if totalLineCount != fileCountMap["ERROR"] {
		t.Errorf("Expected %d, got %d", fileCountMap["ERROR"], totalLineCount)
	}
	fmt.Println("ERROR:", totalLineCount)

	// Test for DEBUG
	input = "grep DEBUG"
	err, totalLineCount = fileServer.GrepMultipleServers(&input, &fileMap, &reply)
	if err != nil {
		fmt.Println("Error:", err)
	}
	if totalLineCount != fileCountMap["DEBUG"] {
		t.Errorf("Expected %d, got %d", fileCountMap["DEBUG"], totalLineCount)
	}
	fmt.Println("DEBUG:", totalLineCount)

	// Test for WARNING
	input = "grep WARNING"
	err, totalLineCount = fileServer.GrepMultipleServers(&input, &fileMap, &reply)
	if err != nil {
		fmt.Println("Error:", err)
	}
	if totalLineCount != fileCountMap["WARNING"] {
		t.Errorf("Expected %d, got %d", fileCountMap["WARNING"], totalLineCount)
	}
	fmt.Println("WARNING:", totalLineCount)

	// Test for CRITICAL
	input = "grep CRITICAL"
	err, totalLineCount = fileServer.GrepMultipleServers(&input, &fileMap, &reply)
	if err != nil {
		fmt.Println("Error:", err)
	}
	if totalLineCount != fileCountMap["CRITICAL"] {
		t.Errorf("Expected %d, got %d", fileCountMap["CRITICAL"], totalLineCount)
	}
	fmt.Println("CRITICAL:", totalLineCount)

}