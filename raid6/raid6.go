package raid6

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type RAID6 struct {
	Nodes     []*Node
	Math      *RAIDMath
	FileNum   int
	FileNames []string
	DiskNum   int
	sync.Mutex
}

func InitRAID6(numDisks int, basePath string) *RAID6 {
	raid := &RAID6{
		DiskNum:   numDisks,
		Nodes:     make([]*Node, numDisks), // 6 data nodes, 2 parity nodes
		Math:      NewRAIDMath(2),          // Generator 2 for GF(2^8)
		FileNames: make([]string, 0),
		FileNum:   0, // no file at the beginning
	}

	for i := 0; i < raid.DiskNum; i++ {
		diskPath := fmt.Sprintf("%s/disk_%d", basePath, i)
		raid.Nodes[i] = InitNode(i, diskPath)
	}

	return raid
}

func (r *RAID6) ScanFileNames() (err error) {
	r.FileNames, err = r.Nodes[0].ScanFileNames()
	if err != nil {
		return err
	}
	r.FileNum = len(r.FileNames)
	return nil
}

// WriteFile Splits input data into blocks, calculates parity blocks and writes them to nodes.
func (r *RAID6) WriteFile(fileName string, data []byte) error {
	r.Lock()
	defer r.Unlock()

	if len(data) == 0 {
		return errors.New("file data is empty")
	}

	// Randomly select two indices for P and Q parity
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) // Seed the random number generator
	parityIndices := rnd.Perm(r.DiskNum)[:2]               // Randomly pick two unique indices from the range of FileNum

	// Assign P and Q parity dataBlocks to the randomly selected indices
	pIndex := parityIndices[0]
	qIndex := parityIndices[1]

	// Number of data disks (excluding the parity disks)
	numDataBlocks := r.DiskNum - 2         // 2 disks for P and Q parity
	blockSize := len(data) / numDataBlocks // Block size with rounding up for padding
	if len(data)%numDataBlocks != 0 {
		blockSize++
	}

	dataBlocks := make([][]byte, numDataBlocks)
	for i := 0; i < r.DiskNum-2; i++ {
		dataBlocks[i] = make([]byte, blockSize)
		start := i * blockSize
		end := start + blockSize
		if end > len(data) {
			end = len(data)
		}
		if start < end {
			copy(dataBlocks[i], data[start:end])
		}
	}

	// Write parity blocks into nodes
	pParity, qParity := r.Math.CalculateParity(dataBlocks, blockSize)
	pBlock := InitBlock(-1, fileName, &pParity, blockSize)
	err := r.Nodes[pIndex].WriteBlockToDisk(pBlock)
	if err != nil {
		return err
	}
	qBlock := InitBlock(-2, fileName, &qParity, blockSize)
	err = r.Nodes[qIndex].WriteBlockToDisk(qBlock)
	if err != nil {
		return err
	}

	// Write data blocks into nodes
	nodeID := 0
	for i, dataBlock := range dataBlocks {
		for nodeID == pIndex || nodeID == qIndex {
			nodeID++
		}
		block := InitBlock(i, fileName, &dataBlock, blockSize)
		err = r.Nodes[nodeID].WriteBlockToDisk(block)
		if err != nil {
			return err
		}
		nodeID++
	}

	r.FileNum++
	r.FileNames = append(r.FileNames, fileName)
	return nil
}

// ReadFile Read the file data from the RAID 6 by file name
func (r *RAID6) ReadFile(fileName string) ([]byte, error) {
	r.Lock()
	defer r.Unlock()

	exist := false
	for _, name := range r.FileNames {
		if name == fileName {
			exist = true
			break
		}
	}
	if !exist {
		return nil, errors.New("file does not exist")
	}
	// Get the data blocks, P parity, and Q parity for the current file (index i)
	dataBlocks, _, _ := r.GetDataBlocks(fileName)

	if !r.CheckStatus() {
		return nil, errors.New("node failure")
	}

	// Concatenate the data blocks to recover the original file data
	fileData := make([]byte, 0)

	for i := 0; i < len(dataBlocks); i++ {
		if dataBlocks[i] != nil {
			fileData = append(fileData, dataBlocks[i]...)
		}
	}
	fileData = bytes.TrimRight(fileData, "\x00") // Remove padding

	return fileData, nil
}

// CheckStatus Check if all nodes are active
func (r *RAID6) CheckStatus() bool {
	for i := 0; i < r.DiskNum; i++ {
		if !r.Nodes[i].status {
			return false
		}
	}
	return true
}

