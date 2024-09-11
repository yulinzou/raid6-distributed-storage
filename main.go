package main

import (
	"fmt"
	"math/rand"
	"raid6-distributed-storage/raid6"
	"time"
)

func main() {
	raid := raid6.InitRAID6()

	// Example file data (representing as bytes)
	fileData := []byte("this_is_a_test_file_data")

	// Write file to RAID 6
	err := raid.WriteFile("test.txt", fileData)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Example file data - 10 different sentences
	fileDataList := []string{
		"File 1: This is the first test file.",
		"File 2: Another test file for RAID-6.",
		"File 3: RAID-6 testing with multiple files.",
		"File 4: Data recovery is crucial in RAID-6.",
		"File 5: This sentence will be stored on RAID-6.",
		"File 6: RAID-6 provides fault tolerance.",
		"File 7: Each file is spread across multiple disks.",
		"File 8: Let's simulate node failures.",
		"File 9: Testing resilience of RAID-6 storage.",
		"File 10: Final test for RAID-6 file system.",
	}

	// Write all the files to RAID-6
	for i, fileData := range fileDataList {
		fileName := fmt.Sprintf("file%d.txt", i+1)
		fmt.Printf("Writing file: %s\n", fileName)
		err := raid.WriteFile(fileName, []byte(fileData))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	return 
	
	// Random seed for picking files
	rand.Seed(time.Now().UnixNano())

	// Test a random number of files (e.g., 5 out of 10) for two-node failures
	numFilesToTest := 5
	selectedFiles := rand.Perm(len(fileDataList))[:numFilesToTest]

	// Randomly pick two nodes for failure
	totalNodes := raid.DiskNum
	nodeID1 := rand.Intn(totalNodes)
	nodeID2 := rand.Intn(totalNodes)
	for nodeID2 == nodeID1 { // Ensure nodeID2 is different from nodeID1
		nodeID2 = rand.Intn(totalNodes)
	}

	fmt.Printf("\nSimulating failure of node %d and node %d...\n", nodeID1, nodeID2)
	raid.TwoNodesFailure(nodeID1, nodeID2)

	// Attempt to recover the failed nodes for randomly selected files
	for _, fileIndex := range selectedFiles {
		fmt.Printf("Recovering node %d and node %d for file%d.txt...\n", nodeID1, nodeID2, fileIndex+1)
		raid.RecoverDoubleNodes(nodeID1, nodeID2)

		// Verify recovery
		recoveredData, err := raid.ReadFile(fileIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Recovered data for file%d.txt: %s\n", fileIndex+1, string(recoveredData))
	}
	
	// Iterate through all nodes and simulate failure + recovery
	// for nodeID := 0; nodeID < raid.DiskNum; nodeID++ {
	// 	// Simulate a single node failure
	// 	fmt.Printf("Simulating a single node failure on node %d...\n", nodeID)
	// 	raid.NodeFailure(nodeID)

	// 	// Recover the failed node
	// 	fmt.Printf("Recovering node %d...\n", nodeID)
	// 	raid.RecoverSingleNode(nodeID)

	// 	// Read the file and check if the data is recovered correctly
	// 	fileData, err = raid.ReadFile(0)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}

	// 	// Output the recovered file data
	// 	fmt.Printf("Recovered data after failure of node %d: %s\n", nodeID, string(fileData))
	// }
	
	// totalNodes := raid.DiskNum

	// // Iterate over all possible combinations of two nodes
	// for nodeID1 := 0; nodeID1 < totalNodes; nodeID1++ {
	// 	for nodeID2 := 0; nodeID2 < totalNodes; nodeID2++ {
	// 		if nodeID1 == nodeID2 {
	// 			continue
	// 		}
	// 		// Simulate two node failures
	// 		fmt.Printf("Simulating failure of node %d and node %d...\n", nodeID1, nodeID2)
	// 		raid.TwoNodesFailure(nodeID1, nodeID2)

	// 		// Attempt to recover the failed nodes
	// 		fmt.Printf("Recovering node %d and node %d...\n", nodeID1, nodeID2)
	// 		raid.RecoverDoubleNodes(nodeID1, nodeID2)

	// 		fileData, err = raid.ReadFile(0)
	// 		if err != nil {	
	// 			fmt.Println(err)
	// 			return
	// 		}
	// 		fmt.Println(string(fileData))
	// 	}
	// }
	

	// // Check if the recovery was successful
	// recoveredData := raid.Nodes[0].BlockList[0].Data
	// if recoveredData != nil {
	// 	fmt.Println("Data recovered successfully:", string(*recoveredData))
	// } else {
	// 	fmt.Println("Recovery failed!")
	// }


	// Check for corruption
	// if raid.CheckCorruption() {
	// 	fmt.Println("Data is corrupted!")
	// } else {
	// 	fmt.Println("Data is safe.")
	// }

	// // Need to edit
	// if err := raid.RecoverData(); err != nil {
	// 	fmt.Println(err)
	// }
}
