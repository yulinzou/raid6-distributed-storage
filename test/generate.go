package test

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

var (
	FilePath = "test/files.txt"
	SFilePath = "test/single_failures.txt"
	DFilePath = "test/double_failures.txt"
)

// Generate files for testing
func GenerateRandomTestData(FileNum, SFailureNum, DFailureNum, MaxFileSize, DiskNum int) (error){
	
	err := GenerateRandomFileData(FileNum, MaxFileSize) 
	if err != nil {
		return err
	}
	err = GenerateSingleFailureCases(SFailureNum, DiskNum)
	if err != nil {
		return err
	}
	err = GenerateDoubleFailureCases(DFailureNum, DiskNum)
	if err != nil {
		return err
	}
	
	return nil
	
}


// GenerateRandomFileData generates random file names and contents (with readable characters).
func GenerateRandomFileData(numFiles int, maxSize int) (error) {
	rand.Seed(time.Now().UnixNano())
	fileNames := make([]string, numFiles)
	fileContents := make([]string, numFiles)

	for i := 0; i < numFiles; i++ {
		// Generate random file name
		fileName := fmt.Sprintf("file_%d.txt", i)
		fileNames[i] = fileName

		// Generate random file content using readable ASCII characters (letters and digits)
		fileSize := rand.Intn(maxSize) + 1
		fileContent := make([]byte, fileSize)
		for j := 0; j < fileSize; j++ {
			fileContent[j] = randomASCIIChar()
		}
		fileContents[i] = string(fileContent)
	}

	err := StoreFileData(fileNames, fileContents)
	if err != nil {
		return err
	} 

	return nil
}

// GenerateSingleFailureCases generates single node failure cases for testing.
func GenerateSingleFailureCases(numFiles int, diskNum int) (error) {
	rand.Seed(time.Now().UnixNano())
	failures := make([]int, numFiles)

	for i := 0; i < numFiles; i++ {
		failures[i] = rand.Intn(diskNum)
	}

	err := StoreSingleFailureData(failures)
	if err != nil {
		return err
	}

	return nil
}

// GenerateDoubleFailureCases generates double node failure cases for testing.
func GenerateDoubleFailureCases(numFiles int, diskNum int) (error) {
	rand.Seed(time.Now().UnixNano())
	failures := make([][2]int, numFiles)

	for i := 0; i < numFiles; i++ {
		nodeID1 := rand.Intn(diskNum)
		nodeID2 := rand.Intn(diskNum)
		for nodeID1 == nodeID2 {
			nodeID2 = rand.Intn(diskNum)
		}
		failures[i] = [2]int{nodeID1, nodeID2}
	}

	err := StoreDoubleFailureData(failures)
	if err != nil {
		return err
	}

	return nil
}

// randomASCIIChar generates a random ASCII character from 'a' to 'z', 'A' to 'Z', or '0' to '9'.
func randomASCIIChar() byte {
	ranges := []struct {
		low, high byte
	}{
		{low: 'a', high: 'z'}, // lowercase letters
		{low: 'A', high: 'Z'}, // uppercase letters
		{low: '0', high: '9'}, // digits
	}

	randRange := ranges[rand.Intn(len(ranges))]
	return randRange.low + byte(rand.Intn(int(randRange.high-randRange.low+1)))
}

// StoreFileData writes filenames and file contents to "files.txt".
func StoreFileData(fileNames []string, fileContents []string) error {
	file, err := os.Create(FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i, fileName := range fileNames {
		fileContent := fileContents[i]
		_, err := file.WriteString(fmt.Sprintf("%s %s\n", fileName, fileContent))
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreSingleFailureData writes single node failure cases to "single_failures.txt".
func StoreSingleFailureData(failures []int) error {
	file, err := os.Create(SFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, failure := range failures {
		_, err := file.WriteString(fmt.Sprintf("%d\n", failure))
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreDoubleFailureData writes double node failure cases to "double_failures.txt".
func StoreDoubleFailureData(failures [][2]int) error {
	file, err := os.Create(DFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, failure := range failures {
		_, err := file.WriteString(fmt.Sprintf("%d %d\n", failure[0], failure[1]))
		if err != nil {
			return err
		}
	}

	return nil
}