// UpdateFile Update the file content given file name and updated data
func (r *RAID6) UpdateFile(fileName string, data []byte) error {
	r.Lock()
	defer r.Unlock()

	if len(data) == 0 {
		return errors.New("file data is empty")
	}

	exist := false
	for _, name := range r.FileNames {
		if name == fileName {
			exist = true
			break
		}
	}
	if !exist {
		return errors.New("file does not exist")
	}

	numDataBlocks := r.DiskNum - 2
	blockSize := len(data) / numDataBlocks
	if len(data)%numDataBlocks != 0 {
		blockSize++
	}

	dataBlocks := make([][]byte, r.DiskNum-2)
	for i := 0; i < r.DiskNum-2; i++ {
		dataBlocks[i] = make([]byte, blockSize)
		start := i * blockSize
		end := start + blockSize
		if end > len(data) {
			end = len(data)
		}
		if start < end {
			copy(dataBlocks[i], data[start:end])
		}
	}

	// Write parity blocks into nodes
	pParity, qParity := r.Math.CalculateParity(dataBlocks, blockSize)
	pBlock := InitBlock(-1, fileName, &pParity, blockSize)
	qBlock := InitBlock(-2, fileName, &qParity, blockSize)
	for _, node := range r.Nodes {
		if node.CheckBlockExists(fileName, -2) {
			err := node.WriteBlockToDisk(qBlock)
			if err != nil {
				return err
			}
		} else if node.CheckBlockExists(fileName, -1) {
			err := node.WriteBlockToDisk(pBlock)
			if err != nil {
				return err
			}
		}
		for j := 0; j < r.DiskNum-2; j++ {
			if node.CheckBlockExists(fileName, j) {
				err := node.WriteBlockToDisk(InitBlock(j, fileName, &dataBlocks[j], blockSize))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetDataBlocks Get data blocks from nodes
func (r *RAID6) GetDataBlocks(fileName string) (dataBlocks [][]byte, P []byte, Q []byte) {
	dataBlocks = make([][]byte, r.DiskNum-2) // Initialize dataBlocks for n-2 data disks
	P = []byte{}                             // Initialize P as empty byte slice
	Q = []byte{}                             // Initialize Q as empty byte slice

	var pFound, qFound bool
	for nodeID := 0; nodeID < r.DiskNum; nodeID++ {
		node := r.Nodes[nodeID]

		// Check for Parity P (-1)
		if !pFound && node.CheckBlockExists(fileName, -1) {
			P, _ = node.ReadBlockFromDisk(fileName, -1)
			pFound = true
			continue
		}

		// Check for Parity Q (-2)
		if !qFound && node.CheckBlockExists(fileName, -2) {
			Q, _ = node.ReadBlockFromDisk(fileName, -2)
			qFound = true
			continue
		}

		for i := 0; i < r.DiskNum-2; i++ {
			if node.CheckBlockExists(fileName, i) {
				data, _ := node.ReadBlockFromDisk(fileName, i)
				dataBlocks[i] = data
				break
			}
		}
	}

	return dataBlocks, P, Q
}

// NodeFailure Simulate single node's failure
func (r *RAID6) NodeFailure(nodeID int) error {
	err := r.Nodes[nodeID].Corrupt()
	if err != nil {
		return err
	}

	return nil
}

// TwoNodesFailure Simulate two nodes' failure
func (r *RAID6) TwoNodesFailure(nodeID1, nodeID2 int) {
	r.Nodes[nodeID1].Corrupt()
	r.Nodes[nodeID2].Corrupt()
}

// RecoverFile Recover single file with single node failure
func (r *RAID6) RecoverFile(nodeID int, fileName string) error {
	if fileName == "" {
		return errors.New("file name is empty")
	}

	dataBlocks, P, Q := r.GetDataBlocks(fileName)
	var blockIndex int
	if len(P) == 0 {
		blockIndex = -1
	} else if len(Q) == 0 {
		blockIndex = -2
	} else {
		for i, dataBlock := range dataBlocks {
			if len(dataBlock) == 0 {
				blockIndex = i
			}
		}
	}

	if blockIndex >= 0 {
		dataBlock := r.Math.RecoverSingleBlockP(dataBlocks, P, blockIndex)
		err := r.Nodes[nodeID].WriteBlockToDisk(InitBlock(blockIndex, fileName, &dataBlock, len(dataBlock)))
		if err != nil {
			return err
		}
	} else if blockIndex == -1 {
		pBlock := r.Math.RecoverPParity(dataBlocks)
		err := r.Nodes[nodeID].WriteBlockToDisk(InitBlock(-1, fileName, &pBlock, len(pBlock)))
		if err != nil {
			return err
		}
	} else if blockIndex == -2 {
		qBlock := r.Math.RecoverQParity(dataBlocks)
		err := r.Nodes[nodeID].WriteBlockToDisk(InitBlock(-2, fileName, &qBlock, len(qBlock)))
		if err != nil {
			return err
		}
	}

	return nil
}

// RecoverSingleNode Single node recovery function
func (r *RAID6) RecoverSingleNode(nodeID int) error {
	// For the ith file
	for _, fileName := range r.FileNames {
		err := r.RecoverFile(nodeID, fileName)
		if err != nil {
			return err
		}
	}

	r.Nodes[nodeID].status = true
	return nil
}

// RecoverDoubleNodes Double nodes recovery function. Assume that nodeID1 < nodeID2
func (r *RAID6) RecoverDoubleNodes(nodeID1, nodeID2 int) error {
	for _, fileName := range r.FileNames {
		// Get the data blocks, P parity, and Q parity for the current file (index i)
		dataBlocks, P, Q := r.GetDataBlocks(fileName)

		// Block types of the failed blocks
		var blockIndices []int
		var err error
		if len(Q) == 0 {
			blockIndices = append(blockIndices, -2)
		}
		if len(P) == 0 {
			blockIndices = append(blockIndices, -1)
		}
		for i, dataBlock := range dataBlocks {
			if len(dataBlock) == 0 {
				blockIndices = append(blockIndices, i)
			}
		}

		blockIndex1 := blockIndices[1]
		blockIndex2 := blockIndices[0]

		// Recovery logic based on the block types of the two failed nodes
		if blockIndex1 >= 0 && blockIndex2 >= 0 {
			// Both blocks are normal data blocks, recover them using both P and Q parities
			dataBlock1, dataBlock2 := r.Math.RecoverTwoDataBlocks(dataBlocks, P, Q, blockIndex1, blockIndex2)
			err = r.Nodes[nodeID1].WriteBlockToDisk(InitBlock(blockIndex1, fileName, &dataBlock1, len(dataBlock1)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex1, err.Error())
			}
			err = r.Nodes[nodeID2].WriteBlockToDisk(InitBlock(blockIndex2, fileName, &dataBlock2, len(dataBlock2)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex2, err.Error())
			}

		} else if blockIndex1 >= 0 && blockIndex2 == -1 {
			// Recover normal data block and recalculate P parity
			dataBlock := r.Math.RecoverSingleBlockQ(dataBlocks, Q, blockIndex1) // Recover normal data block using Q
			pBlock := r.Math.RecoverPParity(dataBlocks)                         // Recalculate P parity
			err = r.Nodes[nodeID1].WriteBlockToDisk(InitBlock(blockIndex1, fileName, &dataBlock, len(dataBlock)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex1, err.Error())
			}
			err = r.Nodes[nodeID2].WriteBlockToDisk(InitBlock(-1, fileName, &pBlock, len(pBlock)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex2, err.Error())
			}
		} else if blockIndex1 >= 0 && blockIndex2 == -2 {
			// Recover normal data block and recalculate Q parity
			dataBlock := r.Math.RecoverSingleBlockP(dataBlocks, P, blockIndex1) // Recover normal data block using P
			qBlock := r.Math.RecoverQParity(dataBlocks)                         // Recalculate Q parity
			err = r.Nodes[nodeID1].WriteBlockToDisk(InitBlock(blockIndex1, fileName, &dataBlock, len(dataBlock)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex1, err.Error())
			}
			err = r.Nodes[nodeID2].WriteBlockToDisk(InitBlock(-2, fileName, &qBlock, len(qBlock)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex2, err.Error())
			}

		} else if blockIndex1 == -1 && blockIndex2 == -2 {
			// Both P and Q parities are missing, recalculate both
			P, Q = r.Math.RecoverPQParities(dataBlocks)
			err = r.Nodes[nodeID1].WriteBlockToDisk(InitBlock(-1, fileName, &P, len(P)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex1, err.Error())
			}
			err = r.Nodes[nodeID2].WriteBlockToDisk(InitBlock(-2, fileName, &Q, len(Q)))
			if err != nil {
				return fmt.Errorf("recovery of block %d failed: %s", blockIndex2, err.Error())
			}
		}
	}

	r.Nodes[nodeID1].status = true
	r.Nodes[nodeID2].status = true
	return nil
}
