package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"testing"
	"time"
)

var logPatterns = map[string]string{
	"INFO":     "Func executed successfully", // Frequent Pattern
	"ERROR":    "Func failed",                // Rare Pattern
	"DEBUG":    "Debugging info",
	"CRITICAL": "Critical error", // Extrmely rare log
	"WARNING":  "Warning",        // Somewhat Frequent
}

// Map of server to filename
var fileMap = map[string]string{
	"fa24-cs425-3101.cs.illinois.edu:2232": "vm1_test.log",
	"fa24-cs425-3102.cs.illinois.edu:2232": "vm2_test.log",
	"fa24-cs425-3103.cs.illinois.edu:2232": "vm3_test.log",
	"fa24-cs425-3104.cs.illinois.edu:2232": "vm4_test.log",
	"fa24-cs425-3105.cs.illinois.edu:2232": "vm5_test.log",
	"fa24-cs425-3106.cs.illinois.edu:2232": "vm6_test.log",
	"fa24-cs425-3107.cs.illinois.edu:2232": "vm7_test.log",
	"fa24-cs425-3108.cs.illinois.edu:2232": "vm8_test.log",
	"fa24-cs425-3109.cs.illinois.edu:2232": "vm9_test.log",
	"fa24-cs425-3110.cs.illinois.edu:2232": "vm10_test.log",
}

// this method creates the different log patterns
func createPattern(logType string, tmp int) string {
	content := ""
	for i := 0; i < tmp; i++ {
		content += fmt.Sprintf("%s: %s\n", logType, logPatterns[logType])
	}
	return content
}

// this method generates the test log for each query pattern and returns count of pattern in map
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

// create test log files with different patterns on each server
func createTestLogFiles() map[string]int {
	var reply string

	// Map to store the count of each log pattern for verification
	fileCountMap := map[string]int{
		"INFO":     0,
		"ERROR":    0,
		"DEBUG":    0,
		"CRITICAL": 0,
		"WARNING":  0,
	}
	for server, filename := range fileMap {
		content, countMap := createTestLog()
		if server == "fa24-cs425-3110.cs.illinois.edu" { // Handle extremely rare log case
			content += createPattern("CRITICAL", 1)
			countMap["CRITICAL"] = 1
		}

		conn, err := net.DialTimeout("tcp", server, 2*time.Second)
		if err != nil {
			fmt.Printf("Failed to connect to server %s with error: %v\n", server, err)
			continue
		}
		client := rpc.NewClient(conn)
		defer client.Close()

		err = client.Call("FileServer.WriteFile", &FileRequest{Filename: filename, Content: content}, &reply)
		fmt.Printf("Created test log file for %s\n", server)

		for k, v := range countMap {
			fileCountMap[k] += v
		}
	}

	return fileCountMap
}

// Function to test the grep method on multiple servers concurrently wiht different query patterns
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
