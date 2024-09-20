package raid6

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type Block struct {
	FileName string
	Data     *[]byte
	BlockID  int
	Size     int
}

type Node struct {
	NodeID   int
	status   bool // true for active, false for inactive(failure)
	DiskPath string
}

func InitBlock(blockID int, fileName string, data *[]byte, blockSize int) *Block {
	return &Block{
		FileName: fileName,
		Data:     data,
		BlockID:  blockID, // -1 for P parity, -2 for Q parity
		Size:     blockSize,
	}
}

func (n *Node) getBlockFilePath(fileName string, blockID int) string {
	return fmt.Sprintf("%s/%s_%d.bin", n.DiskPath, fileName, blockID)
}

func (n *Node) CheckBlockExists(fileName string, blockID int) bool {
	filePath := n.getBlockFilePath(fileName, blockID)
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	// Some other error occurred (e.g., permission denied)
	return false
}

func (n *Node) CheckFileExists(fileName string) (bool, error) {
	pattern := filepath.Join(n.DiskPath, fileName+"*")

	// Find files matching the pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, err
	}

	if len(matches) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (n *Node) ScanFileNames() ([]string, error) {
	pattern := regexp.MustCompile(`^(.+?)_([^_]+)\.bin$`)
	fileNames := []string{}

	// Find files matching the pattern
	entries, err := os.ReadDir(n.DiskPath)
	if err != nil {
		return fileNames, fmt.Errorf("failed to read directory: %w", err)
	}

	// Iterate over each entry
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		// Get the filename
		name := entry.Name()

		// Match the filename against the pattern
		matches := pattern.FindStringSubmatch(name)
		if len(matches) == 3 {
			// Extract fileName (group 1)
			fileName := matches[1]
			fileNames = append(fileNames, fileName)
		}
	}

	return fileNames, nil
}

// ReadBlockFromDisk reads a block's data based on block ID
func (n *Node) ReadBlockFromDisk(fileName string, blockID int) ([]byte, error) {
	filePath := n.getBlockFilePath(fileName, blockID)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	blockData := make([]byte, fileInfo.Size())
	_, err = file.Read(blockData)
	if err != nil {
		return nil, err
	}

	return blockData, nil
}

// WriteBlockToDisk writes data to a block based on block ID and file name
func (n *Node) WriteBlockToDisk(b *Block) error {
	filePath := n.getBlockFilePath(b.FileName, b.BlockID)

	// Remove the old file if it exists
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err // Return error if removal fails for reasons other than the file not existing
	}

	// Open file with read/write permission. Create it if it does not exist.
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(*b.Data)
	if err != nil {
		return err
	}

	return nil
}

func InitNode(nodeID int, diskPath string) *Node {
	// Ensure the diskPath exists
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		os.MkdirAll(diskPath, os.ModePerm)
	}

	return &Node{
		NodeID:   nodeID,
		status:   true,
		DiskPath: diskPath,
	}
}

func (n *Node) Corrupt() error {
	n.status = false
	entries, err := os.ReadDir(n.DiskPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		err = os.Remove(filepath.Join(n.DiskPath, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
