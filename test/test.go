package test

import (
	"fmt"
	"os"
	"raid6-distributed-storage/raid6"
	"strconv"
	"strings"
	"time"
)

// Test function to read test data from files and simulate failures
func RunRecoveryTests(raid *raid6.RAID6) {
	// Read file names and content from files.txt
	fileData, err := os.ReadFile(FilePath)
	if err != nil {
		fmt.Println("Error reading file data:", err)
		return
	}
	lines := strings.Split(string(fileData), "\n")

	// Write all files to RAID 6 and calculate write time
	writeStart := time.Now()

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		fileName := parts[0]
		fileContent := parts[1]

		err := raid.WriteFile(fileName, []byte(fileContent))
		if err != nil {
			return
		}
	}

	writeTime := time.Since(writeStart)
	fmt.Printf("=====================================\n")
	fmt.Printf("Total write time: %s\n", writeTime)
	fmt.Printf("Total number of files written: %d, Average write time per file: %s\n", len(lines)-1, writeTime/time.Duration(len(lines)-1))
	// Simulate single node failure cases
	runSingleFailureTests(raid)

	// Simulate double node failure cases
	runDoubleFailureTests(raid)
}

// Run single node failure recovery tests
func runSingleFailureTests(raid *raid6.RAID6) {
	singleFailures, err := os.ReadFile(SFilePath)
	if err != nil {
		return
	}

	singleFailureLines := strings.Split(string(singleFailures), "\n")
	startTime := time.Now()
	totalTests := 0

	for _, line := range singleFailureLines {
		if line == "" {
			continue
		}

		nodeID, err := strconv.Atoi(line)
		if err != nil {
			continue
		}

		// Simulate and recover single node failure
		raid.NodeFailure(nodeID)
		raid.RecoverSingleNode(nodeID)
		totalTests++
	}

	totalTime := time.Since(startTime)
	fmt.Printf("=====================================\n")
	fmt.Printf("Single node recovery tests completed in: %s\n", totalTime)
	fmt.Printf("Total number of single node failures tested: %d, Average recovery time per test: %s\n", totalTests, totalTime/time.Duration(totalTests))
}

// Run double node failure recovery tests
func runDoubleFailureTests(raid *raid6.RAID6) {
	doubleFailures, err := os.ReadFile(DFilePath)
	if err != nil {
		return
	}

	doubleFailureLines := strings.Split(string(doubleFailures), "\n")
	startTime := time.Now()
	totalTests := 0

	for _, line := range doubleFailureLines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		nodeID1, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		nodeID2, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		// Simulate and recover double node failure
		raid.TwoNodesFailure(nodeID1, nodeID2)
		raid.RecoverDoubleNodes(nodeID1, nodeID2)
		totalTests++
	}

	totalTime := time.Since(startTime)
	fmt.Printf("=====================================\n")
	fmt.Printf("Double node recovery tests completed in: %s\n", totalTime)
	fmt.Printf("Total number of double node failures tested: %d, Average recovery time per test: %s\n", totalTests, totalTime/time.Duration(totalTests))
	
}