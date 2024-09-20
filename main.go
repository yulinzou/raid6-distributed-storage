package main

import (
	"fmt"
	"os"
	"raid6-distributed-storage/raid6"
	"raid6-distributed-storage/test"
)

var (
	FileNum     = 20
	SFailureNum = 5
	DFailureNum = 5
	UpdateNum   = 5
	MaxFileSize = 200
	BasePath    = "./raid6_cluster"
)

func main() {
	err := os.RemoveAll(BasePath)
	if err != nil {
		return
	}

	raid := raid6.InitRAID6(8, BasePath)

	// Generate random file names and contents
	err = test.GenerateRandomTestData(FileNum, SFailureNum, DFailureNum, MaxFileSize, raid.DiskNum)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Run recovery tests
	test.RunRecoveryTests(raid)

	test.RunUpdateTests(raid, UpdateNum, MaxFileSize)
}
