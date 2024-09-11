package main

import (
	"fmt"
	// "math/rand"
	"raid6-distributed-storage/raid6"
	"raid6-distributed-storage/test"
	"time"
)

func main() {
	raid := raid6.InitRAID6()

	// Simulate number of files
	numFiles := 100
	numsingleFailures := 100
	numdoubleFailures := 100
	maxSize := 50 // Max file size (50 characters)

	// Generate random file data
	fileNames, fileContents, err := test.GenerateRandomFileData(numFiles, maxSize)
	if err != nil {
		fmt.Println("Error generating files:", err)
		return
	}

	// Store generated files
	err = test.StoreFileData("files.txt", fileNames, fileContents)
	if err != nil {
		fmt.Println("Error storing files:", err)
		return
	}

	// Generate and store single failure cases
	singleFailures, err := test.GenerateSingleFailureCases(numsingleFailures, raid.DiskNum)
	if err != nil {
		fmt.Println("Error generating single failures:", err)
		return
	}
	err = test.StoreSingleFailureData("single_failures.txt", singleFailures)
	if err != nil {
		fmt.Println("Error storing single failures:", err)
		return
	}

	// Generate and store double failure cases
	doubleFailures, err := test.GenerateDoubleFailureCases(numdoubleFailures, raid.DiskNum)
	if err != nil {
		fmt.Println("Error generating double failures:", err)
		return
	}
	err = test.StoreDoubleFailureData("double_failures.txt", doubleFailures)
	if err != nil {
		fmt.Println("Error storing double failures:", err)
		return
	}

	// Testing single node failure recovery
	for i := 0; i < numFiles; i++ {
		// Write file to RAID 6
		err := raid.WriteFile(fileNames[i], []byte(fileContents[i]))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Simulating a single node failure on node %d...\n", singleFailures[i])
		raid.NodeFailure(singleFailures[i])

		// Recover the failed node and measure recovery time
		startTime := time.Now()
		fmt.Printf("Recovering node %d...\n", singleFailures[i])
		raid.RecoverSingleNode(singleFailures[i])
		elapsedTime := time.Since(startTime)

		fileData, err := raid.ReadFile(i)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Recovered file content for file %d: %s\n", i, string(fileData))
		fmt.Printf("Recovery time for node %d: %v\n\n", singleFailures[i], elapsedTime)
	}

	// Testing double node failure recovery
	for i := 0; i < numFiles; i++ {
		failure := doubleFailures[i]

		fmt.Printf("Simulating failure of nodes %d and %d...\n", failure[0], failure[1])
		raid.TwoNodesFailure(failure[0], failure[1])

		// Recover the failed nodes and measure recovery time
		startTime := time.Now()
		fmt.Printf("Recovering nodes %d and %d...\n", failure[0], failure[1])
		raid.RecoverDoubleNodes(failure[0], failure[1])
		elapsedTime := time.Since(startTime)

		fileData, err := raid.ReadFile(i)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Recovered file content for file %d: %s\n", i, string(fileData))
		fmt.Printf("Recovery time for nodes %d and %d: %v\n\n", failure[0], failure[1], elapsedTime)
	}
}