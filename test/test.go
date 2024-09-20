package test

import (
	"fmt"
	"math/rand"
	"os"
	"raid6-distributed-storage/raid6"
	"strconv"
	"strings"
	"time"
)

// Function to verify the integrity of all files after recovery
func VerifyAllFilesIntegrity(raid *raid6.RAID6) {
	// Read the original file data from files.txt
	fileData, err := os.ReadFile(FilePath)
	if err != nil {
		fmt.Println("Error reading file data:", err)
		return
	}
	lines := strings.Split(string(fileData), "\n")

	// Variable to track total files and mismatches
	mismatchCount := 0

	// Iterate over all files in files.txt
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		fileName := parts[0]
		originalContent := parts[1]

		// Read the file from the RAID system
		recoveredContent, err := raid.ReadFile(fileName)
		if err != nil {
			fmt.Printf("Error reading file %s from RAID: %s\n", fileName, err)
			mismatchCount++
			continue
		}

		// Compare the original content with the recovered content
		if originalContent != string(recoveredContent) {
			fmt.Printf("File %s recovery failed, contents do not match\n", fileName)
			fmt.Println("Original content: ", originalContent, len(originalContent))
			fmt.Println("Retrieved content:", string(recoveredContent), len(recoveredContent))
			mismatchCount++
		}
	}

	// Final report
	fmt.Printf("=====================================\n")
	if mismatchCount == 0 {
		fmt.Printf("All %d files successfully checked and are valid.\n", raid.FileNum)
	} else {
		fmt.Printf("%d files out of %d have mismatches or errors in recovery.\n", mismatchCount, raid.FileNum)
	}
}

func RunUpdateTests(raid *raid6.RAID6, updateNums, maxSize int) {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	fmt.Printf("+++++++++++++++++++++\nUpdate Test begin\n")
	// Read file names and content from files.txt
	fileData, err := os.ReadFile(FilePath)
	if err != nil {
		fmt.Println("Error reading file data:", err)
		return
	}
	lines := strings.Split(string(fileData), "\n")

	perm := rand.Perm(len(lines) - 1)
	fmt.Println(len(lines) - 1)

	updateMap := make(map[string][]byte)
	for i := 0; i < updateNums; i++ {
		fileName := strings.SplitN(lines[perm[i]], " ", 2)[0]
		fileSize := rand.Intn(maxSize) + 1
		fileContent := make([]byte, fileSize)
		for j := 0; j < fileSize; j++ {
			fileContent[j] = randomASCIIChar()
		}
		err = updateSingleFile(fileName, string(fileContent))
		updateMap[fileName] = fileContent
	}

	updateStart := time.Now()
	for name, content := range updateMap {
		err = raid.UpdateFile(name, content)
		if err != nil {
			fmt.Println("Error updating file:", err)
		}
	}
	updateTime := time.Since(updateStart)

	fmt.Printf("Total update time: %s\n", updateTime)
	fmt.Printf("Total number of files updated: %d, Average update time per file: %s\n", updateNums, updateTime/time.Duration(updateNums))
	VerifyAllFilesIntegrity(raid)

}

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

		err = raid.WriteFile(fileName, []byte(fileContent))
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

	// Verify the integrity of all files after recovery
	VerifyAllFilesIntegrity(raid)
}

// Run single node failure recovery tests
func runSingleFailureTests(raid *raid6.RAID6) {
	singleFailures, err := os.ReadFile(SFilePath)
	if err != nil {
		fmt.Println("Error reading file data:", err)
		return
	}

	singleFailureLines := strings.Split(string(singleFailures), "\n")
	startTime := time.Now()
	totalTests := 0

	for _, line := range singleFailureLines {
		if line == "" {
			continue
		}

		var nodeID int
		nodeID, err = strconv.Atoi(line)
		if err != nil {
			fmt.Printf(err.Error())
		}

		// Simulate and recover single node failure
		err = raid.NodeFailure(nodeID)
		if err != nil {
			fmt.Printf(err.Error())
		}
		err = raid.RecoverSingleNode(nodeID)
		if err != nil {
			fmt.Printf(err.Error())
		}
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

		var nodeID1, nodeID2 int
		nodeID1, err = strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		nodeID2, err = strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		// Simulate and recover double node failure
		raid.TwoNodesFailure(nodeID1, nodeID2)
		if nodeID1 < nodeID2 {
			err = raid.RecoverDoubleNodes(nodeID1, nodeID2)
			if err != nil {
				continue
			}
		} else {
			err = raid.RecoverDoubleNodes(nodeID2, nodeID1)
			if err != nil {
				continue
			}
		}
		totalTests++
	}

	totalTime := time.Since(startTime)
	fmt.Printf("=====================================\n")
	fmt.Printf("Double node recovery tests completed in: %s\n", totalTime)
	fmt.Printf("Total number of double node failures tested: %d, Average recovery time per test: %s\n", totalTests, totalTime/time.Duration(totalTests))

}
