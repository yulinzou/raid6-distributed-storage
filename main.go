package main

import (
	"fmt"
	// "math/rand"
	"raid6-distributed-storage/raid6"
	"raid6-distributed-storage/test"
)

var (
	FileNum     = 100
	SFailureNum = 100
	DFailureNum = 100
	MaxFileSize = 50
)

func main() {
	raid := raid6.InitRAID6()

	// Generate random file names and contents
	err := test.GenerateRandomTestData(FileNum, SFailureNum, DFailureNum, MaxFileSize, raid.DiskNum)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Run recovery tests
	test.RunRecoveryTests(raid)

	test.RunUpdateTests(raid)

}
