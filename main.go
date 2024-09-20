package main

import (
	"fmt"
	"raid6-distributed-storage/raid6"
	"raid6-distributed-storage/test"
)

var (
	FileNum     = 20
	SFailureNum = 5
	DFailureNum = 5
	UpdateNum   = 5
	MaxFileSize = 100
)

func main() {
	raid := raid6.InitRAID6(10, "./raid6_cluster")

	// Generate random file names and contents
	err := test.GenerateRandomTestData(FileNum, SFailureNum, DFailureNum, MaxFileSize, raid.DiskNum)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Run recovery tests
	test.RunRecoveryTests(raid)

	test.RunUpdateTests(raid, UpdateNum, MaxFileSize)
}
