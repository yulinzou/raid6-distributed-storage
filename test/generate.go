package test

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// GenerateRandomFileData generates random file names and contents (with readable characters).
func GenerateRandomFileData(numFiles int, maxSize int) ([]string, []string, error) {
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

	return fileNames, fileContents, nil
}

// GenerateSingleFailureCases generates single node failure cases for testing.
func GenerateSingleFailureCases(numFiles int, diskNum int) ([]int, error) {
	rand.Seed(time.Now().UnixNano())
	failures := make([]int, numFiles)

	for i := 0; i < numFiles; i++ {
		failures[i] = rand.Intn(diskNum)
	}

	return failures, nil
}

// GenerateDoubleFailureCases generates double node failure cases for testing.
func GenerateDoubleFailureCases(numFiles int, diskNum int) ([][2]int, error) {
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

	return failures, nil
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
func StoreFileData(outputFile string, fileNames []string, fileContents []string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for i, fileName := range fileNames {
		fileContent := fileContents[i]
		_, err := file.WriteString(fmt.Sprintf("File: %s\nContent: %s\n\n", fileName, fileContent))
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreSingleFailureData writes single node failure cases to "single_failures.txt".
func StoreSingleFailureData(outputFile string, failures []int) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, failure := range failures {
		_, err := file.WriteString(fmt.Sprintf("Failure: Node %d\n", failure))
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreDoubleFailureData writes double node failure cases to "double_failures.txt".
func StoreDoubleFailureData(outputFile string, failures [][2]int) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, failure := range failures {
		_, err := file.WriteString(fmt.Sprintf("Failure: Node %d, Node %d\n", failure[0], failure[1]))
		if err != nil {
			return err
		}
	}

	return nil
}